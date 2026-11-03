package engine

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"easysend/internal/item"
	"easysend/internal/localsend"
	"easysend/internal/settings"
	"easysend/internal/transfer"
)

// Peer is a nearby device as the UI shows it.
type Peer struct {
	localsend.Device
	Favorite bool
}

// Peers returns the nearby devices, favorites first.
func (e *Engine) Peers() []Peer {
	favs := e.Settings.Get().Favorites
	devs := e.node.Devices()
	out := make([]Peer, 0, len(devs))
	for _, d := range devs {
		_, fav := favs[d.Fingerprint]
		out = append(out, Peer{Device: d, Favorite: fav})
	}
	// Stable: favorites first, keeping the name order within each group.
	sorted := make([]Peer, 0, len(out))
	for _, want := range []bool{true, false} {
		for _, p := range out {
			if p.Favorite == want {
				sorted = append(sorted, p)
			}
		}
	}
	return sorted
}

// Scanning reports whether a network scan is running.
func (e *Engine) Scanning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.scanning
}

// Refresh looks for nearby devices again.
func (e *Engine) Refresh() {
	e.mu.Lock()
	if e.scanning {
		e.mu.Unlock()
		return
	}
	e.scanning = true
	e.mu.Unlock()
	e.changed()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		for _, addr := range e.Settings.Get().Manual {
			_, _ = e.node.Register(ctx, addr, true)
		}
		e.node.Refresh(ctx)
		e.mu.Lock()
		e.scanning = false
		e.mu.Unlock()
		e.changed()
	}()
}

// AddByAddress adds a device by IP address (and optional port), for
// networks where discovery doesn't reach it.
func (e *Engine) AddByAddress(ctx context.Context, addr string) (localsend.Device, error) {
	d, err := e.node.Register(ctx, addr, true)
	if err != nil {
		return d, err
	}
	_ = e.Settings.Update(func(s *settings.Settings) {
		for _, a := range s.Manual {
			if a == d.Addr() {
				return
			}
		}
		s.Manual = append(s.Manual, d.Addr())
	})
	return d, nil
}

// SetFavorite marks a device as a favorite (auto-accept candidates).
func (e *Engine) SetFavorite(d localsend.Device, on bool) {
	_ = e.Settings.Update(func(s *settings.Settings) {
		if on {
			s.Favorites[d.Fingerprint] = d.Alias
		} else {
			delete(s.Favorites, d.Fingerprint)
		}
	})
	e.changed()
}

// SendNearby sends items to a nearby device in the background and returns
// the transfer's ID. A failure with "PIN required" means the receiver
// wants a PIN: send again with it.
func (e *Engine) SendNearby(d localsend.Device, items []item.Item, pin string) string {
	ctx, cancel := context.WithCancel(context.Background())
	files := make([]transfer.File, len(items))
	for i, it := range items {
		files[i] = transfer.File{Name: it.Name, Size: it.Size}
	}
	t := transfer.Transfer{Direction: transfer.Outgoing, Method: transfer.Nearby, Peer: d.Alias, Files: files, Status: transfer.Waiting}
	if item.OnlyText(items) && len(items) == 1 {
		t.Text = items[0].Text
	}
	id := e.Transfers.Start(t, cancel)
	go func() {
		defer cancel()
		err := e.node.Send(ctx, d, items, pin, localsend.SendEvents{
			Accepted: func() { e.Transfers.Update(id, func(t *transfer.Transfer) { t.Status = transfer.Active }) },
			Progress: func(i int, n int64) { e.Transfers.AddBytes(id, i, n) },
		})
		switch {
		case err == nil:
			e.Transfers.Finish(id, transfer.Done, nil)
		case ctx.Err() != nil || localsend.IsCanceled(err):
			e.Transfers.Finish(id, transfer.Canceled, nil)
		case errors.Is(err, localsend.ErrDeclined):
			e.Transfers.Finish(id, transfer.Declined, errors.New(errorText(err)))
		default:
			e.Transfers.Finish(id, transfer.Failed, errors.New(errorText(err)))
		}
	}()
	return id
}

