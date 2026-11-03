package ui

import (
	"strings"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/iconbutton"

	"easysend/internal/transfer"
)

// codeCard shows the code of an internet transfer, with a QR code
// Magic Wormhole apps scan.
type codeCard struct {
	u      *UI
	t      transfer.Transfer
	copy   *giode.ButtonElement
	cancel *iconbutton.Button
}

func (u *UI) codeCardFor(t transfer.Transfer) *codeCard {
	c := giode.Keyed("code", t.ID, func() *codeCard {
		c := &codeCard{u: u}
		c.copy = softButton("Copy code", "copy", func() { u.copy(c.t.Code, "Code") })
		c.cancel = iconAction("x", "Cancel", func() { u.eng.Transfers.Cancel(c.t.ID) })
		return c
	})
	c.t = t
	return c
}

func (c *codeCard) Layout(gtx layout.Context) layout.Dimensions {
	t := c.t
	head := sectionTitle(t.Title(), c.cancel)
	if t.Status.Finished() {
		head = sectionTitle(t.Title(), statusChip(t))
	}
	var left []giode.Element
	if t.Code == "" {
		left = []giode.Element{
			giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter},
				giode.Spinner(giode.Styles{Width: 22, Height: 22, Color: "theme.primary"}),
				giode.Text("Creating a code…", stStrong)),
		}
	} else {
		left = []giode.Element{
			giode.Text("TELL THE RECEIVER THIS CODE", stOverline),
			giode.Text(strings.ToUpper(t.Code[:1])+t.Code[1:], giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 30, Color: "theme.primary"}),
			giode.Text("They enter it in EasySend → Receive, the wormhole command or any Magic Wormhole app.", stMuted),
			c.copy,
		}
	}
	left = append(left, giode.Box(giode.Styles{Height: 4}), giode.Text(statusText(t), stSmall))
	if t.Status == transfer.Active {
		left = append(left, c.u.progressBar(t.Progress()))
	}
	kids := []giode.Element{giode.Stack(giode.Styles{Gap: 8, FlexGrow: 1, MinWidth: 240}, left...)}
	if t.Code != "" && !t.Status.Finished() {
		kids = append(kids, qrView("wormhole-transfer:"+t.Code, 150))
	}
	return card(giode.Styles{},
		head,
		sideBySide(giode.Styles{Width: -1, Gap: 20, Align: giode.AlignCenter}, kids...),
	).Layout(gtx)
}
