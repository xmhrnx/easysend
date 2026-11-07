package wormhole

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"

	"easysend/internal/item"
)

type countingReader struct {
	ctx context.Context
	r   io.Reader
	fn  func(int64)
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
