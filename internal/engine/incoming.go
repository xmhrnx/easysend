package engine

import (
	"context"
	"errors"
	"github.com/xmhrnx/giode/i18n"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"easysend/internal/item"
	"easysend/internal/localsend"
	"easysend/internal/transfer"
)

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

type ctxReader struct {
	ctx context.Context
	r   io.Reader
	fn  func(int64)
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

// onReceived finishes a completed incoming transfer: publishes its files
// (in the background: they may be large) and tells the user.
func (e *Engine) onReceived(t transfer.Transfer) {
	go func() {
		if e.cfg.Publish != nil {
			e.publish(t)
		}
		e.notify(i18n.Intl("Received from {name}", "name", t.Peer), t.Title()+" · "+item.HumanSize(t.Total))
	}()
}

// publish hands each saved file of t to Config.Publish and records where
// it went. A failure leaves the file where it was saved.
func (e *Engine) publish(t transfer.Transfer) {
	ctx := context.Background()
	for i, f := range t.Files {
		if f.Saved == "" {
			continue
		}
		var files []string
		_ = filepath.WalkDir(f.Saved, func(p string, d fs.DirEntry, err error) error {
			if err == nil && d.Type().IsRegular() {
				files = append(files, p)
			}
			return nil
		})
		var last string
		for _, p := range files {
			dst, err := e.cfg.Publish(ctx, p)
			if err != nil {
				e.Transfers.Update(t.ID, func(t *transfer.Transfer) { t.Error = i18n.Intl("Kept in app storage: {error}", "error", err.Error()) })
				continue
			}
			os.Remove(p)
			last = dst
		}
		if last != "" {
			if len(files) > 1 {
				os.RemoveAll(f.Saved) // the emptied folder
			}
			e.Transfers.Update(t.ID, func(t *transfer.Transfer) { t.Files[i].Saved = last })
		}
	}
}
