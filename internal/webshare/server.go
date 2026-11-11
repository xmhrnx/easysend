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
	s := &Server{cfg: cfg, shares: map[string]*Share{}, fails: map[string]int{}, batch: map[string]*uploadBatch{}}
	s.srv = &http.Server{Handler: s.routes(), ReadHeaderTimeout: 20 * time.Second}
	if err := s.listen(cfg.Port); err != nil {
		return nil, err
	}
	return s, nil
}

// listen opens the listener on port (any free port when it's taken) and
// serves on it.
func (s *Server) listen(port int) error {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		diag.Logf("browser: TCP port %d unavailable (%v); using another", port, err)
		ln, err = net.Listen("tcp", ":0")
		if err != nil {
			diag.Logf("browser: no TCP port: %v", err)
			return err
		}
	}
	s.port.Store(int32(ln.Addr().(*net.TCPAddr).Port))
	s.ln = ln
	diag.Logf("browser: listening on TCP %d", s.Port())
	go func() { _ = s.srv.Serve(ln) }()
	return nil
}

// Resume replaces the listener, which iOS takes back from a suspended
// app (see localsend.Node.Resume). Open links keep their addresses as
// long as the port is free again, which it normally is.
func (s *Server) Resume() {
	s.lnMu.Lock()
	defer s.lnMu.Unlock()
	if s.ln != nil {
		_ = s.ln.Close()
	}
	_ = s.listen(s.Port())
}

// Port is the listening port.
func (s *Server) Port() int { return int(s.port.Load()) }

// Close stops the server.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}
