package localsend

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"time"
)

// PeerFingerprint is the fingerprint of the certificate the client of r
// presented, or "" when it presented none (LocalSend clients don't).
// With pinned clients on both sides it identifies the device asking.
func PeerFingerprint(r *http.Request) string {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return ""
	}
	return fingerprint(r.TLS.PeerCertificates[0].Raw)
}

// ErrWrongPeer is a server whose certificate isn't the expected device's.
var ErrWrongPeer = errors.New("localsend: the device's certificate changed")

// PinnedClient talks to the device with fingerprint fp only: the server
// must present that certificate, and this device presents its own, so
// the server knows who asks (see PeerFingerprint). Callers bound each
// request with its context.
func (n *Node) PinnedClient(fp string) *http.Client {
	conf := &tls.Config{
		// Self-signed certificates: checked against the fingerprint below
		// instead of a certificate authority.
		InsecureSkipVerify: true, //nolint:gosec
		MinVersion:         tls.VersionTLS12,
		VerifyPeerCertificate: func(raw [][]byte, _ [][]*x509.Certificate) error {
			if len(raw) == 0 || fingerprint(raw[0]) != fp {
				return ErrWrongPeer
			}
			return nil
		},
	}
	if n.cfg.Cert != nil {
		conf.Certificates = []tls.Certificate{*n.cfg.Cert}
	}
	return &http.Client{Transport: &http.Transport{
		TLSClientConfig:     conf,
		DialContext:         (&net.Dialer{Timeout: 3 * time.Second}).DialContext,
		TLSHandshakeTimeout: 4 * time.Second,
		MaxIdleConnsPerHost: 2,
	}}
}

// BaseURL is the device's server address, "https://ip:port".
func (d Device) BaseURL() string { return d.scheme() + "://" + d.Addr() }
