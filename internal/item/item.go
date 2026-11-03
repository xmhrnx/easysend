// Package item describes the things a user sends: files (with an optional
// path inside a sent folder) and text snippets.
package item

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/fs"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Kind tells a file from a text snippet.
type Kind int

const (
	// File is a file on disk.
	File Kind = iota
	// Text is a text snippet held in memory.
	Text
)

// Item is one thing to send.
type Item struct {
	ID string
	// Kind is File or Text.
	Kind Kind
	// Name is the name the receiver sees. For files inside a folder it is
	// the slash-separated path relative to the folder's parent
	// ("Photos/2024/a.jpg").
	Name string
	// Path is the file on disk (File only).
	Path string
	// Size in bytes.
	Size int64
	// MIME type.
	MIME string
	// Text is the snippet (Text only).
	Text string
	// Group names the folder an item was added with, for display.
	Group string
}

// NewID returns a random identifier.
func NewID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// FromFile describes a single file.
func FromFile(p string) (Item, error) {
	st, err := os.Stat(p)
	if err != nil {
		return Item{}, err
	}
	name := filepath.Base(p)
	return Item{ID: NewID(), Kind: File, Name: name, Path: p, Size: st.Size(), MIME: MIMEOf(name)}, nil
}

// FromText describes a text snippet.
func FromText(text string) Item {
	return Item{ID: NewID(), Kind: Text, Name: "Text.txt", Size: int64(len(text)), MIME: "text/plain", Text: text}
}

// FromFolder lists every regular file under dir, named relative to dir's
// parent so the receiver recreates the folder.
func FromFolder(dir string) ([]Item, error) {
	dir = filepath.Clean(dir)
	parent := filepath.Dir(dir)
	group := filepath.Base(dir)
	var out []Item
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(parent, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		out = append(out, Item{ID: NewID(), Kind: File, Name: rel, Path: p, Size: info.Size(), MIME: MIMEOf(p), Group: group})
		return nil
	})
	return out, err
}

// FromPath describes a file or, for a directory, every file in it.
func FromPath(p string) ([]Item, error) {
	st, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	if st.IsDir() {
		return FromFolder(p)
	}
	it, err := FromFile(p)
	if err != nil {
		return nil, err
	}
	return []Item{it}, nil
}

// Open returns the item's content.
func (it Item) Open() (io.ReadCloser, error) {
	if it.Kind == Text {
		return io.NopCloser(strings.NewReader(it.Text)), nil
	}
	return os.Open(it.Path)
}

// BaseName is the last element of Name.
func (it Item) BaseName() string { return path.Base(it.Name) }

// MIMEOf guesses a MIME type from a file name.
func MIMEOf(name string) string {
	if t := mime.TypeByExtension(strings.ToLower(filepath.Ext(name))); t != "" {
		if i := strings.IndexByte(t, ';'); i > 0 {
			t = t[:i]
		}
		return t
	}
	return "application/octet-stream"
}

// TotalSize sums the sizes of items.
func TotalSize(items []Item) int64 {
	var n int64
	for _, it := range items {
		n += it.Size
	}
	return n
}

// OnlyText reports whether items are all text snippets.
func OnlyText(items []Item) bool {
	for _, it := range items {
		if it.Kind != Text {
			return false
		}
	}
	return len(items) > 0
}
