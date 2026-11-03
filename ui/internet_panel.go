package ui

import (
	"time"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"

	"easysend/internal/item"
	"easysend/internal/transfer"
)

// internetPanel sends anywhere with a short code (Magic Wormhole).
type internetPanel struct {
	u      *UI
	create *giode.ButtonElement
}

func newInternetPanel(u *UI) *internetPanel {
	p := &internetPanel{u: u}
	p.create = u.primaryButton("Create code", "zap", p.start)
	return p
}

func (p *internetPanel) start() {
	if len(p.u.items) == 0 {
		p.u.toast(giode.ToastInfo, "Choose something to send first")
		return
	}
	p.u.eng.SendInternet(append([]item.Item(nil), p.u.items...))
}

func (p *internetPanel) Layout(gtx layout.Context) layout.Dimensions {
	s := p.u.eng.Settings.Get()
	kids := []giode.Element{
		card(giode.Styles{},
			sideBySide(giode.Styles{Width: -1, Gap: 18, Align: giode.AlignCenter},
				orb(s, "globe", 64),
				giode.Stack(giode.Styles{Gap: 6, FlexGrow: 1, MinWidth: 240},
					giode.Text("Send anywhere with a code", stHeading),
					giode.Text("Works across the internet, between any networks. Devices connect directly when they can, through a relay when they can't.", stMuted),
				),
				p.create,
			),
			giode.Row(giode.Styles{Gap: 8},
				chip("End-to-end encrypted", colSuccessSoft, "theme.success"),
				chip("Peer-to-peer", colPrimarySoft, "theme.primary"),
				chip("Works with Magic Wormhole apps", colAccentSoft, "theme.accent"),
			),
		),
	}
	for _, t := range p.u.eng.Transfers.List() {
		if t.Method != transfer.Internet || t.Direction != transfer.Outgoing {
			continue
		}
		if t.Status.Finished() && time.Since(t.Finished) > 2*time.Minute {
			continue
		}
		kids = append(kids, p.u.codeCardFor(t))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 16}, kids...).Layout(gtx)
}
