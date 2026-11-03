package ui

import (
	"strconv"

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
	c := giode.Keyed("share", v.Token, func() *shareCard {
		c := &shareCard{u: u}
		c.where = giode.Input.Segmented("Same network", "Internet")
		c.copy = softButton("Copy link", "copy", func() { u.copy(c.url(), "Link") })
		c.stop = dangerButton("Stop sharing", "x", func() { u.eng.StopShare(c.v.Token) })
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
	s := plural(len(v.Items), "item", "items") + " · " + item.HumanSize(v.Size())
	if v.MaxDownloads > 0 {
		s += " · " + strconv.Itoa(v.Downloads) + "/" + strconv.Itoa(v.MaxDownloads) + " downloads"
	} else {
		s += " · " + plural(v.Downloads, "download", "downloads")
	}
	if !v.Expires.IsZero() {
		s += " · until " + v.Expires.Format("Jan 2 15:04")
	}
	return s
}

func (c *shareCard) Layout(gtx layout.Context) layout.Dimensions {
	v := c.v
	status := chip("Live", "theme.success")
	if v.Expired() {
		status = chip("Ended", "theme.muted")
	}
	left := []giode.Element{
		giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter}, status, func() giode.Element {
			if v.PIN != "" {
				return chip("PIN "+v.PIN, "theme.warning")
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
		left = append(left, otherAddresses(v.OtherURLs, "url-alt-"+v.Token))
	}
	return card(giode.Styles{},
		sideBySide(giode.Styles{Width: -1, Gap: 22, Align: giode.AlignCenter},
			qrView(c.url(), 168),
			giode.Stack(giode.Styles{Gap: 10, FlexGrow: 1, MinWidth: 240}, left...),
		),
	).Layout(gtx)
}
