package ui

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/iconbutton"
)

// primaryStyles are the styles of the gradient call-to-action buttons.
func (u *UI) primaryStyles() giode.Styles {
	s := u.eng.Settings.Get()
	return giode.Styles{
		BackgroundGradient: brandGradient(s, 120), Color: "#FFFFFF", BorderRadius: 12,
		FontSize: 14, Padding: giode.SymmetricInset(18, 11),
		BoxShadow: giode.BoxShadow{Y: 6, Blur: 18, Color: giode.Color(currentAccent(s).From + "55")},
	}
}

// primaryButton creates a gradient button; restyle keeps it in step with
// the accent color.
func (u *UI) primaryButton(label, iconName string, fn func()) *giode.ButtonElement {
	b := giode.Button(label, u.primaryStyles()).OnClick(fn)
	if iconName != "" {
		b.Icon(icon(iconName, 17, ""))
	}
	u.primaries = append(u.primaries, b)
	return b
}

// restyle applies the accent to the gradient buttons.
func (u *UI) restyle() {
	st := u.primaryStyles()
	for _, b := range u.primaries {
		b.Styles(st)
	}
}

// softButton is a quiet, tinted button.
func softButton(label, iconName string, fn func()) *giode.ButtonElement {
	b := giode.Button(label, giode.Styles{Background: colSurface3, Color: "theme.text", BorderRadius: 12, FontSize: 14, Padding: giode.SymmetricInset(16, 10)}).OnClick(fn)
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
