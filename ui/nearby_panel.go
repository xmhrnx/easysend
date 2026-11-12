package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/iconbutton"

	"easysend/internal/engine"
	"easysend/internal/item"
)

// nearbyPanel lists the devices on the local network.
type nearbyPanel struct {
	u       *UI
	refresh *iconbutton.Button
	addIP   *giode.ButtonElement
}

func newNearbyPanel(u *UI) *nearbyPanel {
	return &nearbyPanel{
		u:       u,
		refresh: iconAction("refresh", tr("Look again"), u.eng.Refresh),
		addIP:   softButton(tr("Add by IP"), "plus", func() { u.address.open() }),
	}
}

func (n *nearbyPanel) Layout(gtx layout.Context) layout.Dimensions {
	peers := n.u.eng.Peers()
	trailing := []giode.Element{n.addIP}
	if n.u.eng.Scanning() {
		trailing = append(trailing, giode.Spinner(giode.Styles{Width: 20, Height: 20, Color: "theme.primary"}))
	} else {
		trailing = append(trailing, n.refresh)
	}
	head := sectionTitle(tr("Nearby devices"), trailing...)
	if len(peers) == 0 {
		return card(giode.Styles{Align: giode.AlignCenter, Gap: 6},
			head,
			radar(compactOr(150, 200), "wifi"),
			giode.Text(tr("Looking for devices on your network…"), stStrong),
			giode.Text(tr("Open EasySend or LocalSend on the other device and join the same Wi-Fi."), giode.Styles{FontFamily: fontText, FontSize: 13, Color: "theme.muted", TextAlign: giode.TextAlignCenter}),
		).Layout(gtx)
	}
	cards := make([]giode.Element, len(peers))
	for i, p := range peers {
		cards[i] = n.u.deviceCardFor(p)
	}
	hint := tr("Tap a device to send")
	if len(n.u.items) == 0 {
		hint = tr("Choose something to send, then tap a device")
	}
	return card(giode.Styles{},
		head,
		giode.Text(hint, stMuted),
		giode.Row(giode.Styles{Gap: 14, Width: -1}, cards...),
	).Layout(gtx)
}

// sendTo sends the selection to a nearby device.
func (u *UI) sendTo(p engine.Peer) {
	if len(u.items) == 0 {
		u.toast(giode.ToastInfo, tr("Choose something to send first"))
		return
	}
	items := append([]item.Item(nil), u.items...)
	id := u.eng.SendNearby(p.Device, items, "")
	u.pin.watch(id, p, items)
	u.toast(giode.ToastInfo, tr("Waiting for {name} to accept…", "name", p.Alias))
}
