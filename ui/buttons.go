package ui

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/iconbutton"
	"github.com/xmhrnx/giode/components/toggle"
)

// primaryStyles are the call-to-action buttons: an accent outline and
// accent text on the surface; hover and press tint it with a neutral
// shade, never the accent.
var primaryStyles = giode.Styles{
	BorderColor: "theme.primary", Background: "theme.text", Color: "theme.primary", BorderWidth: 1,
	FontWeight: giode.FontWeightSemiBold, BorderRadius: 12, FontSize: 14, Padding: giode.SymmetricInset(18, 10),
}

// primaryButton creates an outlined accent button.
func primaryButton(label, iconName string, fn func()) *giode.ButtonElement {
	b := giode.Button(label, primaryStyles).Variant(giode.VariantOutlined).OnClick(fn)
	if iconName != "" {
		b.Icon(icon(iconName, 17, ""))
	}
	return b
}

// softButton is a quiet button with a neutral outline.
func softButton(label, iconName string, fn func()) *giode.ButtonElement {
	b := giode.Button(label, giode.Styles{BorderColor: "theme.border", Background: "theme.text", Color: "theme.text", BorderWidth: 1,
		BorderRadius: 12, FontSize: 14, Padding: giode.SymmetricInset(16, 10)}).
		Variant(giode.VariantOutlined).OnClick(fn)
	if iconName != "" {
		b.Icon(icon(iconName, 17, ""))
	}
	return b
}

// dangerButton is an outlined button for destructive actions.
func dangerButton(label, iconName string, fn func()) *giode.ButtonElement {
	b := giode.Button(label, giode.Styles{BorderColor: "theme.danger", Color: "theme.danger", BorderRadius: 12, FontSize: 14, Padding: giode.SymmetricInset(14, 9)}).
		Variant(giode.VariantOutlined).OnClick(fn)
	if iconName != "" {
		b.Icon(icon(iconName, 16, ""))
	}
	return b
}

// iconAction is a round ghost icon button.
func iconAction(name, label string, fn func()) *iconbutton.Button {
	return giode.IconButton(icon(name, 18, "theme.muted"), giode.Styles{Color: "theme.text", BorderRadius: 10}).
		Variant(giode.IconButtonGhost).Label(label).OnClick(fn)
}

// newSwitch is a switch that stays dark: a neutral track, and a knob that
// turns accent when on.
func newSwitch(label string, st giode.Styles) *toggle.Switch {
	st.Background = colSurface3
	st.BorderColor = colSurface3
	return giode.Input.Switch(label, st).KnobColor("theme.muted", "theme.primary")
}
