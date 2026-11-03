package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/input"

	"easysend/internal/item"
)

// composeDialog writes a text note to send.
type composeDialog struct {
	u      *UI
	isOpen bool
	dlg    *dialog.Dialog
	text   *input.Input
	add    *giode.ButtonElement
	cancel *giode.ButtonElement
}

func newComposeDialog(u *UI) *composeDialog {
	d := &composeDialog{u: u}
	d.text = giode.Input.Text("Type a message, a link, a password…", giode.Styles{Background: colSurface2, BorderColor: "theme.border", BorderRadius: 12, Color: "theme.text"}).Lines(5, 12)
	d.add = primaryButton("Add", "plus", d.submit)
	d.cancel = softButton("Cancel", "", func() { d.isOpen = false })
	d.dlg = giode.Dialog(&d.isOpen, d.content, giode.Styles{Background: "theme.surface", BorderRadius: 26, Width: 520})
	return d
}

func (d *composeDialog) open() {
	d.isOpen = true
	d.text.Focus()
}

func (d *composeDialog) submit() {
	if d.text.Text() == "" {
		return
	}
	d.u.addItems([]item.Item{item.FromText(d.text.Text())})
	d.text.SetText("")
	d.isOpen = false
}

func (d *composeDialog) content() giode.Element {
	return giode.Stack(giode.Styles{Gap: 14, Width: -1},
		giode.Text("Send text", giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 22, Color: "theme.text"}),
		giode.Text("Receivers can copy it with one tap.", stMuted),
		d.text,
		giode.Row(giode.Styles{Gap: 10, Width: -1, Justify: giode.JustifyEnd}, d.cancel, d.add),
	)
}

func (d *composeDialog) Layout(gtx layout.Context) layout.Dimensions { return d.dlg.Layout(gtx) }
