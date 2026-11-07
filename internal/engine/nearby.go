package engine

import (
	"context"
	"errors"
	"strconv"
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

func describeFiles(files []localsend.FileMeta, total int64) string {
	if len(files) == 1 {
		return files[0].FileName + " · " + item.HumanSize(total)
	}
	return strconv.Itoa(len(files)) + " files · " + item.HumanSize(total)
}
