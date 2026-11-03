// Package webshare serves EasySend's browser pages: people without the
// app open a link (or scan its QR code) to download what was shared, or
// to upload files and text to this device.
package webshare

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"sort"
	"strconv"
	"sync"
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

// Expired reports whether the link stopped working.
func (s Share) Expired() bool {
	return (!s.Expires.IsZero() && time.Now().After(s.Expires)) || (s.MaxDownloads > 0 && s.Downloads >= s.MaxDownloads)
}

// Size is the total size of the items.
func (s Share) Size() int64 { return item.TotalSize(s.Items) }

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
	port int

	mu     sync.Mutex
	shares map[string]*Share
	fails  map[string]int
	batch  map[string]*uploadBatch
}

// Start listens on cfg.Port, or any free port when it's taken.
func Start(cfg Config) (*Server, error) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Port))
	if err != nil {
		ln, err = net.Listen("tcp", ":0")
		if err != nil {
			return nil, err
		}
	}
	s := &Server{cfg: cfg, shares: map[string]*Share{}, fails: map[string]int{}, batch: map[string]*uploadBatch{}}
	s.port = ln.Addr().(*net.TCPAddr).Port
	s.srv = &http.Server{Handler: s.routes(), ReadHeaderTimeout: 20 * time.Second}
	go func() { _ = s.srv.Serve(ln) }()
	return s, nil
}

// Port is the listening port.
func (s *Server) Port() int { return s.port }

// Close stops the server.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

// NewToken returns a random URL-safe secret.
func NewToken() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return base64.RawURLEncoding.EncodeToString(b[:])
}

// Add publishes a share and returns it.
func (s *Server) Add(sh Share) Share {
	sh.Token = NewToken()
	sh.Created = time.Now()
	s.mu.Lock()
	s.shares[sh.Token] = &sh
	s.mu.Unlock()
	return sh
}

// Remove stops a share.
func (s *Server) Remove(token string) {
	s.mu.Lock()
	delete(s.shares, token)
	s.mu.Unlock()
}

// Shares lists the shares, newest first.
func (s *Server) Shares() []Share {
	s.mu.Lock()
	out := make([]Share, 0, len(s.shares))
	for _, sh := range s.shares {
		out = append(out, *sh)
	}
	s.mu.Unlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Created.After(out[j].Created) })
	return out
}

func (s *Server) share(token string) (*Share, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sh, ok := s.shares[token]
	return sh, ok
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleHome)
	mux.HandleFunc("GET /s/{token}", s.handleShare)
	mux.HandleFunc("POST /s/{token}/pin", s.handlePIN)
	mux.HandleFunc("GET /s/{token}/f/{idx}", s.handleFile)
	mux.HandleFunc("GET /s/{token}/zip", s.handleZip)
	mux.HandleFunc("GET /r/{token}", s.handleDrop)
	mux.HandleFunc("POST /r/{token}/pin", s.handlePIN)
	mux.HandleFunc("POST /r/{token}/begin", s.handleBegin)
	mux.HandleFunc("POST /r/{token}/file", s.handleUpload)
	mux.HandleFunc("POST /r/{token}/text", s.handleText)
	mux.HandleFunc("GET /logo.svg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "max-age=86400")
		_, _ = w.Write([]byte(logoSVG))
	})
	return securityHeaders(mux)
}

func securityHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		h.ServeHTTP(w, r)
	})
}

// --- PIN handling ---

func cookieName(token string) string { return "es_" + token[:8] }

func pinCookie(token, pin string) string {
	sum := sha256.Sum256([]byte(token + "\x00" + pin))
	return hex.EncodeToString(sum[:])
}

// unlocked reports whether the request may use token (PIN given or none
// needed).
func unlocked(r *http.Request, token, pin string) bool {
	if pin == "" {
		return true
	}
	c, err := r.Cookie(cookieName(token))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(c.Value), []byte(pinCookie(token, pin))) == 1
}

var errLocked = errors.New("too many wrong PINs")

func (s *Server) handlePIN(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	var want, back string
	if sh, ok := s.share(token); ok {
		want, back = sh.PIN, "/s/"+token
	} else if token == s.cfg.DropToken() {
		want, back = s.cfg.DropPIN(), "/r/"+token
	} else {
		s.gone(w)
		return
	}
	s.mu.Lock()
	locked := s.fails[token] >= 10
	s.mu.Unlock()
	if locked {
		s.renderPIN(w, back, errLocked.Error())
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.FormValue("pin")), []byte(want)) != 1 {
		s.mu.Lock()
		s.fails[token]++
		s.mu.Unlock()
		s.renderPIN(w, back, "That PIN isn't right. Try again.")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName(token), Value: pinCookie(token, want), Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
	http.Redirect(w, r, back, http.StatusSeeOther)
}

func clientName(r *http.Request) string {
	ua := r.UserAgent()
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return browserName(ua) + " · " + host
}
