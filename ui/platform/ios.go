//go:build ios

package platform

import (
	"os"
	"path/filepath"

	"easysend/internal/localsend"
)

// IsDesktop is true on Windows, macOS and Linux.
const IsDesktop = false

// ReceivesToDownloads says received files are published to the system
// Downloads (Engine Config.Publish) rather than kept where they arrive.
const ReceivesToDownloads = false

// DeviceType is the LocalSend device type.
const DeviceType = localsend.TypeMobile

// DownloadDir is the default folder for received files: Documents, which
// the Files app shows (UIFileSharingEnabled).
func DownloadDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(DataDir(), "Received")
	}
	return filepath.Join(home, "Documents", "EasySend")
}
