package webshare

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"easysend/internal/diag"
)

// Start listens on cfg.Port, or any free port when it's taken.
func Start(cfg Config) (*Server, error) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Port))
	if err != nil {
		diag.Logf("browser: TCP port %d unavailable (%v); using another", cfg.Port, err)
		ln, err = net.Listen("tcp", ":0")
		if err != nil {
			diag.Logf("browser: no TCP port: %v", err)
			return nil, err
		}
	}
	s := &Server{cfg: cfg, shares: map[string]*Share{}, fails: map[string]int{}, batch: map[string]*uploadBatch{}}
	s.port = ln.Addr().(*net.TCPAddr).Port
	diag.Logf("browser: listening on TCP %d", s.port)
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
