// Package platform holds what differs between desktops and phones.
package platform

import (
	"os"
	"path/filepath"

	gioapp "gioui.org/app"
)

// Profile is the EASYSEND_PROFILE environment variable: a name for a
// separate set of settings and history, so several copies of EasySend can
// run side by side. Empty for the default profile.
func Profile() string { return os.Getenv("EASYSEND_PROFILE") }

// DataDir is where EasySend keeps its settings and history.
func DataDir() string {
	dir, err := gioapp.DataDir()
	if err != nil || dir == "" {
		dir = os.TempDir()
	}
	name := "EasySend"
	if p := Profile(); p != "" {
		name += "-" + p
	}
	return filepath.Join(dir, name)
}
