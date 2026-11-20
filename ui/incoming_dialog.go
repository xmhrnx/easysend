package ui

import (
	"context"
	"errors"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/checkbox"
	"github.com/xmhrnx/giode/components/sheet"

	"easysend/internal/engine"
	"easysend/internal/item"
	"easysend/internal/settings"
	"easysend/ui/platform"
)

// incomingDialog asks whether to accept a nearby transfer.
type incomingDialog struct {
	u       *UI
	open    bool
	req     engine.Request
	sheet   *sheet.Sheet
	accept  *giode.ButtonElement
	decline *giode.ButtonElement
	trust   *checkbox.Checkbox
	// change picks where an offered synced folder goes (desktops), and
	// root is that choice.
	change *giode.ButtonElement
	root   string
}

const requestFilesShown = 4

func newIncomingDialog(u *UI) *incomingDialog {
	d := &incomingDialog{u: u}
	d.accept = sheetPrimary("Accept", "check", func() { d.answer(true) })
	d.decline = sheetSecondary("Decline", func() { d.answer(false) })
	d.change = softButton("Change…", "folder", d.pickRoot)
	d.trust = giode.Input.Checkbox("Always accept from this device", giode.Styles{Color: "theme.text", Background: "theme.primary", BorderColor: "theme.border"}).Outlined(true)
	d.sheet = newSheet(&d.open, "", d.content, func() giode.Element { return sheetActions(d.decline, d.accept) }).Dismissible(false)
	return d
}

func (d *incomingDialog) answer(ok bool) {
	if d.req.SyncFolder != "" {
		d.u.eng.AnswerSync(d.req.ID, ok, d.root)
		d.open = false
		if ok {
			d.u.toast(giode.ToastSuccess, "Syncing “"+d.req.SyncFolder+"” with "+d.req.Sender.Alias)
		}
		return
	}
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

// pickRoot chooses where to keep an offered synced folder.
func (d *incomingDialog) pickRoot() {
	id := d.req.ID
	go func() {
		dir, err := d.u.app.PickFolder(context.Background(), giode.FolderOptions{Title: "Where should “" + d.req.SyncFolder + "” be kept?"})
		if errors.Is(err, giode.ErrCanceled) {
			return
		}
		d.u.app.Post(func() {
			if err != nil {
				d.u.toast(giode.ToastError, "Couldn't choose the folder: "+err.Error())
				return
			}
			if d.req.ID == id {
				d.root = dir
			}
		})
	}()
}

// syncContent is an offered synced folder: what it is and where it goes.
func (d *incomingDialog) syncContent() giode.Element {
	r := d.req
	where := []giode.Element{
		giode.Row(giode.Styles{Gap: 12, Align: giode.AlignCenter, Width: -1, Wrap: giode.WrapNoWrap},
			tile("folder", 40, "theme.primary"),
			giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
				giode.Text(r.SyncFolder, giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightSemiBold, FontSize: 15, Color: "theme.text", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
				giode.Text("Kept in "+d.u.displayPath(d.root), giode.Styles{FontFamily: fontText, FontSize: 12, Color: "theme.muted", MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis}))),
	}
	if platform.IsDesktop {
		where = append(where, d.change)
	}
	return giode.Stack(giode.Styles{Gap: 14, Width: -1, Align: giode.AlignCenter, Padding: giode.Inset{Top: 4}},
		orb("sync", 64),
		giode.Stack(giode.Styles{Gap: 4, Align: giode.AlignCenter},
			giode.Text(r.Sender.Alias, giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 21, Color: "theme.text", TextAlign: giode.TextAlignCenter}),
			giode.Text("wants to sync a folder with you", giode.Styles{FontFamily: fontText, FontSize: 13.5, Color: "theme.muted", TextAlign: giode.TextAlignCenter})),
		giode.Stack(giode.Styles{Gap: 12, Width: -1, Padding: giode.UniformInset(12), BorderRadius: 14, BorderWidth: 1, BorderColor: "theme.border"}, where...),
		giode.Text("Changes on either device reach the other while both are on the same network. You can stop anytime on the Sync screen.", giode.Styles{FontFamily: fontText, FontSize: 12, Color: "theme.muted", TextAlign: giode.TextAlignCenter}),
	)
}

func (d *incomingDialog) content() giode.Element {
	r := d.req
	if r.SyncFolder != "" {
		return d.syncContent()
	}
	files := []giode.Element{}
	for i, f := range r.Files {
		if i == requestFilesShown {
			files = append(files, giode.Text("+ "+plural(len(r.Files)-requestFilesShown, "more file", "more files"), stSmall))
			break
		}
		files = append(files, d.fileRow(f.FileName, f.Size))
	}
	what := "wants to send you " + plural(len(r.Files), "file", "files") + " · " + item.HumanSize(r.Total)
	if r.Text != "" {
		// A message: show what it says, as files show their names.
		what = "wants to send you a message"
		files = []giode.Element{giode.Text(r.Text, giode.Styles{FontFamily: fontText, FontSize: 14, Color: "theme.text", MaxLines: 6, TextOverflow: giode.TextOverflowEllipsis})}
	}
	return giode.Stack(giode.Styles{Gap: 14, Width: -1, Align: giode.AlignCenter, Padding: giode.Inset{Top: 4}},
		orb(deviceIcon(r.Sender.DeviceType), 64),
		giode.Stack(giode.Styles{Gap: 4, Align: giode.AlignCenter},
			giode.Text(r.Sender.Alias, giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 21, Color: "theme.text", TextAlign: giode.TextAlignCenter}),
			giode.Text(what, giode.Styles{FontFamily: fontText, FontSize: 13.5, Color: "theme.muted", TextAlign: giode.TextAlignCenter})),
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
		d.root = d.req.SyncRoot
		d.open = true
	}
	return d.sheet.Layout(gtx)
}
