package webshare

import (
	"net/http"
	"sync"

	"easysend/internal/transfer"
)

type uploadBatch struct {
	id      string
	files   []transfer.File
	pending map[int]bool
	renames map[string]string
	dir     string
	mu      sync.Mutex
}

type beginRequest struct {
	Files []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"files"`
}

func (s *Server) dropAllowed(w http.ResponseWriter, r *http.Request) bool {
	token := r.PathValue("token")
	if !s.cfg.DropEnabled() || token != s.cfg.DropToken() {
		s.gone(w)
		return false
	}
	return true
}

func (s *Server) dropUnlocked(w http.ResponseWriter, r *http.Request) bool {
	if !unlocked(r, r.PathValue("token"), s.cfg.DropPIN()) {
		http.Error(w, "PIN required", http.StatusUnauthorized)
		return false
	}
	return true
}
