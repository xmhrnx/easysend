// Package engine runs EasySend's services - nearby transfers, browser
// links and internet transfers - and exposes their state to the UI. It
// has no UI code; the UI reads snapshots and calls actions.
package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"easysend/internal/localsend"
	"easysend/internal/natmap"
	"easysend/internal/netutil"
	"easysend/internal/settings"
	"easysend/internal/transfer"
	"easysend/internal/webshare"
)

// Config sets up the engine.
type Config struct {
	// DataDir holds settings, history and the TLS certificate.
	DataDir string
	// DownloadDir is the default folder for received files.
	DownloadDir string
	// DeviceModel ("MacBook Pro", "Pixel 8") and DeviceType
	// (localsend.TypeDesktop or TypeMobile) describe this device.
	DeviceModel string
	DeviceType  string
	// OnChange runs on any goroutine whenever state the UI shows changed.
	OnChange func()
	// Notify shows a system notification.
	Notify func(title, body string)
}

// Engine is the running app core.
type Engine struct {
	cfg       Config
	Settings  *settings.Store
	Transfers *transfer.Registry

	node *localsend.Node
	web  *webshare.Server

	mu       sync.Mutex
	requests []*Request
	public   PublicStatus
	mapping  *natmap.Mapping
	scanning bool
	stopNAT  context.CancelFunc
}

// Start loads the settings and starts every service.
func Start(cfg Config) (*Engine, error) {
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return nil, err
	}
	st, err := settings.Open(cfg.DataDir, cfg.DownloadDir)
	if err != nil {
		return nil, fmt.Errorf("settings: %w", err)
	}
	e := &Engine{cfg: cfg, Settings: st}
	e.Transfers = transfer.NewRegistry(cfg.DataDir, e.changed)

	cert, fp, err := localsend.LoadOrCreateCert(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("certificate: %w", err)
	}
	s := st.Get()
	e.node, err = localsend.Start(localsend.Config{
		Self:         e.selfInfo,
		Cert:         &cert,
		Fingerprint:  fp,
		Port:         s.Port,
		PIN:          e.receivePIN,
		Discoverable: func() bool { return e.Settings.Get().Discoverable },
		Accept:       e.accept,
		OnText:       e.onNearbyText,
		OnDevices:    e.changed,
	})
	if err != nil {
		return nil, fmt.Errorf("nearby sharing: %w", err)
	}
	e.web, err = webshare.Start(webshare.Config{
		Port:        s.WebPort,
		Device:      func() string { return e.Settings.Get().DeviceName },
		Transfers:   e.Transfers,
		SaveDir:     func() string { return e.Settings.Get().DownloadDir },
		DropEnabled: func() bool { return e.Settings.Get().BrowserDrop },
		DropPIN:     func() string { return e.Settings.Get().DropPIN },
		DropToken:   func() string { return e.Settings.Get().DropToken },
		OnText:      e.onBrowserText,
		OnReceived:  e.onReceived,
	})
	if err != nil {
		e.node.Close()
		return nil, fmt.Errorf("browser sharing: %w", err)
	}
	e.applyPublic()
	go func() {
		time.Sleep(1500 * time.Millisecond)
		if len(e.node.Devices()) == 0 {
			e.Refresh()
		}
	}()
	return e, nil
}

// Close stops every service.
func (e *Engine) Close() {
	e.mu.Lock()
	m, stop := e.mapping, e.stopNAT
	e.mu.Unlock()
	if stop != nil {
		stop()
	}
	if m != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_ = m.Remove(ctx)
		cancel()
	}
	e.node.Close()
	e.web.Close()
}

func (e *Engine) changed() {
	if e.cfg.OnChange != nil {
		e.cfg.OnChange()
	}
}

func (e *Engine) notify(title, body string) {
	if e.cfg.Notify != nil && e.Settings.Get().Notifications {
		e.cfg.Notify(title, body)
	}
}

func (e *Engine) selfInfo() localsend.Info {
	model := e.cfg.DeviceModel
	return localsend.Info{Alias: e.Settings.Get().DeviceName, DeviceModel: &model, DeviceType: e.cfg.DeviceType}
}

func (e *Engine) receivePIN() string {
	s := e.Settings.Get()
	if s.RequirePIN {
		return s.PIN
	}
	return ""
}

// Self describes this device for the UI.
type Self struct {
	Name, IP, Model string
	Port, WebPort   int
	Fingerprint     string
}

// Self returns this device's name and addresses.
func (e *Engine) Self() Self {
	info := e.node.Self()
	return Self{Name: info.Alias, IP: netutil.PrimaryIPv4(), Model: e.cfg.DeviceModel, Port: e.node.Port(), WebPort: e.web.Port(), Fingerprint: info.Fingerprint}
}

// UpdateSettings changes the settings and applies what changed.
func (e *Engine) UpdateSettings(fn func(*settings.Settings)) error {
	before := e.Settings.Get()
	err := e.Settings.Update(fn)
	after := e.Settings.Get()
	if before.UPnP != after.UPnP {
		e.applyPublic()
	}
	if before.DeviceName != after.DeviceName || (!before.Discoverable && after.Discoverable) {
		go e.node.Refresh(context.Background())
	}
	e.changed()
	return err
}

// LANBase is the browser address of this device on the local network.
func (e *Engine) LANBase() string {
	ip := netutil.PrimaryIPv4()
	if ip == "" {
		ip = "localhost"
	}
	return "http://" + ip + ":" + strconv.Itoa(e.web.Port())
}

// OtherLANBases are browser addresses on this device's other local
// addresses (a second adapter), for when the first doesn't open.
func (e *Engine) OtherLANBases() []string {
	ips := netutil.LocalIPv4()
	var out []string
	for i, ip := range ips {
		if i == 0 {
			continue
		}
		out = append(out, "http://"+ip.String()+":"+strconv.Itoa(e.web.Port()))
		if len(out) == 2 {
			break
		}
	}
	return out
}

// PublicBase is the internet address of the browser pages, or "".
func (e *Engine) PublicBase() string {
	if u := e.Settings.Get().PublicURL; u != "" {
		return trimSlash(u)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.public.State == PublicReady {
		return "http://" + e.public.IP + ":" + strconv.Itoa(e.public.Port)
	}
	return ""
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

// errorText turns errors into short messages for people.
func errorText(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, localsend.ErrPINRequired):
		return "PIN required"
	case errors.Is(err, localsend.ErrDeclined):
		return "Declined by the receiver"
	case errors.Is(err, localsend.ErrBusy):
		return "The receiver is busy"
	case errors.Is(err, context.DeadlineExceeded):
		return "Timed out"
	}
	return err.Error()
}
