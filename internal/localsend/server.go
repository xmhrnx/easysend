package localsend

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"
)

func (n *Node) routes() http.Handler {
	mux := http.NewServeMux()
	for _, v := range []string{"/api/localsend/v1", apiBase} {
		mux.HandleFunc(v+"/info", n.handleInfo)
		mux.HandleFunc(v+"/register", n.handleRegister)
	}
	mux.HandleFunc("POST "+apiBase+"/prepare-upload", n.handlePrepare)
	mux.HandleFunc("POST "+apiBase+"/upload", n.handleUpload)
	mux.HandleFunc("POST "+apiBase+"/cancel", n.handleCancel)
	return mux
}

func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (n *Node) discoverable() bool {
	return n.cfg.Discoverable == nil || n.cfg.Discoverable()
}

func (n *Node) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("fingerprint") == n.cfg.Fingerprint {
		http.Error(w, "self-discovered", http.StatusPreconditionFailed)
		return
	}
	if !n.discoverable() {
		http.Error(w, "hidden", http.StatusForbidden)
		return
	}
	writeJSON(w, n.Self())
}

func (n *Node) handleRegister(w http.ResponseWriter, r *http.Request) {
	var info Info
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&info); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if info.Fingerprint == n.cfg.Fingerprint {
		http.Error(w, "self-discovered", http.StatusPreconditionFailed)
		return
	}
	n.addDevice(Device{Info: info, IP: remoteIP(r)})
	if !n.discoverable() {
		http.Error(w, "hidden", http.StatusForbidden)
		return
	}
	writeJSON(w, n.Self())
}

func newToken() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

const sessionIdle = 2 * time.Minute

func (n *Node) handlePrepare(w http.ResponseWriter, r *http.Request) {
	if pin := n.pin(); pin != "" && r.URL.Query().Get("pin") != pin {
		http.Error(w, "PIN required", http.StatusUnauthorized)
		return
	}
	var req prepareRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&req); err != nil || len(req.Files) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	sender := Device{Info: req.Info, IP: remoteIP(r)}
	n.addDevice(sender)

	if text, ok := textMessage(req.Files); ok {
		if n.cfg.OnText != nil {
			n.cfg.OnText(sender, text)
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	n.mu.Lock()
	if s := n.session; s != nil {
		if time.Since(s.last) < sessionIdle && s.ctx.Err() == nil {
			n.mu.Unlock()
			http.Error(w, "blocked by another session", http.StatusConflict)
			return
		}
		s.cancel()
		n.session = nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &session{id: newToken(), sender: sender.Fingerprint, tokens: map[string]string{}, pending: map[string]bool{}, ctx: ctx, cancel: cancel, last: time.Now()}
	n.session = s
	n.mu.Unlock()

	files := make([]FileMeta, 0, len(req.Files))
	for id, f := range req.Files {
		f.ID = id
		files = append(files, f)
	}
	sortFiles(files)

	// Stop waiting if the sender gives up.
	waitCtx, waitCancel := context.WithCancel(ctx)
	go func() {
		select {
		case <-r.Context().Done():
		case <-waitCtx.Done():
		}
		waitCancel()
	}()
	var sink Sink
	if n.cfg.Accept != nil {
		sink = n.cfg.Accept(waitCtx, Incoming{Sender: sender, Files: files, Cancel: func() { n.cancelSession(s.id, ErrCanceled) }})
	}
	waitCancel()
	if sink == nil || r.Context().Err() != nil {
		if sink != nil {
			sink.Close(ErrCanceled)
		}
		n.endSession(s.id)
		http.Error(w, "declined", http.StatusForbidden)
		return
	}
	n.mu.Lock()
	s.sink = sink
	s.last = time.Now()
	resp := prepareResponse{SessionID: s.id, Files: map[string]string{}}
	for _, f := range files {
		tok := newToken()
		s.tokens[f.ID] = tok
		s.pending[f.ID] = true
		resp.Files[f.ID] = tok
	}
	n.mu.Unlock()
	go n.watchIdle(s)
	writeJSON(w, resp)
}

// watchIdle ends a session the sender abandoned.
func (n *Node) watchIdle(s *session) {
	t := time.NewTicker(15 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
			n.mu.Lock()
			idle := time.Since(s.last) > sessionIdle
			n.mu.Unlock()
			if idle {
				n.cancelSession(s.id, ErrCanceled)
				return
			}
		}
	}
}

func (n *Node) pin() string {
	if n.cfg.PIN == nil {
		return ""
	}
	return n.cfg.PIN()
}

func textMessage(files map[string]FileMeta) (string, bool) {
	if len(files) != 1 {
		return "", false
	}
	for _, f := range files {
		if strings.HasPrefix(f.FileType, "text/") && f.Preview != "" && int64(len(f.Preview)) >= f.Size {
			return f.Preview, true
		}
	}
	return "", false
}

func (n *Node) handleUpload(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	sid, fid, tok := q.Get("sessionId"), q.Get("fileId"), q.Get("token")
	n.mu.Lock()
	s := n.session
	if s == nil || s.id != sid || s.sink == nil {
		n.mu.Unlock()
		http.Error(w, "no session", http.StatusForbidden)
		return
	}
	if want, ok := s.tokens[fid]; !ok || want != tok || !s.pending[fid] {
		n.mu.Unlock()
		http.Error(w, "bad token", http.StatusForbidden)
		return
	}
	s.last = time.Now()
	n.mu.Unlock()

	body := &touchReader{r: r.Body, touch: func() {
		n.mu.Lock()
		s.last = time.Now()
		n.mu.Unlock()
	}}
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()
	go func() {
		select {
		case <-r.Context().Done():
			cancel()
		case <-ctx.Done():
		}
	}()
	if err := s.sink.Write(ctx, fid, body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		if s.ctx.Err() != nil {
			n.cancelSession(sid, ErrCanceled)
		}
		return
	}
	n.mu.Lock()
	delete(s.pending, fid)
	done := len(s.pending) == 0
	n.mu.Unlock()
	if done {
		s.sink.Close(nil)
		n.endSession(sid)
	}
	w.WriteHeader(http.StatusOK)
}

type touchReader struct {
	r     interface{ Read([]byte) (int, error) }
	touch func()
	last  time.Time
}

func (t *touchReader) Read(p []byte) (int, error) {
	n, err := t.r.Read(p)
	if now := time.Now(); now.Sub(t.last) > time.Second {
		t.last = now
		t.touch()
	}
	return n, err
}

func (n *Node) handleCancel(w http.ResponseWriter, r *http.Request) {
	n.cancelSession(r.URL.Query().Get("sessionId"), ErrCanceled)
	w.WriteHeader(http.StatusOK)
}

// cancelSession stops session id (any session when id is "").
func (n *Node) cancelSession(id string, err error) {
	n.mu.Lock()
	s := n.session
	if s == nil || (id != "" && s.id != id) {
		n.mu.Unlock()
		return
	}
	n.session = nil
	n.mu.Unlock()
	s.cancel()
	if s.sink != nil {
		s.sink.Close(err)
	}
}

func (n *Node) endSession(id string) {
	n.mu.Lock()
	if n.session != nil && n.session.id == id {
		n.session.cancel()
		n.session = nil
	}
	n.mu.Unlock()
}
