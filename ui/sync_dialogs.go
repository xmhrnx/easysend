package ui

import (
	"context"
	"errors"
	"strings"
	"time"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/sheet"

	"easysend/internal/engine"
)

// newFolderDialog names a new synced folder, on phones: it's made in
// EasySend's folder.
type newFolderDialog struct {
	u      *UI
	isOpen bool
	sheet  *sheet.Sheet
	name   *input.Input
	create *giode.ButtonElement
	cancel *giode.ButtonElement
}

func newNewFolderDialog(u *UI) *newFolderDialog {
	d := &newFolderDialog{u: u}
	d.name = giode.Input.Text("Photos, Work, Notes…", giode.Styles{Background: colSurface2, BorderColor: "theme.border", BorderRadius: 12, Color: "theme.text"}).
		Submit(func(string) { d.submit() })
	d.create = sheetPrimary("Create", "plus", d.submit)
	d.cancel = sheetSecondary("Cancel", func() { d.isOpen = false })
	d.sheet = newSheet(&d.isOpen, "New synced folder", d.content, func() giode.Element { return sheetActions(d.cancel, d.create) })
	return d
}

func (d *newFolderDialog) open() {
	d.name.SetText("")
	d.isOpen = true
	d.name.Focus()
}

func (d *newFolderDialog) submit() {
	name := strings.TrimSpace(d.name.Text())
	if name == "" {
		return
	}
	d.isOpen = false
	d.u.startSyncing(name, d.u.eng.SyncFolderDir(name))
}

func (d *newFolderDialog) content() giode.Element {
	return giode.Stack(giode.Styles{Gap: 14, Width: -1},
		giode.Text("A new folder in EasySend's folder, kept in sync with the devices you choose next. Add files to it from the Files app or by receiving them.", stMuted),
		d.name,
	)
}

func (d *newFolderDialog) Layout(gtx layout.Context) layout.Dimensions { return d.sheet.Layout(gtx) }

// syncDevicesDialog shares a folder with nearby devices.
type syncDevicesDialog struct {
	u       *UI
	isOpen  bool
	f       engine.SyncFolderInfo
	sheet   *sheet.Sheet
	done    *giode.ButtonElement
	refresh *giode.ButtonElement
	// waiting holds the devices asked, until they answer.
	waiting map[string]bool
}

func newSyncDevicesDialog(u *UI) *syncDevicesDialog {
	d := &syncDevicesDialog{u: u, waiting: map[string]bool{}}
	d.done = sheetSecondary("Done", func() { d.isOpen = false })
	d.refresh = softButton("Look again", "refresh", u.eng.Refresh)
	d.sheet = newSheet(&d.isOpen, "", d.content, func() giode.Element { return sheetActions(d.done) })
	return d
}

func (d *syncDevicesDialog) open(f engine.SyncFolderInfo) {
	d.f = f
	d.isOpen = true
	d.u.Go(screenSync)
}

// shared is whether the folder already syncs with a device.
func (d *syncDevicesDialog) shared(fp string) bool {
	for _, f := range d.u.eng.SyncFolders() {
		if f.ID == d.f.ID {
			for _, p := range f.Peers {
				if p.FP == fp {
					return true
				}
			}
		}
	}
	return false
}

func (d *syncDevicesDialog) invite(p engine.Peer) {
	if d.waiting[p.Fingerprint] || d.shared(p.Fingerprint) {
		return
	}
	d.waiting[p.Fingerprint] = true
	f := d.f
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		err := d.u.eng.InviteToSync(ctx, f.ID, p)
		d.u.app.Post(func() {
			delete(d.waiting, p.Fingerprint)
			switch {
			case err == nil:
				d.u.toast(giode.ToastSuccess, "Syncing “"+f.Name+"” with "+p.Alias)
			case errors.Is(err, engine.ErrSyncDeclined):
				d.u.toast(giode.ToastWarning, p.Alias+" declined")
			case errors.Is(err, engine.ErrSyncNotSupported):
				d.u.toast(giode.ToastError, p.Alias+" can't sync folders: it needs the latest EasySend")
			default:
				d.u.toast(giode.ToastError, "Couldn't reach "+p.Alias+": "+err.Error())
			}
		})
	}()
}

