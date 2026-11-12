package ui

import (
	"time"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"

	"easysend/internal/engine"
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
	p.create = primaryButton(tr("Create code"), "zap", p.start)
	return p
}

func (p *internetPanel) start() {
	if len(p.u.items) == 0 {
		p.u.toast(giode.ToastInfo, tr("Choose something to send first"))
		return
	}
	items := append([]item.Item(nil), p.u.items...)
	id := p.u.eng.SendInternet(items)
	p.u.pin.watch(id, engine.Peer{}, items) // takes them off the list once delivered
}

func (p *internetPanel) Layout(gtx layout.Context) layout.Dimensions {
	kids := []giode.Element{
		card(giode.Styles{},
			sideBySide(giode.Styles{Width: -1, Gap: 18, Align: giode.AlignCenter},
				orb("globe", 64),
				giode.Stack(giode.Styles{Gap: 6, FlexGrow: 1, MinWidth: 240},
					giode.Text(tr("Send anywhere with a code"), stHeading),
					giode.Text(tr("Works across the internet, between any networks. Devices connect directly when they can, through a relay when they can't."), stMuted),
				),
				p.create,
			),
			giode.Row(giode.Styles{Gap: 8},
				chip(tr("End-to-end encrypted"), "theme.success"),
				chip(tr("Peer-to-peer"), "theme.primary"),
				chip(tr("Works with Magic Wormhole apps"), "theme.accent"),
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
