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
			Label(tr(e.label)).
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
		giode.Text(tr(e.label), giode.Styles{FontFamily: fontText, FontSize: 11.5, FontWeight: giode.FontWeightMedium, Color: giode.Color(fg)}),
	)
}

func (b *bottomNav) Layout(gtx layout.Context) layout.Dimensions {
	kids := make([]giode.Element, len(b.items))
	for i, p := range b.items {
		kids[i] = p
	}
	// A floating pill over the page, which runs to the bottom edge
	// behind it (see UI.View).
	return giode.Row(giode.Styles{Width: -1, Padding: giode.SymmetricInset(4, 2), Wrap: giode.WrapNoWrap,
		Background: "theme.surface", BorderRadius: navRadius, BorderWidth: 1, BorderColor: "theme.border",
		BoxShadow: giode.BoxShadow{Y: 8, Blur: 28, Color: "#00000030"}}, kids...).Layout(gtx)
}

// The floating tab bar's size, and how far it sits from the edges.
const (
	navHeight = 68 // the pill: items and padding
	navRadius = 26
	navMargin = 12 // from the sides, and from the bottom without a home indicator
)

// keyboardInset is the smallest bottom inset taken for an on-screen
// keyboard rather than a home indicator or gesture bar (at most ~48 dp).
const keyboardInset = 120

// navBottom is the gap between the bar and the window's bottom edge:
// just above the home indicator or gesture bar, where there is one, and
// above the keyboard while it's up (on iOS it's part of the bottom
// inset; Android shrinks the window instead).
func navBottom() int {
	b := giode.SafeArea().Bottom
	if b >= keyboardInset {
		return b + navMargin/2
	}
	return max(navMargin, b-12)
}

// navClearance is the space a page leaves at its end so its last
// content can scroll up out from under the floating bar.
func navClearance() int {
	return navHeight + navBottom() + 20
}
