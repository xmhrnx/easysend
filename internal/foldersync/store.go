package foldersync

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// Folder is a synced folder: where it is on this device, and the devices
// it's shared with. Every device has the same ID for it.
type Folder struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Root    string    `json:"root"`
	Peers   []Peer    `json:"peers"`
	Created time.Time `json:"created"`
}

// Peer is a device a folder is shared with.
type Peer struct {
	// FP is the device's certificate fingerprint: the only devices that
	// can read the folder from this one, and that this one trusts.
	FP   string `json:"fp"`
	Name string `json:"name"`
	// Addr is the device's last known address ("https://ip:port"), for
	// when discovery hasn't found it.
	Addr string `json:"addr,omitempty"`
}

// state is what a device remembers of a folder between runs.
type state struct {
	Local Index `json:"local"`
	// Base holds, by peer fingerprint, the version of each file the two
	// devices last agreed on.
	Base map[string]map[string]string `json:"base"`
}

func (m *Manager) foldersPath() string { return filepath.Join(m.cfg.Dir, "folders.json") }

func (m *Manager) statePath(id string) string { return filepath.Join(m.cfg.Dir, id+".json") }

// load reads the folders and their states; a missing file is no folders.
func (m *Manager) load() error {
	var folders []Folder
	if err := readJSON(m.foldersPath(), &folders); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, f := range folders {
		st := state{}
		_ = readJSON(m.statePath(f.ID), &st) // a lost state means a fresh compare, never a loss
		if st.Local == nil {
			st.Local = Index{}
		}
		if st.Base == nil {
			st.Base = map[string]map[string]string{}
		}
		m.folders[f.ID] = &folder{Folder: f, st: st, status: Status{Peers: map[string]PeerStatus{}}}
	}
	return nil
}

// saveFolders writes the folder list; call with m.mu held.
func (m *Manager) saveFolders() error {
	list := make([]Folder, 0, len(m.folders))
	for _, f := range m.folders {
		list = append(list, f.Folder)
	}
	return writeJSON(m.foldersPath(), list)
}

// saveState writes a folder's index and agreed versions; call with m.mu
// held.
func (m *Manager) saveState(f *folder) error {
	return writeJSON(m.statePath(f.ID), f.st)
}

func readJSON(p string, v any) error {
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// writeJSON writes v to p through a temporary file, so a crash never
// leaves half a file.
func writeJSON(p string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
