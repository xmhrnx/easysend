package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/segmented"

	"easysend/internal/engine"
	"easysend/internal/item"
)

// shareCard is a browser link with its QR code.
type shareCard struct {
	u     *UI
	v     engine.ShareView
	where *segmented.Control
	copy  *giode.ButtonElement
	stop  *giode.ButtonElement
}

func (u *UI) shareCardFor(v engine.ShareView) *shareCard {
	c := giode.Keyed(scope("share"), v.Token, func() *shareCard {
		c := &shareCard{u: u}
		c.where = giode.Input.Segmented(tr("Same network"), tr("Internet"))
		c.copy = softButton(tr("Copy link"), "copy", func() { u.copy(c.url(), tr("Link")) })
		c.stop = dangerButton(tr("Stop sharing"), "x", func() { u.eng.StopShare(c.v.Token) })
		return c
	})
	c.v = v
	return c
}

func (c *shareCard) url() string {
	if c.v.PublicURL != "" && c.where.Selected() == 1 {
		return c.v.PublicURL
	}
	return c.v.LANURL
}

func (c *shareCard) details() string {
	v := c.v
	s := count("items", len(v.Items)) + " · " + item.HumanSize(v.Size())
	if v.MaxDownloads > 0 {
		s += " · " + tr("{done}/{max} downloads", "done", v.Downloads, "max", v.MaxDownloads)
	} else {
		s += " · " + count("downloads", v.Downloads)
	}
	if !v.Expires.IsZero() {
		s += tr(" · until {when}", "when", dateTime(v.Expires))
	}
	return s
}

func (c *shareCard) Layout(gtx layout.Context) layout.Dimensions {
	v := c.v
	status := chip(tr("Live"), "theme.success")
	if v.Expired() {
		status = chip(tr("Ended"), "theme.muted")
	}
	left := []giode.Element{
		giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter}, status, func() giode.Element {
			if v.PIN != "" {
				return chip(tr("PIN {pin}", "pin", v.PIN), "theme.warning")
			}
			return giode.Box(giode.Styles{})
		}()),
		giode.Text(c.details(), stMuted),
	}
	if v.PublicURL != "" {
		left = append(left, c.where)
	}
	left = append(left,
		giode.Box(giode.Styles{Background: colSurface2, BorderRadius: 12, Padding: giode.SymmetricInset(12, 10), Width: -1},
			giode.Text(c.url(), giode.Styles{FontFamily: fontText, FontSize: 13, Color: "theme.text"}).Selectable("url-"+v.Token)),
		giode.Row(giode.Styles{Gap: 10}, c.copy, c.stop),
	)
	if c.where.Selected() == 0 || v.PublicURL == "" {
		left = append(left, linkHelp(v.OtherURLs, "url-alt-"+v.Token))
	}
	return card(giode.Styles{},
		sideBySide(giode.Styles{Width: -1, Gap: 22, Align: giode.AlignCenter},
			qrView(c.url(), 168),
			giode.Stack(giode.Styles{Gap: 10, FlexGrow: 1, MinWidth: 240}, left...),
		),
	).Layout(gtx)
}
