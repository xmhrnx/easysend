package item

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// SafeJoin turns an untrusted, slash-separated name received from a peer
// into a path inside dir: absolute paths, "..", drive letters and
// characters invalid on Windows are removed.
func SafeJoin(dir, name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	parts := strings.Split(path.Clean("/"+name), "/")
	clean := make([]string, 0, len(parts))
	for _, p := range parts {
		p = sanitizeComponent(p)
		if p == "" || p == "." || p == ".." {
			continue
		}
		clean = append(clean, p)
	}
	if len(clean) == 0 {
		clean = []string{"file"}
	}
	return filepath.Join(append([]string{dir}, clean...)...)
}

func sanitizeComponent(s string) string {
	s = strings.Map(func(r rune) rune {
		switch r {
		case '<', '>', ':', '"', '|', '?', '*':
			return '_'
		}
		if r < 32 {
			return -1
		}
		return r
	}, s)
	return strings.TrimRight(strings.TrimSpace(s), ".")
}

// Unique returns p, or p with " (2)", " (3)", ... before the extension
// when a file already exists there.
func Unique(p string) string {
	if _, err := os.Lstat(p); os.IsNotExist(err) {
		return p
	}
	ext := filepath.Ext(p)
	base := strings.TrimSuffix(p, ext)
	for i := 2; ; i++ {
		c := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Lstat(c); os.IsNotExist(err) {
			return c
		}
	}
}

// UniqueTopLevel makes the first path element of a received name unique
// inside dir, so a received folder "Photos" lands in "Photos (2)" when
// "Photos" exists, while every file of the same transfer keeps landing
// in the same renamed folder. renames remembers choices per transfer.
func UniqueTopLevel(dir, name string, renames map[string]string) string {
	full := SafeJoin(dir, name)
	rel, err := filepath.Rel(dir, full)
	if err != nil {
		return Unique(full)
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) == 1 {
		return Unique(full)
	}
	top := parts[0]
	if r, ok := renames[top]; ok {
		parts[0] = r
	} else {
		r := filepath.Base(Unique(filepath.Join(dir, top)))
		renames[top] = r
		parts[0] = r
	}
	return filepath.Join(append([]string{dir}, parts...)...)
}
