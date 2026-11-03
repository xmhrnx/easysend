// Package transfer tracks transfers in progress and keeps their history.
package transfer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"easysend/internal/item"
)

// Direction says whether a transfer goes out or comes in.
type Direction int

const (
	// Outgoing transfers send from this device.
	Outgoing Direction = iota
	// Incoming transfers arrive at this device.
	Incoming
)

// Method is the channel a transfer uses.
type Method string

const (
	// Nearby is the LocalSend-compatible local network protocol.
	Nearby Method = "nearby"
	// Internet is Magic Wormhole.
	Internet Method = "internet"
	// Browser is the built-in web page.
	Browser Method = "browser"
)

// Status is where a transfer stands.
type Status int

const (
	// Waiting means the peer hasn't accepted or connected yet.
	Waiting Status = iota
	// Active means bytes are flowing.
	Active
	// Done means it finished successfully.
	Done
	// Failed means it stopped with an error.
	Failed
	// Canceled means a user stopped it.
	Canceled
	// Declined means the receiver refused it.
	Declined
)

// Finished reports whether the status is final.
func (s Status) Finished() bool { return s >= Done }

// File is one file of a transfer.
type File struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Done  int64  `json:"done"`
	Saved string `json:"saved,omitempty"`
}

// Transfer is one transfer's state.
type Transfer struct {
	ID        string    `json:"id"`
	Direction Direction `json:"direction"`
	Method    Method    `json:"method"`
	Peer      string    `json:"peer"`
	Files     []File    `json:"files"`
	Text      string    `json:"text,omitempty"`
	Total     int64     `json:"total"`
	Done      int64     `json:"done"`
	Status    Status    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Code      string    `json:"code,omitempty"`
	Started   time.Time `json:"started"`
	Finished  time.Time `json:"finished,omitempty"`
	Folder    string    `json:"folder,omitempty"`
	// Speed in bytes per second, smoothed; not persisted.
	Speed float64 `json:"-"`

	cancel   context.CancelFunc
	lastTick time.Time
	lastDone int64
}

// Progress is the completed fraction in [0, 1].
func (t Transfer) Progress() float32 {
	if t.Total <= 0 {
		if t.Status == Done {
			return 1
		}
		return 0
	}
	p := float32(t.Done) / float32(t.Total)
	if p > 1 {
		p = 1
	}
	return p
}

// ETA estimates the remaining time.
func (t Transfer) ETA() time.Duration {
	if t.Speed <= 1 {
		return -1
	}
	return time.Duration(float64(t.Total-t.Done) / t.Speed * float64(time.Second))
}

// Title is a short description: the single file's name or "3 files".
func (t Transfer) Title() string {
	switch {
	case t.Text != "" && len(t.Files) <= 1:
		return "Text message"
	case len(t.Files) == 1:
		return filepath.Base(t.Files[0].Name)
	case len(t.Files) == 0:
		return "Transfer"
	}
	return strconv.Itoa(len(t.Files)) + " files"
}

// Registry holds every transfer; it is safe for concurrent use.
type Registry struct {
	mu       sync.Mutex
	items    map[string]*Transfer
	path     string
	onChange func()
}

// NewRegistry loads the history from dir/history.json. onChange runs
// (on any goroutine) after every change.
func NewRegistry(dir string, onChange func()) *Registry {
	r := &Registry{items: map[string]*Transfer{}, path: filepath.Join(dir, "history.json"), onChange: onChange}
	if data, err := os.ReadFile(r.path); err == nil {
		var list []*Transfer
		if json.Unmarshal(data, &list) == nil {
			for _, t := range list {
				if !t.Status.Finished() {
					t.Status = Failed
					t.Error = "interrupted"
				}
				r.items[t.ID] = t
			}
		}
	}
	return r
}

