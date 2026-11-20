// Package foldersync keeps a folder the same on several EasySend devices,
// peer to peer: each device indexes its copy, pulls the indexes of the
// devices it shares the folder with, and fetches, deletes or keeps files
// by comparing the two against the last version they agreed on.
package foldersync

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Entry is one file of a synced folder as a device knows it.
type Entry struct {
	Size int64 `json:"size,omitempty"`
	// MTime is the file's modification time on this device, in Unix
	// nanoseconds: with Size, how a scan tells the file is unchanged.
	MTime   int64  `json:"mtime,omitempty"`
	Hash    string `json:"hash,omitempty"`
	Deleted bool   `json:"deleted,omitempty"`
	// Changed is when this version came to be, in Unix nanoseconds, and
	// By the device it happened on: they carry over as the version
	// spreads, and decide which of two edits made apart is the newer.
	Changed int64  `json:"changed"`
	By      string `json:"by,omitempty"`
}

// deletedVersion stands for a deleted file in version comparisons.
const deletedVersion = "-"

// version is what two copies are compared by: the content's hash, or
// deletedVersion. Two entries with the same version agree whatever
// their times.
func (e Entry) version() string {
	if e.Deleted {
		return deletedVersion
	}
	return e.Hash
}

// Index is a folder's files by path: slash-separated, relative to the
// folder.
type Index map[string]Entry

// internalDir holds EasySend's temporary files inside a synced folder;
// it's never synced.
const internalDir = ".easysend-sync"

// tombstoneAge is how long a deleted file is remembered, so devices that
// were away still learn about the deletion.
const tombstoneAge = 30 * 24 * time.Hour

// ignored reports whether a file or folder name is left out of syncing:
// EasySend's own and the systems' bookkeeping.
func ignored(name string) bool {
	switch name {
	case internalDir, ".DS_Store", "Thumbs.db", "desktop.ini", ".localized":
		return true
	}
	return strings.HasPrefix(name, "._") // macOS resource forks on foreign disks
}

// scan walks root and returns its index: unchanged files (same size and
// time) keep their entry from prev, others are hashed; files gone since
// prev become deletions. changed reports whether anything differs.
func scan(root string, prev Index, self string, now time.Time) (idx Index, changed bool, err error) {
	idx = make(Index, len(prev))
	seen := make(map[string]bool, len(prev))
	unreadable := map[string]bool{}
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, werr error) error {
		if werr != nil {
			if p == root {
				return werr
			}
			// A folder that can't be read: keep what was known under it,
			// rather than take its files for deleted.
			if rel, rerr := relPath(root, p); rerr == nil {
				unreadable[rel] = true
			}
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if p == root {
			return nil
		}
		if ignored(d.Name()) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil // folders are walked; links and devices aren't synced
		}
		rel, rerr := relPath(root, p)
		if rerr != nil {
			return nil
		}
		info, ierr := d.Info()
		if ierr != nil {
			unreadable[rel] = true
			return nil
		}
		seen[rel] = true
		size, mtime := info.Size(), info.ModTime().UnixNano()
		old, had := prev[rel]
		if had && !old.Deleted && old.Size == size && old.MTime == mtime {
			idx[rel] = old
			return nil
		}
		hash, herr := hashFile(p)
		if herr != nil {
			unreadable[rel] = true
			if had {
				idx[rel] = old
			}
			return nil
		}
		if had && !old.Deleted && old.Hash == hash {
			// Touched, same content: same version.
			old.Size, old.MTime = size, mtime
			idx[rel] = old
			changed = true
			return nil
		}
		idx[rel] = Entry{Size: size, MTime: mtime, Hash: hash, Changed: now.UnixNano(), By: self}
		changed = true
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	for rel, old := range prev {
		if seen[rel] {
			continue
		}
		if underAny(rel, unreadable) {
			idx[rel] = old
			continue
		}
		if old.Deleted {
			if now.Sub(time.Unix(0, old.Changed)) < tombstoneAge {
				idx[rel] = old
			} else {
				changed = true
			}
			continue
		}
		idx[rel] = Entry{Deleted: true, Changed: now.UnixNano(), By: self}
		changed = true
	}
	return idx, changed, nil
}

// underAny reports whether rel is one of paths or inside one of them.
func underAny(rel string, paths map[string]bool) bool {
	for p := range paths {
		if rel == p || strings.HasPrefix(rel, p+"/") {
			return true
		}
	}
	return false
}

func relPath(root, p string) (string, error) {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

func hashFile(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// cleanPath checks a path from a peer: relative, slash-separated, inside
// the folder and not in EasySend's own folder. It returns "" otherwise.
func cleanPath(p string) string {
	if p == "" || strings.Contains(p, "\\") || strings.HasPrefix(p, "/") || strings.ContainsRune(p, 0) {
		return ""
	}
	c := path.Clean(p)
	if c != p || c == "." || c == ".." || strings.HasPrefix(c, "../") {
		return ""
	}
	for _, part := range strings.Split(c, "/") {
		if part == "" || part == "." || part == ".." || ignored(part) {
			return ""
		}
		// A drive or volume name would leave the folder on Windows.
		if strings.Contains(part, ":") {
			return ""
		}
	}
	return c
}
