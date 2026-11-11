// Package engine runs EasySend's services - nearby transfers, browser
// links and internet transfers - and exposes their state to the UI. It
// has no UI code; the UI reads snapshots and calls actions.
package engine

import (
	"context"
	"fmt"
	"os"
	"time"

	"easysend/internal/diag"
	"easysend/internal/localsend"
	"easysend/internal/netutil"
	"easysend/internal/settings"
	"easysend/internal/transfer"
	"easysend/internal/webshare"
)

// Start loads the settings and starts every service.
func Start(cfg Config) (*Engine, error) {
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return nil, err
	}
	st, err := settings.Open(cfg.DataDir, cfg.DownloadDir)
	if err != nil {
		return nil, fmt.Errorf("settings: %w", err)
	}
	if cfg.FixedDownloadDir && st.Get().DownloadDir != cfg.DownloadDir {
		if err := st.Update(func(s *settings.Settings) { s.DownloadDir = cfg.DownloadDir }); err != nil {
			return nil, fmt.Errorf("settings: %w", err)
		}
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
	diag.Logf("adapters:\n%s", netutil.Describe())
	diag.Logf("links use %s; nearby port %d, browser port %d", e.LANBase(), e.node.Port(), e.web.Port())
	go func() {
		time.Sleep(1500 * time.Millisecond)
		if len(e.node.Devices()) == 0 {
			e.Refresh()
		}
	}()
	return e, nil
}

// Resume reopens the listening sockets, which iOS takes back while the
// app is suspended: call it when the app returns to the foreground.
func (e *Engine) Resume() {
	e.node.Resume()
	e.web.Resume()
	e.changed()
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
