package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"
	"github.com/xmhrnx/giode/components/sheet"

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
	sheet   *sheet.Sheet
	accept  *giode.ButtonElement
	decline *giode.ButtonElement
	trust   *checkbox.Checkbox
}

func newIncomingDialog(u *UI) *incomingDialog {
	d := &incomingDialog{u: u}
	d.accept = sheetPrimary("Accept", "check", func() { d.answer(true) })
	d.decline = sheetSecondary("Decline", func() { d.answer(false) })
	d.trust = giode.Input.Checkbox("Always accept from this device", giode.Styles{Color: "theme.text", Background: "theme.primary", BorderColor: "theme.border"}).Outlined(true)
	d.sheet = newSheet(&d.open, "", d.content, func() giode.Element { return sheetActions(d.decline, d.accept) }).Dismissible(false)
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

func (d *incomingDialog) fileRow(f string, size int64) giode.Element {
	return giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Width: -1, Wrap: giode.WrapNoWrap},
		tile(fileIcon(f), 32, "theme.primary"),
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(f, giode.Styles{FontFamily: fontText, FontSize: 13.5, Color: "theme.text", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis})),
		giode.Text(item.HumanSize(size), stSmall))
}

func (d *incomingDialog) content() giode.Element {
	r := d.req
	files := []giode.Element{}
	for i, f := range r.Files {
		if i == requestFilesShown {
			files = append(files, giode.Text("+ "+plural(len(r.Files)-requestFilesShown, "more file", "more files"), stSmall))
			break
		}
		files = append(files, d.fileRow(f.FileName, f.Size))
	}
	return giode.Stack(giode.Styles{Gap: 14, Width: -1, Align: giode.AlignCenter, Padding: giode.Inset{Top: 4}},
		orb(deviceIcon(r.Sender.DeviceType), 64),
		giode.Stack(giode.Styles{Gap: 4, Align: giode.AlignCenter},
			giode.Text(r.Sender.Alias, giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 21, Color: "theme.text", TextAlign: giode.TextAlignCenter}),
			giode.Text("wants to send you "+plural(len(r.Files), "file", "files")+" · "+item.HumanSize(r.Total), giode.Styles{FontFamily: fontText, FontSize: 13.5, Color: "theme.muted", TextAlign: giode.TextAlignCenter})),
		giode.Stack(giode.Styles{Gap: 10, Width: -1, Padding: giode.UniformInset(12), BorderRadius: 14, BorderWidth: 1, BorderColor: "theme.border"}, files...),
		giode.Box(giode.Styles{Width: -1}, d.trust),
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
	return d.sheet.Layout(gtx)
}
