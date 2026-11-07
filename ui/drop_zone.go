package ui

import (
	"github.com/xmhrnx/giode"

	"easysend/ui/platform"
)

// dropAction is one way to add something to send.
type dropAction struct {
	label, icon string
	button      *giode.ButtonElement    // wide windows
	tile        *giode.PressableElement // phones
}

// dropZone invites the user to drop or pick what to send.
type dropZone struct {
	u       *UI
	actions []*dropAction
}

func newDropZone(u *UI) *dropZone {
	d := &dropZone{u: u}
	add := func(label, iconName string, primary bool, fn func()) {
		a := &dropAction{label: label, icon: iconName}
		if primary {
			a.button = primaryButton(label, iconName, fn)
		} else {
			a.button = softButton(label, iconName, fn)
		}
		a.tile = giode.Pressable(func(ps giode.PressState) giode.Element { return actionTile(a, primary, ps) }, giode.Styles{FlexGrow: 1}).
			Label(label).OnClick(func(giode.ClickEvent) { fn() })
		d.actions = append(d.actions, a)
	}
	add("Files", "file", true, u.pickFiles)
	if platform.IsDesktop {
		add("Folder", "folder", false, u.pickFolder)
	}
	add("Text", "text", false, func() { u.compose.open() })
	add("Paste", "clipboard", false, u.pasteText)
	return d
}

// actionTile is an action as a square-ish tile, the phone layout.
func actionTile(a *dropAction, primary bool, ps giode.PressState) giode.Element {
	fg, border, bg := "theme.text", "theme.border", "transparent"
	if primary {
		fg, border = "theme.primary", colRing
	}
	if ps.Pressed {
		bg = colPressed
	}
	return giode.Stack(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 6, Padding: giode.SymmetricInset(4, 14),
		BorderRadius: 16, BorderWidth: 1, BorderColor: giode.Color(border), Background: giode.Color(bg)},
		icon(a.icon, 22, fg),
		giode.Text(a.label, giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightMedium, FontSize: 13, Color: giode.Color(fg)}),
	)
}
