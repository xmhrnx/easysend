package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"

	"easysend/ui/platform"
)

// dropZone invites the user to drop or pick what to send.
type dropZone struct {
	u                          *UI
	files, folder, text, paste *giode.ButtonElement
}

func newDropZone(u *UI) *dropZone {
	return &dropZone{
		u:      u,
		files:  u.primaryButton("Files", "file", u.pickFiles),
		folder: softButton("Folder", "folder", u.pickFolder),
		text:   softButton("Text", "text", func() { u.compose.open() }),
		paste:  softButton("Paste", "clipboard", u.pasteText),
	}
}

func (d *dropZone) Layout(gtx layout.Context) layout.Dimensions {
	s := d.u.eng.Settings.Get()
	_, dragging := d.u.app.FileDrag()
	compact := len(d.u.items) > 0
	border, bg := "theme.border", colSurface2
	if dragging {
		border, bg = "theme.primary", colPrimarySoft
	}
	buttons := []giode.Element{d.files}
	if platform.IsDesktop {
		buttons = append(buttons, d.folder)
	}
	buttons = append(buttons, d.text, d.paste)
	actions := giode.Row(giode.Styles{Gap: 10, Justify: giode.JustifyCenter}, buttons...)

	title, sub := "Drop files to send", "Or pick files, a folder, a note or what's on your clipboard"
	if !platform.IsDesktop {
		title, sub = "What do you want to send?", "Pick files, write a note or paste from the clipboard"
	}
	if dragging {
		title = "Release to add"
	}
	if compact {
		return giode.Row(giode.Styles{Width: -1, Gap: 14, Align: giode.AlignCenter, Padding: giode.UniformInset(16), BorderRadius: 20,
			Background: giode.Color(bg), BorderWidth: 1, BorderColor: giode.Color(border)},
			orb(s, "plus", 40),
			giode.Stack(giode.Styles{Gap: 2, FlexGrow: 1},
				giode.Text("Add more", stStrong),
				giode.Text("Drop here or pick more to send together", stMuted)),
			actions,
		).Layout(gtx)
	}
	return giode.Stack(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 10, Padding: giode.Inset{Top: 34, Bottom: 30, Left: 20, Right: 20},
		BorderRadius: 26, Background: giode.Color(bg), BorderWidth: 2, BorderColor: giode.Color(border)},
		orb(s, "upload", 72),
		giode.Box(giode.Styles{Height: 6}),
		giode.Text(title, giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 22, Color: "theme.text", TextAlign: giode.TextAlignCenter}),
		giode.Text(sub, giode.Styles{FontFamily: fontText, FontSize: 14, Color: "theme.muted", TextAlign: giode.TextAlignCenter}),
		giode.Box(giode.Styles{Height: 8}),
		actions,
	).Layout(gtx)
}
