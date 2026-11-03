package ui

import (
	"time"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dropdown"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/toggle"

	"easysend/internal/engine"
	"easysend/internal/item"
)

var expiryChoices = []struct {
	label string
	d     time.Duration
}{{"1 hour", time.Hour}, {"24 hours", 24 * time.Hour}, {"7 days", 7 * 24 * time.Hour}, {"Never", 0}}

var limitChoices = []struct {
	label string
	n     int
}{{"Unlimited", 0}, {"1 download", 1}, {"5 downloads", 5}, {"10 downloads", 10}}

// linkPanel shares through a link and QR code anyone can open in a
// browser, without the app.
type linkPanel struct {
	u      *UI
	expiry *dropdown.Select
	limit  *dropdown.Select
	lock   *toggle.Switch
	pin    *input.Input
	create *giode.ButtonElement
}

func newLinkPanel(u *UI) *linkPanel {
	p := &linkPanel{u: u}
	sel := giode.Styles{Background: colSurface2, BorderColor: "theme.border", BorderRadius: 10, Color: "theme.text", Width: 160}
	var exp, lim []string
	for _, c := range expiryChoices {
		exp = append(exp, c.label)
	}
	for _, c := range limitChoices {
		lim = append(lim, c.label)
	}
	p.expiry = giode.Input.Select(sel, exp...)
	p.expiry.SetSelected(1)
	p.limit = giode.Input.Select(sel, lim...)
	p.lock = giode.Input.Switch("", giode.Styles{})
	p.pin = giode.Input.Text("PIN", giode.Styles{Width: 110, Background: colSurface2, BorderColor: "theme.border", BorderRadius: 10, Color: "theme.text"}).Keyboard(giode.KeyboardNumeric)
	p.create = u.primaryButton("Create link", "link", p.start)
	return p
}

func (p *linkPanel) start() {
	if len(p.u.items) == 0 {
		p.u.toast(giode.ToastInfo, "Choose something to send first")
		return
	}
	o := engine.ShareOptions{Expires: expiryChoices[p.expiry.Selected()].d, MaxDownloads: limitChoices[p.limit.Selected()].n}
	if p.lock.Checked() {
		o.PIN = p.pin.Text()
		if o.PIN == "" {
			p.u.toast(giode.ToastWarning, "Enter a PIN or turn the lock off")
			return
		}
	}
	p.u.eng.CreateShare(append([]item.Item(nil), p.u.items...), o)
	p.u.toast(giode.ToastSuccess, "Link created — scan the QR code to download")
}

func option(label string, el giode.Element) giode.Element {
	return giode.Stack(giode.Styles{Gap: 6}, giode.Text(label, stOverline), el)
}

func (p *linkPanel) Layout(gtx layout.Context) layout.Dimensions {
	s := p.u.eng.Settings.Get()
	lockRow := []giode.Element{p.lock}
	if p.lock.Checked() {
		lockRow = append(lockRow, p.pin)
	}
	kids := []giode.Element{
		card(giode.Styles{},
			sideBySide(giode.Styles{Width: -1, Gap: 18, Align: giode.AlignCenter},
				orb(s, "qr", 64),
				giode.Stack(giode.Styles{Gap: 6, FlexGrow: 1, MinWidth: 240},
					giode.Text("Share with a link or QR code", stHeading),
					giode.Text("Anyone can download in their browser — no app needed. Scan the code with a phone camera.", stMuted),
				),
			),
			giode.Row(giode.Styles{Gap: 18, Align: giode.AlignEnd, Width: -1},
				option("Expires after", p.expiry),
				option("Limit", p.limit),
				option("PIN lock", giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Height: 40}, lockRow...)),
				spacer(),
				p.create,
			),
			p.u.publicNote(),
		),
	}
	for _, v := range p.u.eng.Shares() {
		kids = append(kids, p.u.shareCardFor(v))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 16}, kids...).Layout(gtx)
}

// publicNote says whether browser links work from the internet.
func (u *UI) publicNote() giode.Element {
	st := u.eng.Public()
	var ic, fg, text string
	switch {
	case u.eng.Settings.Get().PublicURL != "":
		ic, fg, text = "globe", "theme.success", "Links also work from the internet through "+u.eng.PublicBase()
	case st.State == engine.PublicReady:
		ic, fg, text = "globe", "theme.success", "Links also work from the internet: your router opened a port (UPnP)."
	case st.State == engine.PublicWorking:
		ic, fg, text = "refresh", "theme.muted", "Checking whether links can work from the internet…"
	case st.State == engine.PublicUnavailable:
		ic, fg, text = "info", "theme.warning", "Links work on your network. "+st.Err+" Use Internet → code for people elsewhere."
	default:
		ic, fg, text = "wifi", "theme.muted", "Links work for people on your network. Turn on UPnP in Settings to reach the internet."
	}
	return giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Padding: giode.SymmetricInset(12, 10), Background: colSurface2, BorderRadius: 12, Width: -1, Wrap: giode.WrapNoWrap},
		icon(ic, 16, fg),
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(text, stSmall)),
	)
}
