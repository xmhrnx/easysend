package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// Store holds the settings and saves them on every change.
type Store struct {
	mu   sync.RWMutex
	path string
	s    Settings
}

// Open loads settings from dir/settings.json, creating them with
// defaults when missing.
func Open(dir, downloadDir string) (*Store, error) {
	st := &Store{path: filepath.Join(dir, "settings.json"), s: Defaults(downloadDir)}
	data, err := os.ReadFile(st.path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return st, st.save()
	case err != nil:
		return st, err
	}
	if err := json.Unmarshal(data, &st.s); err != nil {
		return st, err
	}
	if st.s.Favorites == nil {
		st.s.Favorites = map[string]string{}
	}
	if st.s.Fingerprint == "" {
		st.s.Fingerprint = randomHex(16)
	}
	if st.s.DropToken == "" {
		st.s.DropToken = randomHex(12)
	}
	if st.s.DownloadDir == "" {
		st.s.DownloadDir = downloadDir
	}
	return st, nil
}

// Get returns a copy of the settings.
func (st *Store) Get() Settings {
	st.mu.RLock()
	defer st.mu.RUnlock()
	s := st.s
	s.Favorites = make(map[string]string, len(st.s.Favorites))
	for k, v := range st.s.Favorites {
		s.Favorites[k] = v
	}
	s.Manual = append([]string(nil), st.s.Manual...)
	return s
}

// Update changes the settings and saves them.
func (st *Store) Update(fn func(*Settings)) error {
	st.mu.Lock()
	fn(&st.s)
	st.mu.Unlock()
	return st.save()
}

func (st *Store) save() error {
	st.mu.RLock()
	data, err := json.MarshalIndent(st.s, "", "  ")
	st.mu.RUnlock()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(st.path), 0o700); err != nil {
		return err
	}
	tmp := st.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, st.path)
}
