package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/iconbutton"

	"easysend/internal/engine"
)

// deviceCard is a nearby device; clicking it sends the selection there.
type deviceCard struct {
	u     *UI
	peer  engine.Peer
	press *giode.PressableElement
	star  *iconbutton.Button
}

func newDeviceCard(u *UI) *deviceCard {
	c := &deviceCard{u: u}
	c.press = giode.Pressable(c.content).OnClick(func(giode.ClickEvent) { u.sendTo(c.peer) })
	c.star = giode.IconButton(icon("star", 18, "theme.muted"), giode.Styles{BorderRadius: 10}).Variant(giode.IconButtonGhost).
		Label("Favorite").OnClick(func() { u.eng.SetFavorite(c.peer.Device, !c.peer.Favorite) })
	return c
}

// deviceCardFor returns the card of a peer, keeping its state across frames.
func (u *UI) deviceCardFor(p engine.Peer) *deviceCard {
	c := giode.Keyed("device", p.Fingerprint, func() *deviceCard { return newDeviceCard(u) })
	c.peer = p
	return c
}

func (c *deviceCard) content(ps giode.PressState) giode.Element {
	p := c.peer
	bg, border := "theme.surface", "theme.border"
	shadow := giode.BoxShadow{}
	if ps.Hovered {
		border = "theme.primary"
		shadow = giode.BoxShadow{Y: 10, Blur: 30, Color: "#00000033"}
	}
	if ps.Pressed {
		bg = colPrimarySoft
	}
	if p.Favorite {
		c.star.Icon(icon("starfilled", 18, "#FBBF24"))
	} else {
		c.star.Icon(icon("star", 18, "theme.muted"))
	}
	sub := p.Model()
	if sub == "" {
		sub = p.IP
	}
	return giode.Stack(giode.Styles{Width: 196, Gap: 12, Padding: giode.UniformInset(16), Background: giode.Color(bg), BorderRadius: 20,
		BorderWidth: 1, BorderColor: giode.Color(border), BoxShadow: shadow, Transition: giode.StyleTransition{Duration: 120e6}},
		giode.Row(giode.Styles{Width: -1, Align: giode.AlignStart, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{Width: 52, Height: 52, BorderRadius: 26, BackgroundGradient: avatarGradient(p.Fingerprint), Align: giode.AlignCenter, Justify: giode.JustifyCenter,
				BoxShadow: giode.BoxShadow{Y: 6, Blur: 16, Color: "#0000002A"}},
				icon(deviceIcon(p.DeviceType), 24, "#FFFFFF")),
			spacer(),
			c.star,
		),
		giode.Stack(giode.Styles{Gap: 3, Width: -1},
			giode.Text(p.Alias, giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightSemiBold, FontSize: 15, Color: "theme.text", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
			giode.Text(sub, giode.Styles{FontFamily: fontText, FontSize: 12.5, Color: "theme.muted", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
		),
		giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter},
			giode.Box(giode.Styles{Width: 7, Height: 7, BorderRadius: 4, Background: "theme.success"}),
			giode.Text(p.IP, stSmall),
		),
	)
}

func (c *deviceCard) Layout(gtx layout.Context) layout.Dimensions {
	c.press.Label("Send to " + c.peer.Alias)
	return c.press.Layout(gtx)
}
