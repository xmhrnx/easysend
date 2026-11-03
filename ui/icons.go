package ui

import (
	"embed"

	"github.com/xmhrnx/giode"
	prop "github.com/xmhrnx/giode/styles/properties"
)

//go:embed icons
var iconFS embed.FS

// LoadIcons registers the embedded icon set.
func LoadIcons() error { return giode.RegisterIconsFS(iconFS, "icons") }

// icon renders a registered icon at size dp in color ("" inherits).
func icon(name string, size int, color string) giode.Element {
	return giode.Icon(name, giode.Styles{Width: prop.Width(size), Height: prop.Height(size), Color: giode.Color(color)})
}
