package webshare

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"easysend/internal/item"
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

func (s *Server) dropAllowed(w http.ResponseWriter, r *http.Request) bool {
	token := r.PathValue("token")
	if !s.cfg.DropEnabled() || token != s.cfg.DropToken() {
		s.gone(w)
		return false
	}
	return true
}

func (s *Server) handleDrop(w http.ResponseWriter, r *http.Request) {
	if !s.dropAllowed(w, r) {
		return
	}
	token := r.PathValue("token")
	if !unlocked(r, token, s.cfg.DropPIN()) {
		s.renderPIN(w, "/r/"+token, "")
		return
	}
	s.render(w, "drop", pageData{Device: s.cfg.Device(), Token: token})
}

type beginRequest struct {
	Files []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"files"`
}

func (s *Server) handleBegin(w http.ResponseWriter, r *http.Request) {
	if !s.dropAllowed(w, r) || !s.dropUnlocked(w, r) {
		return
	}
	var req beginRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil || len(req.Files) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	b := &uploadBatch{pending: map[int]bool{}, renames: map[string]string{}, dir: s.cfg.SaveDir()}
	for i, f := range req.Files {
		b.files = append(b.files, transfer.File{Name: f.Name, Size: f.Size})
		b.pending[i] = true
	}
	b.id = s.cfg.Transfers.Start(transfer.Transfer{Direction: transfer.Incoming, Method: transfer.Browser, Peer: clientName(r), Files: b.files, Status: transfer.Active, Folder: b.dir}, nil)
	s.mu.Lock()
	s.batch[b.id] = b
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"batch": b.id})
}

func (s *Server) dropUnlocked(w http.ResponseWriter, r *http.Request) bool {
	if !unlocked(r, r.PathValue("token"), s.cfg.DropPIN()) {
		http.Error(w, "PIN required", http.StatusUnauthorized)
		return false
	}
	return true
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if !s.dropAllowed(w, r) || !s.dropUnlocked(w, r) {
		return
	}
	s.mu.Lock()
	b, ok := s.batch[r.URL.Query().Get("batch")]
	s.mu.Unlock()
	i, err := strconv.Atoi(r.URL.Query().Get("i"))
	if !ok || err != nil {
		http.Error(w, "unknown upload", http.StatusBadRequest)
		return
	}
	b.mu.Lock()
	if !b.pending[i] {
		b.mu.Unlock()
		http.Error(w, "unknown file", http.StatusBadRequest)
		return
	}
	dst := item.UniqueTopLevel(b.dir, b.files[i].Name, b.renames)
	b.mu.Unlock()

	err = saveUpload(dst, r.Body, func(n int64) { s.cfg.Transfers.AddBytes(b.id, i, n) })
	if err != nil {
		s.cfg.Transfers.Finish(b.id, transfer.Failed, err)
		s.dropBatch(b.id)
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}
	s.cfg.Transfers.Update(b.id, func(t *transfer.Transfer) { t.Files[i].Saved = dst })
	b.mu.Lock()
	delete(b.pending, i)
	done := len(b.pending) == 0
	b.mu.Unlock()
	if done {
		s.cfg.Transfers.Finish(b.id, transfer.Done, nil)
		s.dropBatch(b.id)
		if t, ok := s.cfg.Transfers.Get(b.id); ok && s.cfg.OnReceived != nil {
			s.cfg.OnReceived(t)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) dropBatch(id string) {
	s.mu.Lock()
	delete(s.batch, id)
	s.mu.Unlock()
}

func saveUpload(dst string, body io.Reader, count func(int64)) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(&countingWriter{ResponseWriter: nopRW{f}, fn: count}, body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(dst)
	}
	return err
}

func (s *Server) handleText(w http.ResponseWriter, r *http.Request) {
	if !s.dropAllowed(w, r) || !s.dropUnlocked(w, r) {
		return
	}
	b, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil || len(b) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if s.cfg.OnText != nil {
		s.cfg.OnText(clientName(r), string(b))
	}
	w.WriteHeader(http.StatusOK)
}
