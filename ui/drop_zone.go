package ui

import (
	"gioui.org/layout"
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

func (d *dropZone) Layout(gtx layout.Context) layout.Dimensions {
	if giode.Compact() {
		return d.compact().Layout(gtx)
	}
	return d.wide().Layout(gtx)
}

func (d *dropZone) compact() giode.Element {
	tiles := make([]giode.Element, len(d.actions))
	for i, a := range d.actions {
		tiles[i] = a.tile
	}
	title := "What do you want to send?"
	if len(d.u.items) > 0 {
		title = "Add more"
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 12},
		giode.Text(title, stHeading),
		giode.Row(giode.Styles{Width: -1, Gap: 10, Wrap: giode.WrapNoWrap}, tiles...),
	)
}

func (d *dropZone) wide() giode.Element {
	_, dragging := d.u.app.FileDrag()
	border, bg := "theme.border", colSurface2
	if dragging {
		border, bg = "theme.primary", colSurface3
	}
	buttons := make([]giode.Element, len(d.actions))
	for i, a := range d.actions {
		buttons[i] = a.button
	}
	actions := giode.Row(giode.Styles{Gap: 10, Justify: giode.JustifyCenter}, buttons...)
	if len(d.u.items) > 0 {
		return giode.Row(giode.Styles{Width: -1, Gap: 14, Align: giode.AlignCenter, Padding: giode.UniformInset(16), BorderRadius: 20,
			Background: giode.Color(bg), BorderWidth: 1, BorderColor: giode.Color(border)},
			orb("plus", 40),
			giode.Stack(giode.Styles{Gap: 2, FlexGrow: 1},
				giode.Text("Add more", stStrong),
				giode.Text("Drop here or pick more to send together", stMuted)),
			actions,
		)
	}
	title := "Drop files to send"
	if dragging {
		title = "Release to add"
	}
	return giode.Stack(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 10, Padding: giode.Inset{Top: 34, Bottom: 30, Left: 20, Right: 20},
		BorderRadius: 26, Background: giode.Color(bg), BorderWidth: 1, BorderColor: giode.Color(border)},
		orb("upload", 72),
		giode.Box(giode.Styles{Height: 6}),
		giode.Text(title, giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 22, Color: "theme.text", TextAlign: giode.TextAlignCenter}),
		giode.Text("Or pick files, a folder, a note or what's on your clipboard", giode.Styles{FontFamily: fontText, FontSize: 14, Color: "theme.muted", TextAlign: giode.TextAlignCenter}),
		giode.Box(giode.Styles{Height: 8}),
		actions,
	)
}
