// Package settings stores EasySend's preferences as JSON.
package settings

import (
	"crypto/rand"
	"encoding/hex"
)

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
