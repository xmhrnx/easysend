package foldersync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// apiBase is EasySend's own API, beside LocalSend's on the same server.
const apiBase = "/api/easysend/v1/sync"

type inviteRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	From string `json:"from"`
}

type indexResponse struct {
	Name  string `json:"name"`
	Index Index  `json:"index"`
}

// Routes adds the sync endpoints to a server whose requests carry the
// client's certificate (see Config.PeerFP).
func (m *Manager) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST "+apiBase+"/invite", m.handleInvite)
	mux.HandleFunc("GET "+apiBase+"/{id}/index", m.handleIndex)
	mux.HandleFunc("GET "+apiBase+"/{id}/file", m.handleFile)
	mux.HandleFunc("POST "+apiBase+"/{id}/changed", m.handleChanged)
	mux.HandleFunc("POST "+apiBase+"/{id}/leave", m.handleLeave)
}

// member returns the folder of r's {id} when the device asking shares
// it, answering the request itself otherwise.
func (m *Manager) member(w http.ResponseWriter, r *http.Request) (*folder, string, bool) {
	fp := m.cfg.PeerFP(r)
	if fp == "" {
		http.Error(w, "a device certificate is required", http.StatusUnauthorized)
		return nil, "", false
	}
	m.mu.Lock()
	f, ok := m.folders[r.PathValue("id")]
	member := ok && f.hasPeer(fp)
	m.mu.Unlock()
	switch {
	case !ok:
		http.Error(w, "no such folder", http.StatusNotFound)
		return nil, "", false
	case !member:
		http.Error(w, "this folder isn't shared with your device", http.StatusForbidden)
		return nil, "", false
	}
	return f, fp, true
}

func (m *Manager) handleInvite(w http.ResponseWriter, r *http.Request) {
	fp := m.cfg.PeerFP(r)
	if fp == "" {
		http.Error(w, "a device certificate is required", http.StatusUnauthorized)
		return
	}
	var req inviteRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&req); err != nil || !validID(req.ID) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		req.Name = "Shared folder"
	}
	// Already syncing this folder (with another device): just add this
	// one, if the user agrees.
	root, ok := m.cfg.OnInvite(r.Context(), Invite{FP: fp, From: req.From, ID: req.ID, Name: req.Name})
	if !ok {
		http.Error(w, "declined", http.StatusForbidden)
		return
	}
	peer := Peer{FP: fp, Name: req.From}
	m.mu.Lock()
	f, exists := m.folders[req.ID]
	if !exists {
		if err := os.MkdirAll(root, 0o755); err != nil {
			m.mu.Unlock()
			http.Error(w, "can't create the folder", http.StatusInternalServerError)
			return
		}
		f = &folder{
			Folder: Folder{ID: req.ID, Name: req.Name, Root: root, Created: time.Now()},
			st:     state{Local: Index{}, Base: map[string]map[string]string{}},
			status: Status{Peers: map[string]PeerStatus{}},
		}
		m.folders[req.ID] = f
	}
	f.addPeer(peer)
	err := m.saveFolders()
	m.mu.Unlock()
	if err != nil {
		http.Error(w, "can't save the folder", http.StatusInternalServerError)
		return
	}
	m.changed()
	m.SyncNow()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"name": m.name()})
}

func (m *Manager) handleIndex(w http.ResponseWriter, r *http.Request) {
	f, _, ok := m.member(w, r)
	if !ok {
		return
	}
	m.mu.Lock()
	resp := indexResponse{Name: f.Name, Index: make(Index, len(f.st.Local))}
	for p, e := range f.st.Local {
		resp.Index[p] = e
	}
	m.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (m *Manager) handleFile(w http.ResponseWriter, r *http.Request) {
	f, _, ok := m.member(w, r)
	if !ok {
		return
	}
	rel := cleanPath(r.URL.Query().Get("path"))
	if rel == "" {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	file, err := os.Open(filepath.Join(f.Root, filepath.FromSlash(rel)))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprint(info.Size()))
	_, _ = io.Copy(w, file)
}

func (m *Manager) handleChanged(w http.ResponseWriter, r *http.Request) {
	f, fp, ok := m.member(w, r)
	if !ok {
		return
	}
	m.mu.Lock()
	if f.asked == nil {
		f.asked = map[string]bool{}
	}
	f.asked[fp] = true
	m.mu.Unlock()
	m.SyncNow()
	w.WriteHeader(http.StatusNoContent)
}

func (m *Manager) handleLeave(w http.ResponseWriter, r *http.Request) {
	f, fp, ok := m.member(w, r)
	if !ok {
		return
	}
	m.mu.Lock()
	f.removePeer(fp)
	delete(f.st.Base, fp)
	_ = m.saveFolders()
	_ = m.saveState(f)
	m.mu.Unlock()
	m.changed()
	w.WriteHeader(http.StatusNoContent)
}

func validID(id string) bool {
	if len(id) != 32 {
		return false
	}
	for _, c := range id {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}

// --- client --------------------------------------------------------------

func (m *Manager) do(ctx context.Context, fp, base, method, path string, body any) (*http.Response, error) {
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, base+apiBase+path, rd)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := m.cfg.Client(fp).Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, &statusErr{code: resp.StatusCode, msg: strings.TrimSpace(string(msg))}
	}
	return resp, nil
}

type statusErr struct {
	code int
	msg  string
}

func (e *statusErr) Error() string {
	if e.msg != "" {
		return e.msg
	}
	return http.StatusText(e.code)
}

func (m *Manager) sendInvite(ctx context.Context, fp, base string, req inviteRequest) error {
	resp, err := m.do(ctx, fp, base, http.MethodPost, "/invite", req)
	if err != nil {
		if se, ok := err.(*statusErr); ok {
			switch se.code {
			case http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusUnauthorized:
				return ErrNotEasySend
			case http.StatusForbidden:
				return ErrDeclined
			}
		}
		return err
	}
	resp.Body.Close()
	return nil
}

// ErrDeclined is an invitation the other device's user turned down.
var ErrDeclined = fmt.Errorf("declined")

func (m *Manager) fetchIndex(ctx context.Context, fp, base, id string) (Index, error) {
	resp, err := m.do(ctx, fp, base, http.MethodGet, "/"+id+"/index", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var ir indexResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 256<<20)).Decode(&ir); err != nil {
		return nil, err
	}
	// Only well-formed paths: a peer can't reach outside the folder.
	idx := make(Index, len(ir.Index))
	for p, e := range ir.Index {
		if c := cleanPath(p); c != "" && c == p {
			idx[p] = e
		}
	}
	return idx, nil
}

func (m *Manager) fetchFile(ctx context.Context, fp, base, id, rel string) (io.ReadCloser, error) {
	resp, err := m.do(ctx, fp, base, http.MethodGet, "/"+id+"/file?path="+url.QueryEscape(rel), nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (m *Manager) sendChanged(fp, base, id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if resp, err := m.do(ctx, fp, base, http.MethodPost, "/"+id+"/changed", nil); err == nil {
		resp.Body.Close()
	}
}

func (m *Manager) sendLeave(id string, p Peer) {
	base, ok := m.address(p)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if resp, err := m.do(ctx, p.FP, base, http.MethodPost, "/"+id+"/leave", nil); err == nil {
		resp.Body.Close()
	}
}

// address is where to reach a peer: where discovery found it, or where
// it was last.
func (m *Manager) address(p Peer) (string, bool) {
	if m.cfg.Resolve != nil {
		if base, ok := m.cfg.Resolve(p.FP); ok {
			return base, true
		}
	}
	return p.Addr, p.Addr != ""
}