// Start registers a new transfer and returns its ID.
func (r *Registry) Start(t Transfer, cancel context.CancelFunc) string {
	if t.ID == "" {
		t.ID = item.NewID()
	}
	if t.Started.IsZero() {
		t.Started = time.Now()
	}
	if t.Total == 0 {
		for _, f := range t.Files {
			t.Total += f.Size
		}
	}
	t.cancel = cancel
	t.lastTick = time.Now()
	r.mu.Lock()
	r.items[t.ID] = &t
	r.mu.Unlock()
	r.changed()
	return t.ID
}

// Update changes a transfer.
func (r *Registry) Update(id string, fn func(*Transfer)) {
	r.mu.Lock()
	t, ok := r.items[id]
	if ok {
		fn(t)
	}
	r.mu.Unlock()
	if ok {
		r.changed()
	}
}

// AddBytes records progress on file index i (or the transfer as a whole
// when i < 0).
func (r *Registry) AddBytes(id string, i int, n int64) {
	r.mu.Lock()
	t, ok := r.items[id]
	if ok {
		if t.Status == Waiting {
			t.Status = Active
		}
		t.Done += n
		if i >= 0 && i < len(t.Files) {
			t.Files[i].Done += n
		}
		now := time.Now()
		if dt := now.Sub(t.lastTick).Seconds(); dt >= 0.25 {
			inst := float64(t.Done-t.lastDone) / dt
			if t.Speed == 0 {
				t.Speed = inst
			} else {
				t.Speed = t.Speed*0.6 + inst*0.4
			}
			t.lastTick, t.lastDone = now, t.Done
		} else {
			ok = false // throttle redraws
		}
	}
	r.mu.Unlock()
	if ok {
		r.changed()
	}
}

// Finish marks a transfer final.
func (r *Registry) Finish(id string, status Status, err error) {
	r.Update(id, func(t *Transfer) {
		if t.Status.Finished() {
			return
		}
		t.Status = status
		t.Finished = time.Now()
		t.Speed = 0
		if err != nil {
			t.Error = err.Error()
		}
		if status == Done && t.Total > 0 {
			t.Done = t.Total
		}
	})
	r.save()
}

// Cancel stops a running transfer.
func (r *Registry) Cancel(id string) {
	r.mu.Lock()
	t, ok := r.items[id]
	var cancel context.CancelFunc
	if ok && !t.Status.Finished() {
		cancel = t.cancel
	}
	r.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	r.Finish(id, Canceled, nil)
}

// Remove deletes a finished transfer from the history.
func (r *Registry) Remove(id string) {
	r.mu.Lock()
	if t, ok := r.items[id]; ok && t.Status.Finished() {
		delete(r.items, id)
	}
	r.mu.Unlock()
	r.changed()
	r.save()
}

// ClearFinished empties the history.
func (r *Registry) ClearFinished() {
	r.mu.Lock()
	for id, t := range r.items {
		if t.Status.Finished() {
			delete(r.items, id)
		}
	}
	r.mu.Unlock()
	r.changed()
	r.save()
}

// Get returns a copy of one transfer.
func (r *Registry) Get(id string) (Transfer, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.items[id]
	if !ok {
		return Transfer{}, false
	}
	return copyOf(t), true
}

// List returns copies of every transfer, newest first.
func (r *Registry) List() []Transfer {
	r.mu.Lock()
	out := make([]Transfer, 0, len(r.items))
	for _, t := range r.items {
		out = append(out, copyOf(t))
	}
	r.mu.Unlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Started.After(out[j].Started) })
	return out
}

func copyOf(t *Transfer) Transfer {
	c := *t
	c.Files = append([]File(nil), t.Files...)
	c.cancel = nil
	return c
}

func (r *Registry) changed() {
	if r.onChange != nil {
		r.onChange()
	}
}

const historyLimit = 300

func (r *Registry) save() {
	list := r.List()
	var keep []Transfer
	for _, t := range list {
		if t.Status.Finished() {
			keep = append(keep, t)
		}
	}
	if len(keep) > historyLimit {
		keep = keep[:historyLimit]
	}
	data, err := json.Marshal(keep)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(r.path), 0o700)
	_ = os.WriteFile(r.path, data, 0o600)
}
