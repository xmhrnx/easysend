//go:build !android && !ios

package platform

import (
	"os"
	"path/filepath"

	"easysend/internal/localsend"
)

// IsDesktop is true on Windows, macOS and Linux.
const IsDesktop = true

// DeviceType is the LocalSend device type.
const DeviceType = localsend.TypeDesktop

// DownloadDir is the default folder for received files.
func DownloadDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(DataDir(), "Received")
	}
	return filepath.Join(home, "Downloads", "EasySend")
}
