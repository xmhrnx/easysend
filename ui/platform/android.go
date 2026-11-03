//go:build android

package platform

import (
	"os"
	"path/filepath"

	"easysend/internal/localsend"
)

// IsDesktop is true on Windows, macOS and Linux.
const IsDesktop = false

// DeviceType is the LocalSend device type.
const DeviceType = localsend.TypeMobile

// DownloadDir is the default folder for received files: the shared
// Download folder, where the Files app shows them, when the app may write
// there, else the app's own storage.
func DownloadDir() string {
	shared := "/storage/emulated/0/Download/EasySend"
	if err := os.MkdirAll(shared, 0o755); err == nil {
		probe := filepath.Join(shared, ".easysend")
		if f, err := os.Create(probe); err == nil {
			f.Close()
			os.Remove(probe)
			return shared
		}
	}
	return filepath.Join(DataDir(), "Received")
}
