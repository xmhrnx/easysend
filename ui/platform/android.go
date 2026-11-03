//go:build android

package platform

import (
	"path/filepath"

	"easysend/internal/localsend"
)

// IsDesktop is true on Windows, macOS and Linux.
const IsDesktop = false

// ReceivesToDownloads says received files are published to the system
// Downloads (Engine Config.Publish) rather than kept where they arrive.
const ReceivesToDownloads = true

// DeviceType is the LocalSend device type.
const DeviceType = localsend.TypeMobile

// DownloadDir is where received files land first, in the app's own
// storage; once a transfer completes they move to Download/EasySend.
func DownloadDir() string {
	return filepath.Join(DataDir(), "Received")
}
