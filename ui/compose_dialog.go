package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/sheet"

	"easysend/internal/item"
)

// composeDialog writes a text note to send.
type composeDialog struct {
	u      *UI
	isOpen bool
	sheet  *sheet.Sheet
	text   *input.Input
	add    *giode.ButtonElement
	cancel *giode.ButtonElement
}

func newComposeDialog(u *UI) *composeDialog {
	d := &composeDialog{u: u}
	d.text = giode.Input.Text(tr("Type a message, a link, a password…"), giode.Styles{Background: colSurface2, BorderColor: "theme.border", BorderRadius: 12, Color: "theme.text"}).Lines(5, 12)
	d.add = sheetPrimary(tr("Add"), "plus", d.submit)
	d.cancel = sheetSecondary(tr("Cancel"), func() { d.isOpen = false })
	d.sheet = newSheet(&d.isOpen, tr("Send text"), d.content, func() giode.Element { return sheetActions(d.cancel, d.add) })
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
	return giode.Stack(giode.Styles{Gap: 12, Width: -1},
		giode.Text(tr("Receivers can copy it with one tap."), stMuted),
		d.text,
	)
}

func (d *composeDialog) Layout(gtx layout.Context) layout.Dimensions { return d.sheet.Layout(gtx) }
