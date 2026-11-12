package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/scroll"
	"github.com/xmhrnx/giode/components/toggle"

	"easysend/internal/settings"
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
	r.visible = newSwitch(tr("Visible to nearby devices"), giode.Styles{Color: "theme.text"}).OnChange(func(on bool) {
		_ = u.eng.UpdateSettings(func(s *settings.Settings) { s.Discoverable = on })
	})
	r.code = giode.Input.Text(tr("e.g. 7-crossover-clockwork"), giode.Styles{Background: colSurface2, BorderColor: "theme.border", BorderRadius: 12, Color: "theme.text"}).
		Submit(func(string) { r.start() })
	r.receive = primaryButton(tr("Receive"), "download", r.start)
	r.scroll = giode.Scroll(r.content, scrollStyles)
	return r
}

func (r *receiveScreen) start() {
	code := r.code.Text()
	if code == "" {
		r.u.toast(giode.ToastInfo, tr("Enter the code the sender gave you"))
		return
	}
	r.u.eng.ReceiveInternet(code)
	r.code.SetText("")
	r.u.app.HideKeyboard()
	r.u.toast(giode.ToastInfo, tr("Connecting…"))
}

func (r *receiveScreen) Layout(gtx layout.Context) layout.Dimensions { return r.scroll.Layout(gtx) }
