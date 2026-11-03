//go:build android

package platform

import (
	"path/filepath"

	"easysend/internal/localsend"
)

// IsDesktop is true on Windows, macOS and Linux.
const IsDesktop = false

// DeviceType is the LocalSend device type.
const DeviceType = localsend.TypeMobile

// DownloadDir is the default folder for received files: the app's own
// storage. Received files are handed to other apps with the share sheet.
func DownloadDir() string {
	return filepath.Join(DataDir(), "Received")
}
