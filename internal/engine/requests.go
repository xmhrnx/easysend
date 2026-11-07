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
	ID       string
	Sender   localsend.Device
	Files    []localsend.FileMeta
	Total    int64
	Received time.Time
	answer   chan bool
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
func (e *Engine) Answer(id string, accept bool) {
	e.mu.Lock()
	var r *Request
	for i, q := range e.requests {
		if q.ID == id {
			r = q
			e.requests = append(e.requests[:i], e.requests[i+1:]...)
			break
		}
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
	if !e.autoAccepts(in.Sender) {
		e.mu.Lock()
		e.requests = append(e.requests, req)
		e.mu.Unlock()
		e.changed()
		e.notify(in.Sender.Alias+" wants to send you files", describeFiles(in.Files, req.Total))
		select {
		case ok := <-req.answer:
			if !ok {
				return nil
			}
		case <-ctx.Done():
			e.Answer(req.ID, false)
			return nil
		}
	}
	return e.newSink(in)
}
