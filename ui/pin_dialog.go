package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/input"

	"easysend/internal/engine"
	"easysend/internal/item"
	"easysend/internal/transfer"
)

// pendingSend is a nearby send that may need a PIN.
type pendingSend struct {
	peer  engine.Peer
	items []item.Item
}

// pinDialog asks for the receiver's PIN when a send needs one.
type pinDialog struct {
	u       *UI
	open    bool
	dlg     *dialog.Dialog
	input   *input.Input
	send    *giode.ButtonElement
	cancel  *giode.ButtonElement
	sends   map[string]pendingSend
	current pendingSend
	shown   map[string]bool
}

func newPINDialog(u *UI) *pinDialog {
	d := &pinDialog{u: u, sends: map[string]pendingSend{}, shown: map[string]bool{}}
	d.input = giode.Input.Text("PIN", giode.Styles{Background: colSurface2, BorderColor: "theme.border", BorderRadius: 12, Color: "theme.text", FontSize: 20}).
		Keyboard(giode.KeyboardNumeric).Submit(func(string) { d.submit() })
	d.send = u.primaryButton("Send", "send", d.submit)
	d.cancel = softButton("Cancel", "", func() { d.open = false })
	d.dlg = giode.Dialog(&d.open, d.content, giode.Styles{Background: "theme.surface", BorderRadius: 26, Width: 380})
	return d
}

// watch remembers a send, to ask for a PIN if the receiver wants one.
func (d *pinDialog) watch(id string, p engine.Peer, items []item.Item) {
	d.sends[id] = pendingSend{peer: p, items: items}
}

// reopen asks for the PIN of a failed send again.
func (d *pinDialog) reopen(t transfer.Transfer) {
	if s, ok := d.sends[t.ID]; ok {
		d.current = s
		d.input.SetText("")
		d.open = true
		d.input.Focus()
	}
}

func (d *pinDialog) submit() {
	pin := d.input.Text()
	if pin == "" {
		return
	}
	id := d.u.eng.SendNearby(d.current.peer.Device, d.current.items, pin)
	d.watch(id, d.current.peer, d.current.items)
	d.open = false
	d.input.SetText("")
}

func (d *pinDialog) content() giode.Element {
	s := d.u.eng.Settings.Get()
	return giode.Stack(giode.Styles{Gap: 14, Width: -1, Align: giode.AlignCenter},
		orb(s, "lock", 60),
		giode.Text("PIN required", giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 22, Color: "theme.text"}),
		giode.Text(d.current.peer.Alias+" asks for a PIN. It's shown in their settings.", giode.Styles{FontFamily: fontText, FontSize: 13.5, Color: "theme.muted", TextAlign: giode.TextAlignCenter}),
		d.input,
		giode.Row(giode.Styles{Gap: 10, Width: -1, Justify: giode.JustifyEnd}, d.cancel, d.send),
	)
}

func (d *pinDialog) Layout(gtx layout.Context) layout.Dimensions {
	for id, s := range d.sends {
		t, ok := d.u.eng.Transfers.Get(id)
		if !ok || !t.Status.Finished() || d.shown[id] {
			continue
		}
		d.shown[id] = true
		if engine.IsPINError(t) && !d.open {
			d.current = s
			d.open = true
			d.input.Focus()
		}
	}
	return d.dlg.Layout(gtx)
}
