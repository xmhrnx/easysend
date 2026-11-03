// Package wormhole sends and receives over the internet with the Magic
// Wormhole protocol: a short code pairs two devices anywhere, the
// transfer is end-to-end encrypted and goes directly between them when
// the networks allow, through a relay otherwise. It interoperates with
// the wormhole command line tool and the other Magic Wormhole apps.
package wormhole

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	ww "github.com/psanford/wormhole-william/wormhole"

	"easysend/internal/item"
)

// Options select the servers; empty fields use the public defaults.
type Options struct {
	RendezvousURL string
	TransitRelay  string
}

func (o Options) client() *ww.Client {
	return &ww.Client{
		AppID:                     ww.WormholeCLIAppID,
		RendezvousURL:             o.RendezvousURL,
		TransitRelayAddress:       o.TransitRelay,
		PassPhraseComponentLength: 2,
	}
}

// Send offers items and returns the code the receiver types. The returned
// channel yields the outcome once the receiver took them (nil) or the
// transfer failed.
func Send(ctx context.Context, o Options, items []item.Item, progress func(sent, total int64)) (string, <-chan error, error) {
	if len(items) == 0 {
		return "", nil, errors.New("nothing to send")
	}
	c := o.client()
	opt := ww.WithProgress(progress)
	var (
		code string
		res  chan ww.SendResult
		err  error
		file *os.File
	)
	switch {
	case len(items) == 1 && items[0].Kind == item.Text:
		code, res, err = c.SendText(ctx, items[0].Text, opt)
	case len(items) == 1:
		file, err = os.Open(items[0].Path)
		if err != nil {
			return "", nil, err
		}
		code, res, err = c.SendFile(ctx, items[0].BaseName(), file, opt)
	default:
		code, res, err = c.SendDirectory(ctx, folderName(items), entries(items), opt)
	}
	if err != nil {
		if file != nil {
			file.Close()
		}
		return "", nil, err
	}
	out := make(chan error, 1)
	go func() {
		if file != nil {
			defer file.Close()
		}
		select {
		case r := <-res:
			if !r.OK && r.Error == nil {
				r.Error = errors.New("transfer failed")
			}
			out <- r.Error
		case <-ctx.Done():
			out <- ctx.Err()
		}
	}()
	return code, out, nil
}

func folderName(items []item.Item) string {
	group := items[0].Group
	for _, it := range items {
		if it.Group != group {
			group = ""
			break
		}
	}
	if group != "" {
		return group
	}
	return fmt.Sprintf("EasySend (%d files)", len(items))
}

// entries lists items for a directory transfer. Items from one folder
// drop the folder's own name, which becomes the directory's.
func entries(items []item.Item) []ww.DirectoryEntry {
	single := folderName(items) == items[0].Group && items[0].Group != ""
	out := make([]ww.DirectoryEntry, 0, len(items))
	for _, it := range items {
		name := it.Name
		if single {
			name = strings.TrimPrefix(name, it.Group+"/")
		}
		it := it
		out = append(out, ww.DirectoryEntry{Path: name, Mode: 0o644, Reader: it.Open})
	}
	return out
}

// Offer is what a sender offered.
type Offer struct {
	Name  string
	Size  int64
	Files int
	// IsText is set for a text message.
	IsText bool
	// IsFolder is set for several files.
	IsFolder bool
	msg      *ww.IncomingMessage
}

// Receive connects with code and returns the offer; Save or Reject it.
func Receive(ctx context.Context, o Options, code string) (*Offer, error) {
	code = NormalizeCode(code)
	msg, err := o.client().Receive(ctx, code)
	if err != nil {
		return nil, err
	}
	of := &Offer{Name: msg.Name, Size: msg.UncompressedBytes64, Files: msg.FileCount, msg: msg}
	switch msg.Type {
	case ww.TransferText:
		of.IsText = true
	case ww.TransferDirectory:
		of.IsFolder = true
	default:
		of.Files = 1
	}
	return of, nil
}

// NormalizeCode accepts codes typed with spaces or pasted as a
// wormhole-transfer: link.
func NormalizeCode(code string) string {
	code = strings.TrimSpace(code)
	code = strings.TrimPrefix(code, "wormhole-transfer:")
	code = strings.Join(strings.Fields(code), "-")
	return strings.ToLower(code)
}

// Reject declines the offer.
func (o *Offer) Reject() error { return o.msg.Reject() }

// ReadText returns a text message.
func (o *Offer) ReadText() (string, error) {
	b, err := io.ReadAll(io.LimitReader(o.msg, 16<<20))
	return string(b), err
}

// Save stores the offered file or folder in dir and returns the path.
func (o *Offer) Save(ctx context.Context, dir string, progress func(n int64)) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	r := &countingReader{ctx: ctx, r: o.msg, fn: progress}
	if !o.IsFolder {
		dst := item.Unique(item.SafeJoin(dir, filepath.Base(o.Name)))
		return dst, writeFile(dst, r)
	}
	tmp, err := os.CreateTemp(dir, ".easysend-*.zip")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if _, err := io.Copy(tmp, r); err != nil {
		return "", err
	}
	target := item.Unique(item.SafeJoin(dir, o.Name))
	return target, unzip(tmp, target)
}

func writeFile(dst string, r io.Reader) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(dst)
		return err
	}
	return f.Close()
}

func unzip(f *os.File, target string) error {
	st, err := f.Stat()
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(f, st.Size())
	if err != nil {
		return err
	}
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		dst := item.SafeJoin(target, zf.Name)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		err = writeFile(dst, rc)
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

type countingReader struct {
	ctx context.Context
	r   io.Reader
	fn  func(int64)
}

func (c *countingReader) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := c.r.Read(p)
	if n > 0 && c.fn != nil {
		c.fn(int64(n))
	}
	return n, err
}
