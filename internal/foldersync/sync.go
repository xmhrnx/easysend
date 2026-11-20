package foldersync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// syncFolder runs one round for a folder: scan it, then pull every
// device it's shared with that can be reached.
func (m *Manager) syncFolder(f *folder) {
	if !f.run.TryLock() {
		return
	}
	defer f.run.Unlock()

	m.mu.Lock()
	if _, still := m.folders[f.ID]; !still {
		m.mu.Unlock()
		return
	}
	prev := f.st.Local
	peers := append([]Peer(nil), f.Peers...)
	f.asked = nil
	f.status.Syncing, f.status.Done, f.status.Total = true, 0, 0
	m.mu.Unlock()
	m.changed()

	idx, scanned, err := scan(f.Root, prev, m.cfg.Self, time.Now())
	m.mu.Lock()
	if err != nil {
		f.status.Syncing, f.status.Err = false, "Can't read the folder: "+err.Error()
		m.mu.Unlock()
		m.changed()
		return
	}
	f.status.Err = ""
	f.st.Local = idx
	if scanned {
		_ = m.saveState(f)
	}
	m.mu.Unlock()

	fetched := false
	for _, p := range peers {
		n, err := m.pull(f, p)
		fetched = fetched || n > 0
		m.mu.Lock()
		ps := f.status.Peers[p.FP]
		switch {
		case errors.Is(err, errOffline):
			ps.Online, ps.Err = false, ""
		case err != nil:
			ps.Online, ps.Err = true, err.Error()
		default:
			ps.Online, ps.Err, ps.LastSync = true, "", time.Now()
		}
		f.status.Peers[p.FP] = ps
		m.mu.Unlock()
	}

	m.mu.Lock()
	f.status.Syncing = false
	f.status.LastSync = time.Now()
	f.status.Files, f.status.Bytes = 0, 0
	for _, e := range f.st.Local {
		if !e.Deleted {
			f.status.Files++
			f.status.Bytes += e.Size
		}
	}
	m.mu.Unlock()
	m.changed()

	// Changes made here, or fetched from one device, reach the others
	// at once rather than at their next round.
	if scanned || fetched {
		for _, p := range peers {
			if base, ok := m.address(p); ok {
				go m.sendChanged(p.FP, base, f.ID)
			}
		}
	}
}

var errOffline = errors.New("offline")

// pull brings a folder up to date with one device and returns how many
// files it fetched or removed.
func (m *Manager) pull(f *folder, p Peer) (int, error) {
	base, ok := m.address(p)
	if !ok {
		return 0, errOffline
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	remote, err := m.fetchIndex(ctx, p.FP, base, f.ID)
	cancel()
	if err != nil {
		var se *statusErr
		if errors.As(err, &se) {
			return 0, err // reachable, but turned us away
		}
		return 0, errOffline
	}
	m.mu.Lock()
	if base != p.Addr {
		// Remember where it was, for when discovery doesn't find it.
		for i := range f.Peers {
			if f.Peers[i].FP == p.FP {
				f.Peers[i].Addr = base
			}
		}
		_ = m.saveFolders()
	}
	agreed := f.st.Base[p.FP]
	if agreed == nil {
		agreed = map[string]string{}
		f.st.Base[p.FP] = agreed
	}
	acts := plan(f.st.Local, remote, agreed)
	work := 0
	for _, a := range acts {
		if a.kind != agree {
			work++
		}
	}
	f.status.Total += work
	m.mu.Unlock()
	if len(acts) == 0 {
		return 0, nil
	}

	n := 0
	var firstErr error
	for _, a := range acts {
		var err error
		switch a.kind {
		case agree:
			m.mu.Lock()
			agreed[a.path] = a.remote.version()
			m.mu.Unlock()
			continue
		case fetch, keepBoth:
			err = m.fetch(f, p, base, a)
		case remove:
			err = m.remove(f, a)
		}
		m.mu.Lock()
		f.status.Done++
		if err == nil {
			agreed[a.path] = a.remote.version()
			n++
		}
		m.mu.Unlock()
		m.changed()
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("%s: %w", a.path, err)
		}
	}
	m.mu.Lock()
	_ = m.saveState(f)
	m.mu.Unlock()
	return n, firstErr
}

