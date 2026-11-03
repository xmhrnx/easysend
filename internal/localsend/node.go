package localsend

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Incoming is a transfer offered by a peer.
type Incoming struct {
	Sender Device
	Files  []FileMeta
	// Cancel stops the transfer from this side.
	Cancel func()
}

// Sink receives the files of an accepted transfer.
type Sink interface {
	// Write stores the content of file id, read from r.
	Write(ctx context.Context, id string, r io.Reader) error
	// Close ends the transfer: err is nil when every file arrived.
	Close(err error)
}

// Config sets up a Node.
type Config struct {
	// Self returns this device's current alias, model and type; the node
	// fills in the fingerprint, port and protocol.
	Self func() Info
	// Cert is the TLS certificate; nil serves plain HTTP.
	Cert *tls.Certificate
	// Fingerprint identifies this device.
	Fingerprint string
	// Port to listen on (0: the LocalSend default, -1: any); another is
	// chosen when it's taken.
	Port int
	// PIN returns the PIN senders must give, or "" for none.
	PIN func() string
	// Discoverable reports whether to answer discovery.
	Discoverable func() bool
	// Accept decides on an offered transfer, blocking until the user
	// answers. It returns a nil Sink to decline.
	Accept func(ctx context.Context, in Incoming) Sink
	// OnText receives a text message (a text-only transfer).
	OnText func(from Device, text string)
	// OnDevices runs after the device list changes.
	OnDevices func()
}

// Node is a LocalSend peer: a server for incoming transfers, discovery and
// a client for sending.
type Node struct {
	cfg      Config
	ln       net.Listener
	srv      *http.Server
	client   *http.Client
	port     int
	protocol string

	mu      sync.Mutex
	devices map[string]*Device
	session *session

	disc *discovery
}

type session struct {
	id      string
	sender  string
	tokens  map[string]string
	pending map[string]bool
	sink    Sink
	ctx     context.Context
	cancel  context.CancelFunc
	last    time.Time
}

// Start listens and begins discovery.
func Start(cfg Config) (*Node, error) {
	n := &Node{cfg: cfg, devices: map[string]*Device{}, protocol: "http"}
	if cfg.Cert != nil {
		n.protocol = "https"
	}
	port := cfg.Port
	switch {
	case port == 0:
		port = DefaultPort
	case port < 0:
		port = 0 // any free port (tests)
	}
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		ln, err = net.Listen("tcp", ":0")
		if err != nil {
			return nil, err
		}
	}
	n.port = ln.Addr().(*net.TCPAddr).Port
	if cfg.Cert != nil {
		ln = tls.NewListener(ln, &tls.Config{Certificates: []tls.Certificate{*cfg.Cert}, MinVersion: tls.VersionTLS12})
	}
	n.ln = ln
	n.client = newClient()
	n.srv = &http.Server{Handler: n.routes(), ReadHeaderTimeout: 15 * time.Second}
	go func() { _ = n.srv.Serve(ln) }()
	n.disc = startDiscovery(n)
	return n, nil
}

func newClient() *http.Client {
	tr := &http.Transport{
		// Peers use self-signed certificates, as in LocalSend.
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		DialContext:         (&net.Dialer{Timeout: 3 * time.Second}).DialContext,
		TLSHandshakeTimeout: 4 * time.Second,
		MaxIdleConnsPerHost: 4,
	}
	return &http.Client{Transport: tr}
}

// Close stops the node.
func (n *Node) Close() error {
	n.disc.close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return n.srv.Shutdown(ctx)
}

// Port is the port the node listens on.
func (n *Node) Port() int { return n.port }

// Protocol is "https" or "http".
func (n *Node) Protocol() string { return n.protocol }

// Self returns this device's full info.
func (n *Node) Self() Info {
	info := n.cfg.Self()
	info.Version = ProtocolVer
	info.Fingerprint = n.cfg.Fingerprint
	info.Port = n.port
	info.Protocol = n.protocol
	return info
}

// Devices returns the known peers, sorted by name.
func (n *Node) Devices() []Device {
	n.mu.Lock()
	out := make([]Device, 0, len(n.devices))
	for _, d := range n.devices {
		out = append(out, *d)
	}
	n.mu.Unlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Alias < out[j].Alias })
	return out
}

// Device returns the peer with a fingerprint.
func (n *Node) Device(fp string) (Device, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	d, ok := n.devices[fp]
	if !ok {
		return Device{}, false
	}
	return *d, true
}

// ForgetDevices clears the list, before a fresh discovery.
func (n *Node) ForgetDevices() {
	n.mu.Lock()
	for fp, d := range n.devices {
		if !d.Manual {
			delete(n.devices, fp)
		}
	}
	n.mu.Unlock()
	n.devicesChanged()
}

func (n *Node) addDevice(d Device) {
	if d.Fingerprint == "" || d.Fingerprint == n.cfg.Fingerprint {
		return
	}
	d.LastSeen = time.Now()
	n.mu.Lock()
	if old, ok := n.devices[d.Fingerprint]; ok && old.Manual {
		d.Manual = true
	}
	n.devices[d.Fingerprint] = &d
	n.mu.Unlock()
	n.devicesChanged()
}

func (n *Node) devicesChanged() {
	if n.cfg.OnDevices != nil {
		n.cfg.OnDevices()
	}
}

// Register asks the device at addr ("ip" or "ip:port") to identify itself
// and adds it, trying HTTPS, then HTTP.
func (n *Node) Register(ctx context.Context, addr string, manual bool) (Device, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		host, portStr = addr, strconv.Itoa(DefaultPort)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Device{}, fmt.Errorf("bad port %q", portStr)
	}
	var lastErr error
	for _, proto := range []string{"https", "http"} {
		d := Device{IP: host, Info: Info{Port: port, Protocol: proto}}
		info, err := n.register(ctx, d)
		if err != nil {
			lastErr = err
			continue
		}
		d.Info = info
		d.Port, d.Protocol, d.Manual = port, proto, manual
		n.addDevice(d)
		return d, nil
	}
	return Device{}, lastErr
}

func (n *Node) register(ctx context.Context, d Device) (Info, error) {
	var info Info
	err := n.postJSON(ctx, d.url("/register"), n.Self(), &info)
	if err != nil {
		return info, err
	}
	if info.Fingerprint == "" {
		return info, errors.New("not a LocalSend device")
	}
	return info, nil
}
