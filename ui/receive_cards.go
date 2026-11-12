package ui

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"easysend/internal/transfer"
)

func (r *receiveScreen) hero() giode.Element {
	s := r.u.eng.Settings.Get()
	self := r.u.eng.Self()
	r.visible.SetChecked(s.Discoverable)
	pinChip := chip(tr("No PIN"), "theme.muted")
	if s.RequirePIN && s.PIN != "" {
		pinChip = chip(tr("PIN required"), "theme.warning")
	}
	accept := []string{tr("Asks before saving"), tr("Auto-accepts favorites"), tr("Auto-accepts everyone")}[s.AutoAccept]
	return card(giode.Styles{Padding: giode.UniformInset(24)},
		sideBySide(giode.Styles{Width: -1, Gap: 24, Align: giode.AlignCenter, Justify: giode.JustifyCenter},
			radar(compactOr(124, 170), "inbox"),
			giode.Stack(giode.Styles{Gap: 8, FlexGrow: 1, MinWidth: 260},
				giode.Text(tr("READY TO RECEIVE"), giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightBold, FontSize: 11, Color: "theme.success"}),
				giode.Text(self.Name, giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 26, Color: "theme.text"}),
				giode.Text(tr("Nearby devices running EasySend or LocalSend see this name."), stMuted),
				giode.Row(giode.Styles{Gap: 8},
					chip(self.IP+":"+strconv.Itoa(self.Port), "theme.text"),
					pinChip,
					chip(accept, "theme.primary"),
				),
				giode.Box(giode.Styles{Width: -1, MaxWidth: 360, Padding: giode.Inset{Top: 6}}, r.visible),
			),
		),
	)
}

func (r *receiveScreen) codeCard() giode.Element {
	return card(giode.Styles{FlexGrow: 1, Width: 0, MinWidth: 300},
		giode.Row(giode.Styles{Gap: 12, Align: giode.AlignCenter}, orb("globe", 40), giode.Text(tr("Receive with a code"), stHeading)),
		giode.Text(tr("From anywhere in the world: type the code the sender sees in EasySend, the wormhole command or a Magic Wormhole app."), stMuted),
		giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, giode.Box(giode.Styles{FlexGrow: 1}, r.code), r.receive),
	)
}

func (r *receiveScreen) recent() []giode.Element {
	var rows []giode.Element
	for _, t := range r.u.eng.Transfers.List() {
		if t.Direction != transfer.Incoming {
			continue
		}
		rows = append(rows, r.u.transferRowFor(t))
		if len(rows) == 5 {
			break
		}
	}
	if len(rows) == 0 {
		return nil
	}
	return append([]giode.Element{sectionTitle(tr("Recently received"))}, rows...)
}

func (r *receiveScreen) content() giode.Element {
	body := []giode.Element{
		r.hero(),
		giode.Row(giode.Styles{Width: -1, Gap: 20, Align: giode.AlignStretch}, r.codeCard(), r.browser),
	}
	body = append(body, r.recent()...)
	return pageFrame(tr("Receive"), tr("Others can send to you nearby, with a code, or from any browser."), nil, body...)
}
