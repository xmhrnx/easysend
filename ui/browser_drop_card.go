package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/segmented"
	"github.com/xmhrnx/giode/components/toggle"

	"easysend/internal/settings"
)

// browserDropCard lets people without the app upload from a browser.
type browserDropCard struct {
	u     *UI
	on    *toggle.Switch
	where *segmented.Control
	copy  *giode.ButtonElement
}

func newBrowserDropCard(u *UI) *browserDropCard {
	c := &browserDropCard{u: u}
	c.on = giode.Input.Switch("Accept uploads from browsers", giode.Styles{Color: "theme.text"}).OnChange(func(on bool) {
		_ = u.eng.UpdateSettings(func(s *settings.Settings) { s.BrowserDrop = on })
	})
	c.where = giode.Input.Segmented("Same network", "Internet")
	c.copy = softButton("Copy link", "copy", func() { u.copy(c.url(), "Link") })
	return c
}

func (c *browserDropCard) url() string {
	lan, pub := c.u.eng.DropURLs()
	if pub != "" && c.where.Selected() == 1 {
		return pub
	}
	return lan
}

func (c *browserDropCard) Layout(gtx layout.Context) layout.Dimensions {
	s := c.u.eng.Settings.Get()
	c.on.SetChecked(s.BrowserDrop)
	kids := []giode.Element{
		giode.Row(giode.Styles{Gap: 12, Align: giode.AlignCenter}, orb("qr", 40), giode.Text("Receive from a browser", stHeading)),
		giode.Text("No app on the other phone or computer? Let them scan this code and upload in their browser.", stMuted),
		c.on,
	}
	if s.BrowserDrop {
		_, pub := c.u.eng.DropURLs()
		if pub != "" {
			kids = append(kids, c.where)
		}
		kids = append(kids,
			sideBySide(giode.Styles{Width: -1, Gap: 16, Align: giode.AlignCenter},
				qrView(c.url(), 140),
				giode.Stack(giode.Styles{Gap: 10, FlexGrow: 1, MinWidth: 160},
					giode.Text(c.url(), giode.Styles{FontFamily: fontText, FontSize: 12.5, Color: "theme.text"}).Selectable("drop-url"),
					c.copy,
				),
			),
		)
	}
	return card(giode.Styles{FlexGrow: 1, Width: 0, MinWidth: 300}, kids...).Layout(gtx)
}
