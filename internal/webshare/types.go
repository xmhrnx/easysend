// Package webshare serves EasySend's browser pages: people without the
// app open a link (or scan its QR code) to download what was shared, or
// to upload files and text to this device.
package webshare

import (
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"easysend/internal/item"
	"easysend/internal/transfer"
)

// Share is a set of items offered through a link.
type Share struct {
	Token        string
	Items        []item.Item
	Created      time.Time
	Expires      time.Time // zero: never
	MaxDownloads int       // 0: unlimited
	Downloads    int
	PIN          string
	Public       bool
}

// Config sets up the server.
type Config struct {
	Port      int
	Device    func() string // this device's name, shown on the pages
	Transfers *transfer.Registry
	// SaveDir returns where uploads go.
	SaveDir func() string
	// DropEnabled reports whether browser uploads are on, and DropPIN
	// the PIN they need ("" for none).
	DropEnabled func() bool
	DropPIN     func() string
	// DropToken returns the secret in the upload link.
	DropToken func() string
	// OnText receives text sent from a browser.
	OnText func(from, text string)
	// OnReceived runs when a browser upload finished.
	OnReceived func(t transfer.Transfer)
}

// Server is the web server.
type Server struct {
	cfg  Config
	srv  *http.Server
	lnMu sync.Mutex // guards ln, which Resume replaces
	ln   net.Listener
	port atomic.Int32 // Resume may change it

	mu     sync.Mutex
	shares map[string]*Share
	fails  map[string]int
	batch  map[string]*uploadBatch
}

// Expired reports whether the link stopped working.
func (s Share) Expired() bool {
	return (!s.Expires.IsZero() && time.Now().After(s.Expires)) || (s.MaxDownloads > 0 && s.Downloads >= s.MaxDownloads)
}

// Size is the total size of the items.
func (s Share) Size() int64 { return item.TotalSize(s.Items) }
