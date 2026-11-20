package ui

import (
	"context"
	"os"
	"strconv"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/iconbutton"

	"easysend/internal/engine"
	"easysend/internal/item"
	"easysend/ui/platform"
)

// syncCard is a synced folder: its status, its devices and actions.
type syncCard struct {
	u      *UI
	f      engine.SyncedFolder
	invite *giode.ButtonElement
	open   *giode.ButtonElement
	now    *iconbutton.Button
	stop   *giode.ButtonElement
}

func newSyncCard(u *UI) *syncCard {
	c := &syncCard{u: u}
	c.invite = primaryButton("Add device", "plus", func() { u.syncDevices.open(c.f.Folder) })
	c.open = softButton("Open", "external", c.openFolder)
	c.now = iconAction("refresh", "Sync now", func() {
		u.eng.SyncNow()
		u.toast(giode.ToastInfo, "Syncing…")
	})
	c.stop = dangerButton("Stop syncing", "x", func() {
		f := c.f
		u.confirm.ask("Stop syncing “"+f.Name+"”?",
			"It stops syncing with every device. The files stay on each device, as they are now.",
			"Stop syncing", func() {
				u.eng.RemoveSyncFolder(f.ID)
				u.toast(giode.ToastSuccess, "Stopped syncing "+f.Name)
			})
	})
	return c
}

// syncCardFor returns the card of a folder, keeping its state across
// frames.
func (u *UI) syncCardFor(f engine.SyncedFolder) *syncCard {
	c := giode.Keyed("syncfolder", f.ID, func() *syncCard { return newSyncCard(u) })
	c.f = f
	return c
}

func (c *syncCard) openFolder() {
	root := c.f.Root
	go func() {
		err := os.MkdirAll(root, 0o755)
		if err == nil {
			err = c.u.app.OpenFolder(context.Background(), root)
		}
		if err != nil {
			c.u.app.Post(func() { c.u.toast(giode.ToastError, "Couldn't open the folder: "+err.Error()) })
		}
	}()
}

// status is the folder's state in a few words, and its color.
func (c *syncCard) status() (string, string) {
	st := c.f.Status
	switch {
	case st.Err != "":
		return st.Err, "theme.danger"
	case st.Syncing && st.Total > 0:
		return "Syncing " + strconv.Itoa(st.Done) + " of " + plural(st.Total, "file", "files"), "theme.primary"
	case st.Syncing:
		return "Checking for changes…", "theme.muted"
	case len(c.f.Peers) == 0:
		return "Not shared with any device yet", "theme.muted"
	}
	for _, p := range c.f.Peers {
		if ps := st.Peers[p.FP]; ps.Err != "" {
			return "Couldn't sync everything with " + c.u.syncDeviceName(p), "theme.danger"
		}
	}
	if st.LastSync.IsZero() {
		return "Waiting to sync", "theme.muted"
	}
	return "Up to date · checked " + timeAgo(st.LastSync), "theme.success"
}

func (c *syncCard) peerRow(p engine.SyncPeer) giode.Element {
	f := c.f
	name := c.u.syncDeviceName(p)
	ps := f.Status.Peers[p.FP]
	dot, sub := "theme.muted", "Offline — syncs when it's back"
	switch {
	case ps.Err != "":
		dot, sub = "theme.danger", ps.Err
	case ps.Online && !ps.LastSync.IsZero():
		dot, sub = "theme.success", "Synced "+timeAgo(ps.LastSync)
	case ps.Online:
		dot, sub = "theme.success", "Online"
	case f.Status.LastSync.IsZero():
		sub = "Looking for it…"
	}
	remove := giode.Keyed("syncpeer", f.ID+p.FP, func() *iconbutton.Button {
		return iconAction("x", "Stop syncing with this device", nil)
	})
	remove.OnClick(func() {
		c.u.confirm.ask("Stop syncing with "+name+"?",
			"“"+f.Name+"” stops syncing with "+name+". The files stay on both devices.",
			"Stop", func() {
				c.u.eng.StopSyncWith(f.ID, p.FP)
				c.u.toast(giode.ToastSuccess, "Stopped syncing with "+name)
			})
	})
	return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{Width: 8, Height: 8, BorderRadius: 4, Background: giode.Color(dot)}),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
			giode.Text(name, giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightMedium, FontSize: 14, Color: "theme.text", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
			giode.Text(sub, giode.Styles{FontFamily: fontText, FontSize: 12, Color: "theme.muted", MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis}),
		),
		remove,
	)
}

func (c *syncCard) content() giode.Element {
	f := c.f
	status, color := c.status()
	size := plural(f.Status.Files, "file", "files") + " · " + item.HumanSize(f.Status.Bytes)
	head := giode.Row(giode.Styles{Width: -1, Gap: 14, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		tile("folder", 44, "theme.primary"),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
			giode.Text(f.Name, giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightSemiBold, FontSize: 16, Color: "theme.text", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
			giode.Text(c.u.displayPath(f.Root), giode.Styles{FontFamily: fontText, FontSize: 12, Color: "theme.muted", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
		),
		c.now,
	)
	statusRow := giode.Row(giode.Styles{Width: -1, Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(status, giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightMedium, FontSize: 13, Color: giode.Color(color), MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis})),
		giode.Text(size, stSmall),
	)
	var peers []giode.Element
	for _, p := range f.Peers {
		peers = append(peers, c.peerRow(p))
	}
	if len(peers) == 0 {
		peers = append(peers, giode.Text("Add the devices this folder should be on. Each one is asked first.", stMuted))
	}
	devices := giode.Stack(giode.Styles{Width: -1, Gap: 12, Padding: giode.UniformInset(14), BorderRadius: 14, BorderWidth: 1, BorderColor: "theme.border"},
		append([]giode.Element{giode.Text("Devices", stOverline)}, peers...)...)
	actions := []giode.Element{c.invite}
	if platform.CanOpenFolders {
		actions = append(actions, c.open)
	}
	actions = append(actions, spacer(), c.stop)
	return card(giode.Styles{},
		head,
		statusRow,
		devices,
		sideBySide(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter}, actions...),
	)
}

func (c *syncCard) Layout(gtx layout.Context) layout.Dimensions { return c.content().Layout(gtx) }
