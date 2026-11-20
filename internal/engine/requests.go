package engine

import (
	"context"
	"time"

	"easysend/internal/item"
	"easysend/internal/localsend"
	"easysend/internal/settings"
)

// Request is an incoming nearby transfer waiting for the user.
type Request struct {
	ID     string
	Sender localsend.Device
	Files  []localsend.FileMeta
	Total  int64
	// Text is the message of a text-only request, which has no files.
	Text string
	// SyncFolder is the name of a folder the sender offers to sync, for a
	// sync invitation (which has no files); SyncRoot is where it will be
	// kept here, which AnswerSync can change.
	SyncFolder string
	SyncRoot   string
	Received   time.Time
	answer     chan bool
}

// Requests returns the transfers waiting for an answer, oldest first.
func (e *Engine) Requests() []Request {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]Request, len(e.requests))
	for i, r := range e.requests {
		out[i] = *r
	}
	return out
}

// Answer accepts or declines a request.
func (e *Engine) Answer(id string, accept bool) { e.answer(id, accept, "") }

// AnswerSync answers a sync invitation, keeping the folder at root (the
// suggested place when root is "").
func (e *Engine) AnswerSync(id string, accept bool, root string) { e.answer(id, accept, root) }

func (e *Engine) answer(id string, accept bool, root string) {
	e.mu.Lock()
	var r *Request
	for i, q := range e.requests {
		if q.ID == id {
			r = q
			e.requests = append(e.requests[:i], e.requests[i+1:]...)
			break
		}
	}
	if r != nil && root != "" {
		r.SyncRoot = root
	}
	e.mu.Unlock()
	if r != nil {
		r.answer <- accept
	}
	e.changed()
}

func (e *Engine) autoAccepts(sender localsend.Device) bool {
	s := e.Settings.Get()
	switch s.AutoAccept {
	case settings.AcceptEveryone:
		return true
	case settings.AcceptFavorites:
		_, ok := s.Favorites[sender.Fingerprint]
		return ok
	}
	return false
}

func (e *Engine) accept(ctx context.Context, in localsend.Incoming) localsend.Sink {
	req := &Request{ID: item.NewID(), Sender: in.Sender, Files: in.Files, Received: time.Now(), answer: make(chan bool, 1)}
	for _, f := range in.Files {
		req.Total += f.Size
	}
	if !e.ask(ctx, req, in.Sender.Alias+" wants to send you files", describeFiles(in.Files, req.Total)) {
		return nil
	}
	return e.newSink(in)
}

// ask waits for the user to accept req (at once for senders accepted
// automatically) and reports the answer; a sender giving up declines.
func (e *Engine) ask(ctx context.Context, req *Request, title, body string) bool {
	if req.SyncFolder == "" && e.autoAccepts(req.Sender) {
		return true
	}
	e.mu.Lock()
	e.requests = append(e.requests, req)
	e.mu.Unlock()
	e.changed()
	e.notify(title, body)
	select {
	case ok := <-req.answer:
		return ok
	case <-ctx.Done():
		e.Answer(req.ID, false)
		return false
	}
}
