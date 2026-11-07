package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
)

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
