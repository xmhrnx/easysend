package ui

import (
	"strconv"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/scroll"
	"github.com/xmhrnx/giode/components/toggle"

	"easysend/internal/settings"
	"easysend/internal/transfer"
)

// receiveScreen shows how others reach this device.
type receiveScreen struct {
	u       *UI
	scroll  *scroll.Scroll
	visible *toggle.Switch
	code    *input.Input
	receive *giode.ButtonElement
	browser *browserDropCard
}

func newReceiveScreen(u *UI) *receiveScreen {
	r := &receiveScreen{u: u, browser: newBrowserDropCard(u)}
	r.visible = newSwitch("Visible to nearby devices", giode.Styles{Color: "theme.text"}).OnChange(func(on bool) {
		_ = u.eng.UpdateSettings(func(s *settings.Settings) { s.Discoverable = on })
	})
	r.code = giode.Input.Text("e.g. 7-crossover-clockwork", giode.Styles{Background: colSurface2, BorderColor: "theme.border", BorderRadius: 12, Color: "theme.text"}).
		Submit(func(string) { r.start() })
	r.receive = primaryButton("Receive", "download", r.start)
	r.scroll = giode.Scroll(r.content, scrollStyles)
	return r
}

func (r *receiveScreen) start() {
	code := r.code.Text()
	if code == "" {
		r.u.toast(giode.ToastInfo, "Enter the code the sender gave you")
		return
	}
	r.u.eng.ReceiveInternet(code)
	r.code.SetText("")
	r.u.app.HideKeyboard()
	r.u.toast(giode.ToastInfo, "Connecting…")
}

func (r *receiveScreen) hero() giode.Element {
	s := r.u.eng.Settings.Get()
	self := r.u.eng.Self()
	r.visible.SetChecked(s.Discoverable)
	pinChip := chip("No PIN", "theme.muted")
	if s.RequirePIN && s.PIN != "" {
		pinChip = chip("PIN required", "theme.warning")
	}
	accept := []string{"Asks before saving", "Auto-accepts favorites", "Auto-accepts everyone"}[s.AutoAccept]
	return card(giode.Styles{Padding: giode.UniformInset(24)},
		sideBySide(giode.Styles{Width: -1, Gap: 24, Align: giode.AlignCenter, Justify: giode.JustifyCenter},
			radar(compactOr(124, 170), "inbox"),
			giode.Stack(giode.Styles{Gap: 8, FlexGrow: 1, MinWidth: 260},
				giode.Text("READY TO RECEIVE", giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightBold, FontSize: 11, Color: "theme.success"}),
				giode.Text(self.Name, giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 26, Color: "theme.text"}),
				giode.Text("Nearby devices running EasySend or LocalSend see this name.", stMuted),
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
		giode.Row(giode.Styles{Gap: 12, Align: giode.AlignCenter}, orb("globe", 40), giode.Text("Receive with a code", stHeading)),
		giode.Text("From anywhere in the world: type the code the sender sees in EasySend, the wormhole command or a Magic Wormhole app.", stMuted),
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
	return append([]giode.Element{sectionTitle("Recently received")}, rows...)
}

func (r *receiveScreen) content() giode.Element {
	body := []giode.Element{
		r.hero(),
		giode.Row(giode.Styles{Width: -1, Gap: 20, Align: giode.AlignStretch}, r.codeCard(), r.browser),
	}
	body = append(body, r.recent()...)
	return pageFrame("Receive", "Others can send to you nearby, with a code, or from any browser.", nil, body...)
}

func (r *receiveScreen) Layout(gtx layout.Context) layout.Dimensions { return r.scroll.Layout(gtx) }
