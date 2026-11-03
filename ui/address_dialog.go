package ui

import (
	"context"
	"time"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/input"
)

// addressDialog adds a device by its IP address.
type addressDialog struct {
	u      *UI
	isOpen bool
	dlg    *dialog.Dialog
	addr   *input.Input
	add    *giode.ButtonElement
	cancel *giode.ButtonElement
	err    string
}

func newAddressDialog(u *UI) *addressDialog {
	d := &addressDialog{u: u}
	d.addr = giode.Input.Text("192.168.1.20 or 192.168.1.20:53317", giode.Styles{Background: colSurface2, BorderColor: "theme.border", BorderRadius: 12, Color: "theme.text"}).
		Keyboard(giode.KeyboardURL).Submit(func(string) { d.submit() })
	d.add = u.primaryButton("Connect", "plus", d.submit)
	d.cancel = softButton("Cancel", "", func() { d.isOpen = false })
	d.dlg = giode.Dialog(&d.isOpen, d.content, giode.Styles{Background: "theme.surface", BorderRadius: 26, Width: 440})
	return d
}

func (d *addressDialog) open() {
	d.err = ""
	d.isOpen = true
	d.addr.Focus()
}

func (d *addressDialog) submit() {
	addr := d.addr.Text()
	if addr == "" {
		return
	}
	d.add.Loading(true)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		dev, err := d.u.eng.AddByAddress(ctx, addr)
		d.u.app.Post(func() {
			d.add.Loading(false)
			if err != nil {
				d.err = "No EasySend or LocalSend device answered at " + addr + "."
				return
			}
			d.isOpen = false
			d.addr.SetText("")
			d.u.toast(giode.ToastSuccess, "Added "+dev.Alias)
		})
	}()
}

func (d *addressDialog) content() giode.Element {
	kids := []giode.Element{
		giode.Text("Add a device", giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 22, Color: "theme.text"}),
		giode.Text("For networks that block discovery. The address is shown on the other device's Receive screen.", stMuted),
		d.addr,
	}
	if d.err != "" {
		kids = append(kids, giode.Text(d.err, giode.Styles{FontFamily: fontText, FontSize: 13, Color: "theme.danger"}))
	}
	kids = append(kids, giode.Row(giode.Styles{Gap: 10, Width: -1, Justify: giode.JustifyEnd}, d.cancel, d.add))
	return giode.Stack(giode.Styles{Gap: 14, Width: -1}, kids...)
}

func (d *addressDialog) Layout(gtx layout.Context) layout.Dimensions { return d.dlg.Layout(gtx) }
