package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
)

// bottomNav is the tab bar on narrow windows and phones.
type bottomNav struct {
	u     *UI
	items []*giode.PressableElement
}

func newBottomNav(u *UI) *bottomNav {
	b := &bottomNav{u: u}
	for _, e := range navEntries {
		e := e
		p := giode.Pressable(func(ps giode.PressState) giode.Element { return b.item(e, ps) }, giode.Styles{FlexGrow: 1}).
			Label(e.label).
			OnClick(func(giode.ClickEvent) { u.Go(e.screen) })
		b.items = append(b.items, p)
	}
	return b
}

func (b *bottomNav) item(e navEntry, ps giode.PressState) giode.Element {
	active := b.u.current == e.screen
	fg := "theme.muted"
	pill := "transparent"
	if active {
		fg, pill = "theme.primary", colRing
	}
	ic := icon(e.icon, 21, fg)
	if n := b.u.badge(e.screen); n > 0 {
		ic = giode.Badge("").On(ic)
	}
	return giode.Stack(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 4, Padding: giode.SymmetricInset(0, 8)},
		giode.Box(giode.Styles{BorderWidth: 1, BorderColor: giode.Color(pill), BorderRadius: 99, Padding: giode.SymmetricInset(18, 4)}, ic),
		giode.Text(e.label, giode.Styles{FontFamily: fontText, FontSize: 11.5, FontWeight: giode.FontWeightMedium, Color: giode.Color(fg)}),
	)
}

func (b *bottomNav) Layout(gtx layout.Context) layout.Dimensions {
	kids := make([]giode.Element, len(b.items))
	for i, p := range b.items {
		kids[i] = p
	}
	return giode.Row(giode.Styles{Width: -1, Background: "theme.surface", Padding: giode.Inset{Top: 4, Bottom: 6}, Wrap: giode.WrapNoWrap,
		BoxShadow: giode.BoxShadow{Y: -2, Blur: 16, Color: "#00000022"}}, kids...).Layout(gtx)
}
