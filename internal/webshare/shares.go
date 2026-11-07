package webshare

import (
	"crypto/rand"
	"encoding/base64"
	"sort"
	"time"
)

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
