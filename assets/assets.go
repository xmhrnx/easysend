// Package assets embeds the app's images and fonts.
package assets

import _ "embed"

// Icon is the app icon (1024px PNG).
//
//go:embed icon.png
var Icon []byte

// Logo is the app icon at 192px, for the UI.
//
//go:embed logo.png
var Logo []byte

// Tray is the tray icon in color (Windows, Linux).
//
//go:embed tray.png
var Tray []byte

// TrayTemplate is the black tray icon macOS tints for the menu bar.
//
//go:embed tray-template.png
var TrayTemplate []byte

// Fonts are the Inter faces the UI uses.
var (
	//go:embed fonts/Inter-Regular.ttf
	InterRegular []byte
	//go:embed fonts/Inter-Medium.ttf
	InterMedium []byte
	//go:embed fonts/Inter-SemiBold.ttf
	InterSemiBold []byte
	//go:embed fonts/Inter-Bold.ttf
	InterBold []byte
	//go:embed fonts/InterDisplay-Bold.ttf
	InterDisplayBold []byte
)