func (d *syncDevicesDialog) row(p engine.Peer) giode.Element {
	press := giode.Keyed("syncinvite", p.Fingerprint, func() *giode.PressableElement {
		return giode.Pressable(func(giode.PressState) giode.Element { return giode.Box(giode.Styles{}) })
	})
	shared, waiting := d.shared(p.Fingerprint), d.waiting[p.Fingerprint]
	press.Label("Share with " + p.Alias).OnClick(func(giode.ClickEvent) { d.invite(p) })
	press.Content(func(ps giode.PressState) giode.Element {
		border, bg := "theme.border", "transparent"
		switch {
		case shared || waiting:
		case ps.Pressed:
			bg = colPressed
		case ps.Hovered:
			border = "theme.primary"
		}
		var trail giode.Element = icon("plus", 18, "theme.primary")
		switch {
		case shared:
			trail = chip("Syncing", "theme.success")
		case waiting:
			trail = chip("Waiting for an answer…", "theme.muted")
		}
		sub := p.Model()
		if sub == "" {
			sub = p.IP
		}
		return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(12, 10),
			BorderRadius: 14, BorderWidth: 1, BorderColor: giode.Color(border), Background: giode.Color(bg)},
			orb(deviceIcon(p.DeviceType), 38),
			giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
				giode.Text(p.Alias, giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightMedium, FontSize: 14, Color: "theme.text", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
				giode.Text(sub, giode.Styles{FontFamily: fontText, FontSize: 12, Color: "theme.muted", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis})),
			trail)
	})
	return press
}

func (d *syncDevicesDialog) content() giode.Element {
	kids := []giode.Element{
		giode.Stack(giode.Styles{Gap: 4, Width: -1},
			giode.Text("Share “"+d.f.Name+"”", giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 21, Color: "theme.text"}),
			giode.Text("Choose the devices to keep it in sync with. Each one is asked to accept, and chooses where to keep it.", stMuted)),
	}
	peers := d.u.eng.Peers()
	for _, p := range peers {
		kids = append(kids, d.row(p))
	}
	if len(peers) == 0 {
		kids = append(kids, card(giode.Styles{Align: giode.AlignCenter, Gap: 10},
			giode.Text("No devices nearby", stStrong),
			giode.Text("Open EasySend on the other device. It shows up here when it's on the same network.", giode.Styles{FontFamily: fontText, FontSize: 13, Color: "theme.muted", TextAlign: giode.TextAlignCenter}),
			d.refresh))
	}
	return giode.Stack(giode.Styles{Gap: 12, Width: -1}, kids...)
}

func (d *syncDevicesDialog) Layout(gtx layout.Context) layout.Dimensions { return d.sheet.Layout(gtx) }

// confirmDialog asks before doing something that can't be undone.
type confirmDialog struct {
	isOpen      bool
	title, body string
	sheet       *sheet.Sheet
	yes, no     *giode.ButtonElement
	then        func()
}

func newConfirmDialog() *confirmDialog {
	d := &confirmDialog{}
	d.yes = wide(dangerButton("", "", func() {
		d.isOpen = false
		if d.then != nil {
			d.then()
		}
	}), giode.Styles{BorderColor: "theme.danger", Color: "theme.danger", BorderRadius: 12, FontSize: 14})
	d.no = sheetSecondary("Cancel", func() { d.isOpen = false })
	d.sheet = newSheet(&d.isOpen, "", d.content, func() giode.Element { return sheetActions(d.no, d.yes) })
	return d
}

// ask shows the question; then runs when the user confirms with action.
func (d *confirmDialog) ask(title, body, action string, then func()) {
	d.title, d.body, d.then = title, body, then
	d.yes.Label(action)
	d.isOpen = true
}

func (d *confirmDialog) content() giode.Element {
	return giode.Stack(giode.Styles{Gap: 8, Width: -1},
		giode.Text(d.title, giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 20, Color: "theme.text"}),
		giode.Text(d.body, stMuted))
}

func (d *confirmDialog) Layout(gtx layout.Context) layout.Dimensions { return d.sheet.Layout(gtx) }
