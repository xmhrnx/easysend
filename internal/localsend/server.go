package localsend

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

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
