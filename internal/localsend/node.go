package localsend

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

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
