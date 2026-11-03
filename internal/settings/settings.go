// Package settings stores EasySend's preferences as JSON.
package settings

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// AutoAccept says whose incoming nearby transfers are saved without asking.
type AutoAccept int

const (
	// AcceptAsk always asks.
	AcceptAsk AutoAccept = iota
	// AcceptFavorites saves transfers from favorite devices without asking.
	AcceptFavorites
	// AcceptEveryone saves every transfer without asking.
	AcceptEveryone
)

// Theme is the color mode preference.
type Theme int

const (
	// ThemeSystem follows the platform.
	ThemeSystem Theme = iota
	// ThemeDark is always dark.
	ThemeDark
	// ThemeLight is always light.
	ThemeLight
)

// Settings are the user's preferences.
type Settings struct {
	DeviceName  string `json:"deviceName"`
	Fingerprint string `json:"fingerprint"`
	Theme       Theme  `json:"theme"`
	Accent      int    `json:"accent"`

	DownloadDir   string     `json:"downloadDir"`
	AutoAccept    AutoAccept `json:"autoAccept"`
	RequirePIN    bool       `json:"requirePin"`
	PIN           string     `json:"pin"`
	Discoverable  bool       `json:"discoverable"`
	Notifications bool       `json:"notifications"`

	MinimizeToTray bool `json:"minimizeToTray"`
	CloseToTray    bool `json:"closeToTray"`
	StartHidden    bool `json:"startHidden"`

	Port        int    `json:"port"`
	WebPort     int    `json:"webPort"`
	UPnP        bool   `json:"upnp"`
	PublicURL   string `json:"publicUrl"`
	BrowserDrop bool   `json:"browserDrop"`
	DropPIN     string `json:"dropPin"`
	DropToken   string `json:"dropToken"`

	RendezvousURL string `json:"rendezvousUrl"`
	TransitRelay  string `json:"transitRelay"`

	// Favorites maps device fingerprints to their last known alias.
	Favorites map[string]string `json:"favorites"`
	// Manual lists addresses ("192.168.1.20:53317") added by hand.
	Manual []string `json:"manual"`
}

// Defaults returns the settings of a fresh install.
func Defaults(downloadDir string) Settings {
	return Settings{
		DeviceName:     defaultName(),
		Fingerprint:    randomHex(16),
		DownloadDir:    downloadDir,
		Discoverable:   true,
		Notifications:  true,
		MinimizeToTray: false,
		CloseToTray:    true,
		Port:           53317,
		WebPort:        53318,
		UPnP:           true,
		Favorites:      map[string]string{},
		DropToken:      randomHex(12),
	}
}

var adjectives = []string{"Swift", "Bright", "Calm", "Clever", "Cosmic", "Golden", "Lucky", "Nimble", "Quiet", "Brave", "Silver", "Sunny"}
var nouns = []string{"Falcon", "Otter", "Comet", "Maple", "Panda", "Orbit", "Fox", "Harbor", "Nova", "Lynx", "Pebble", "Willow"}

func defaultName() string {
	var b [2]byte
	_, _ = rand.Read(b[:])
	return adjectives[int(b[0])%len(adjectives)] + " " + nouns[int(b[1])%len(nouns)]
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

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
