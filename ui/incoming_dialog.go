package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"
	"github.com/xmhrnx/giode/components/dialog"

	"easysend/internal/engine"
	"easysend/internal/item"
	"easysend/internal/settings"
)

const requestFilesShown = 4

// incomingDialog asks whether to accept a nearby transfer.
type incomingDialog struct {
	u       *UI
	open    bool
	req     engine.Request
	dlg     *dialog.Dialog
	accept  *giode.ButtonElement
	decline *giode.ButtonElement
	trust   *checkbox.Checkbox
}

func newIncomingDialog(u *UI) *incomingDialog {
	d := &incomingDialog{u: u}
	d.accept = primaryButton("Accept", "check", func() { d.answer(true) })
	d.decline = softButton("Decline", "", func() { d.answer(false) })
	d.trust = giode.Input.Checkbox("Always accept from this device", giode.Styles{Color: "theme.text", Background: "theme.primary", BorderColor: "theme.border"}).Outlined(true)
	d.dlg = giode.Dialog(&d.open, d.content, giode.Styles{Background: "theme.surface", BorderRadius: 26, Width: 440}).Dismissible(false)
	return d
}

func (d *incomingDialog) answer(ok bool) {
	if ok && d.trust.Checked() {
		d.u.eng.SetFavorite(d.req.Sender, true)
		if d.u.eng.Settings.Get().AutoAccept == settings.AcceptAsk {
			_ = d.u.eng.UpdateSettings(func(s *settings.Settings) { s.AutoAccept = settings.AcceptFavorites })
		}
	}
	d.u.eng.Answer(d.req.ID, ok)
	d.trust.SetChecked(false)
	d.open = false
	if ok {
		d.u.toast(giode.ToastSuccess, "Receiving from "+d.req.Sender.Alias)
	}
}

func (d *incomingDialog) content() giode.Element {
	r := d.req
	files := []giode.Element{}
	for i, f := range r.Files {
		if i == requestFilesShown {
			files = append(files, giode.Text("+ "+plural(len(r.Files)-requestFilesShown, "more file", "more files"), stSmall))
			break
		}
		files = append(files, giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Width: -1, Wrap: giode.WrapNoWrap},
			tile(fileIcon(f.FileName), 32, "theme.accent"),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(f.FileName, giode.Styles{FontFamily: fontText, FontSize: 13.5, Color: "theme.text", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis})),
			giode.Text(item.HumanSize(f.Size), stSmall)))
	}
	return giode.Stack(giode.Styles{Gap: 16, Width: -1, Align: giode.AlignCenter},
		orb(deviceIcon(r.Sender.DeviceType), 72),
		giode.Stack(giode.Styles{Gap: 4, Align: giode.AlignCenter},
			giode.Text(r.Sender.Alias, giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 22, Color: "theme.text"}),
			giode.Text("wants to send you "+plural(len(r.Files), "file", "files")+" · "+item.HumanSize(r.Total), stMuted)),
		giode.Stack(giode.Styles{Gap: 8, Width: -1, Padding: giode.UniformInset(12), Background: colSurface2, BorderRadius: 14}, files...),
		giode.Box(giode.Styles{Width: -1}, d.trust),
		giode.Row(giode.Styles{Gap: 10, Width: -1, Justify: giode.JustifyEnd}, d.decline, d.accept),
	)
}

func (d *incomingDialog) Layout(gtx layout.Context) layout.Dimensions {
	reqs := d.u.eng.Requests()
	switch {
	case len(reqs) == 0:
		d.open = false
	case !d.open || d.req.ID != reqs[0].ID:
		d.req = reqs[0]
		d.open = true
	}
	return d.dlg.Layout(gtx)
}
