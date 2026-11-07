package localsend

import (
	"context"
	"net/http"
	"time"
)

type touchReader struct {
	r     interface{ Read([]byte) (int, error) }
	touch func()
	last  time.Time
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