// errChangedHere is a file that changed on this device while syncing:
// it's left for the next round to compare again.
var errChangedHere = errors.New("changed on this device meanwhile")

// fetch downloads a peer's version of a file, through a temporary file
// checked against its hash, over this device's copy (kept aside first as
// a conflict copy for keepBoth).
func (m *Manager) fetch(f *folder, p Peer, base string, a action) error {
	dest := filepath.Join(f.Root, filepath.FromSlash(a.path))
	tmpDir := filepath.Join(f.Root, internalDir)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(tmpDir, "*.part")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()
	body, err := m.fetchFile(ctx, p.FP, base, f.ID, a.path)
	if err != nil {
		tmp.Close()
		return err
	}
	h := sha256.New()
	_, err = io.Copy(io.MultiWriter(tmp, h), body)
	body.Close()
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	if hex.EncodeToString(h.Sum(nil)) != a.remote.Hash {
		return errors.New("changed on the other device while downloading")
	}

	m.mu.Lock()
	local, had := f.st.Local[a.path]
	m.mu.Unlock()
	if !onDisk(dest, local, had) {
		return errChangedHere
	}
	if a.kind == keepBoth {
		copyRel := m.conflictName(f.Root, a.path)
		if err := os.Rename(dest, filepath.Join(f.Root, filepath.FromSlash(copyRel))); err != nil {
			return err
		}
		m.mu.Lock()
		e := local
		e.Changed, e.By = time.Now().UnixNano(), m.cfg.Self
		f.st.Local[copyRel] = e
		m.mu.Unlock()
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), dest); err != nil {
		return err
	}
	if a.remote.MTime != 0 {
		mt := time.Unix(0, a.remote.MTime)
		_ = os.Chtimes(dest, mt, mt)
	}
	info, err := os.Stat(dest)
	if err != nil {
		return err
	}
	m.mu.Lock()
	f.st.Local[a.path] = Entry{Size: info.Size(), MTime: info.ModTime().UnixNano(), Hash: a.remote.Hash, Changed: a.remote.Changed, By: a.remote.By}
	m.mu.Unlock()
	return nil
}

// remove deletes this device's copy of a file the peer deleted, unless
// it changed here meanwhile, and folders left empty by it.
func (m *Manager) remove(f *folder, a action) error {
	dest := filepath.Join(f.Root, filepath.FromSlash(a.path))
	m.mu.Lock()
	local, had := f.st.Local[a.path]
	m.mu.Unlock()
	if !onDisk(dest, local, had) {
		return errChangedHere
	}
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return err
	}
	for dir := filepath.Dir(dest); dir != f.Root && strings.HasPrefix(dir, f.Root); dir = filepath.Dir(dir) {
		if os.Remove(dir) != nil { // only succeeds when empty
			break
		}
	}
	m.mu.Lock()
	f.st.Local[a.path] = Entry{Deleted: true, Changed: a.remote.Changed, By: a.remote.By}
	m.mu.Unlock()
	return nil
}

// onDisk reports whether the file at p is still the one the index knows:
// absent for no entry or a deletion, else the same size and time.
func onDisk(p string, e Entry, had bool) bool {
	info, err := os.Stat(p)
	if !had || e.Deleted {
		return os.IsNotExist(err)
	}
	return err == nil && info.Mode().IsRegular() && info.Size() == e.Size && info.ModTime().UnixNano() == e.MTime
}

// conflictName is a free name for this device's version of a file both
// devices edited: "report (conflict, Silver Fox 2026-11-20 14.05).pdf".
func (m *Manager) conflictName(root, rel string) string {
	dir, name := path.Split(rel)
	ext := path.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	device := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, r) || r < 32 {
			return -1
		}
		return r
	}, m.name())
	label := fmt.Sprintf("conflict, %s %s", device, time.Now().Format("2006-01-02 15.04"))
	for i := 1; ; i++ {
		n := fmt.Sprintf("%s (%s)%s", stem, label, ext)
		if i > 1 {
			n = fmt.Sprintf("%s (%s %d)%s", stem, label, i, ext)
		}
		rel := dir + n
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); os.IsNotExist(err) {
			return rel
		}
	}
}
