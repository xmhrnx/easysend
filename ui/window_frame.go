package ui

import (
	"strconv"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/titlebar"

	"easysend/internal/transfer"
	"easysend/ui/platform"
)

// windowFrame is the custom title bar of the frameless desktop window: the
// logo and name, a live status, and the window buttons.
type windowFrame struct {
	u   *UI
	bar *titlebar.TitleBar
}

func newWindowFrame(u *UI) *windowFrame {
	f := &windowFrame{u: u}
	variant := giode.TitleBarWindows
	if platform.MacWindowControls {
		variant = giode.TitleBarMac
	}
	f.bar = giode.TitleBar(u.app, "EasySend", giode.Styles{
		Background: "theme.background", Color: "theme.text", Height: 40,
		FontSize: 13, FontWeight: giode.FontWeightSemiBold,
		BorderWidth: 1, BorderColor: "theme.border",
	}).Variant(variant).Icon(u.logoElement(18)).
		Trailing(giode.Raw(func(gtx layout.Context) layout.Dimensions {
			return giode.Box(giode.Styles{Padding: giode.Inset{Right: 8, Left: 8}}, f.status()).Layout(gtx)
		}))
	return f
}

// status summarizes what's happening for the bar: active transfers, or
// whether nearby devices can see this one.
func (f *windowFrame) status() giode.Element {
	active := 0
	for _, t := range f.u.eng.Transfers.List() {
		if !t.Status.Finished() && t.Status != transfer.Waiting {
			active++
		}
	}
	dot, label := "theme.success", "Visible nearby"
	switch {
	case active > 0:
		dot, label = "theme.primary", plural(active, "transfer", "transfers")+" in progress"
	case len(f.u.eng.Requests()) > 0:
		dot, label = "theme.primary", strconv.Itoa(len(f.u.eng.Requests()))+" waiting for you"
	case !f.u.eng.Settings.Get().Discoverable:
		dot, label = "theme.muted", "Hidden"
	}
	return giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter, Padding: giode.SymmetricInset(10, 3), BorderRadius: 99, BorderWidth: 1, BorderColor: "theme.border"},
		giode.Box(giode.Styles{Width: 6, Height: 6, BorderRadius: 3, Background: giode.Color(dot)}),
		giode.Text(label, giode.Styles{FontFamily: fontText, FontSize: 11.5, Color: "theme.muted"}),
	)
}

func (f *windowFrame) Layout(gtx layout.Context) layout.Dimensions { return f.bar.Layout(gtx) }