// IsPINError reports whether a transfer failed because a PIN is needed.
func IsPINError(t transfer.Transfer) bool {
	return t.Status == transfer.Failed && t.Error == errorText(localsend.ErrPINRequired)
}

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

func describeFiles(files []localsend.FileMeta, total int64) string {
	if len(files) == 1 {
		return files[0].FileName + " · " + item.HumanSize(total)
	}
	return strconv.Itoa(len(files)) + " files · " + item.HumanSize(total)
}

// sink saves the files of an accepted nearby transfer.
type sink struct {
	e       *Engine
	id      string
	dir     string
	index   map[string]int
	names   map[string]string
	mu      sync.Mutex
	renames map[string]string
}

func (e *Engine) newSink(in localsend.Incoming) *sink {
	s := &sink{e: e, dir: e.Settings.Get().DownloadDir, index: map[string]int{}, names: map[string]string{}, renames: map[string]string{}}
	files := make([]transfer.File, len(in.Files))
	for i, f := range in.Files {
		s.index[f.ID] = i
		s.names[f.ID] = f.FileName
		files[i] = transfer.File{Name: f.FileName, Size: f.Size}
	}
	cancel := func() {
		if in.Cancel != nil {
			in.Cancel()
		}
	}
	s.id = e.Transfers.Start(transfer.Transfer{Direction: transfer.Incoming, Method: transfer.Nearby, Peer: in.Sender.Alias, Files: files, Status: transfer.Active, Folder: s.dir}, cancel)
	return s
}

func (s *sink) Write(ctx context.Context, id string, r io.Reader) error {
	i, ok := s.index[id]
	if !ok {
		return errors.New("unknown file")
	}
	s.mu.Lock()
	dst := item.UniqueTopLevel(s.dir, s.names[id], s.renames)
	s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, &ctxReader{ctx: ctx, r: r, fn: func(n int64) { s.e.Transfers.AddBytes(s.id, i, n) }})
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(dst)
		return err
	}
	s.e.Transfers.Update(s.id, func(t *transfer.Transfer) { t.Files[i].Saved = dst })
	return nil
}

func (s *sink) Close(err error) {
	switch {
	case err == nil:
		s.e.Transfers.Finish(s.id, transfer.Done, nil)
		if t, ok := s.e.Transfers.Get(s.id); ok {
			s.e.onReceived(t)
		}
	case localsend.IsCanceled(err):
		s.e.Transfers.Finish(s.id, transfer.Canceled, nil)
	default:
		s.e.Transfers.Finish(s.id, transfer.Failed, err)
	}
}

type ctxReader struct {
	ctx context.Context
	r   io.Reader
	fn  func(int64)
}

func (c *ctxReader) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := c.r.Read(p)
	if n > 0 {
		c.fn(int64(n))
	}
	return n, err
}

func (e *Engine) onReceived(t transfer.Transfer) {
	e.notify("Received from "+t.Peer, t.Title()+" · "+item.HumanSize(t.Total))
}

func (e *Engine) recordText(method transfer.Method, from, text string) {
	id := e.Transfers.Start(transfer.Transfer{Direction: transfer.Incoming, Method: method, Peer: from, Text: text,
		Files: []transfer.File{{Name: "Text message", Size: int64(len(text))}}, Status: transfer.Active}, nil)
	e.Transfers.Finish(id, transfer.Done, nil)
	preview := text
	if len(preview) > 120 {
		preview = preview[:120] + "…"
	}
	e.notify("Message from "+from, preview)
}

func (e *Engine) onNearbyText(from localsend.Device, text string) {
	e.recordText(transfer.Nearby, from.Alias, text)
}

func (e *Engine) onBrowserText(from, text string) {
	e.recordText(transfer.Browser, from, text)
}
