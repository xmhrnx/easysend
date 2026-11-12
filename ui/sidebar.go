package ui

import (
	"strconv"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
)

// sidebar is the navigation column on wide windows.
type sidebar struct {
	u     *UI
	items []*giode.PressableElement
	self  *giode.PressableElement
}

func newSidebar(u *UI) *sidebar {
	s := &sidebar{u: u}
	for _, e := range navEntries {
		e := e
		p := giode.Pressable(func(ps giode.PressState) giode.Element { return s.item(e, ps) }).
			Label(tr(e.label)).
			OnClick(func(giode.ClickEvent) { u.Go(e.screen) })
		s.items = append(s.items, p)
	}
	s.self = giode.Pressable(func(ps giode.PressState) giode.Element { return s.deviceCard(ps) }).
		Label(tr("This device")).
		OnClick(func(giode.ClickEvent) { u.Go(screenReceive) })
	return s
}

func (s *sidebar) item(e navEntry, ps giode.PressState) giode.Element {
	active := s.u.current == e.screen
	bg, border, fg, ic := "transparent", "transparent", "theme.muted", "theme.muted"
	switch {
	case active:
		border, fg, ic = colRing, "theme.primary", "theme.primary"
	case ps.Pressed:
		bg, fg = colPressed, "theme.text"
	case ps.Hovered:
		bg, fg = colHover, "theme.text"
	}
	weight := giode.FontWeightMedium
	if active {
		weight = giode.FontWeightSemiBold
	}
	kids := []giode.Element{
		icon(e.icon, 20, ic),
		giode.Text(tr(e.label), giode.Styles{FontFamily: fontText, FontSize: 14.5, FontWeight: weight, Color: giode.Color(fg)}),
		spacer(),
	}
	if n := s.u.badge(e.screen); n > 0 {
		kids = append(kids, chip(strconv.Itoa(n), "theme.primary"))
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Padding: giode.SymmetricInset(14, 11), BorderRadius: 12, Background: giode.Color(bg), BorderWidth: 1, BorderColor: giode.Color(border), Wrap: giode.WrapNoWrap}, kids...)
}

func (s *sidebar) Layout(gtx layout.Context) layout.Dimensions {
	navs := make([]giode.Element, len(s.items))
	for i, p := range s.items {
		navs[i] = p
	}
	brand := giode.Row(giode.Styles{Gap: 12, Align: giode.AlignCenter, Padding: giode.Inset{Left: 6, Bottom: 22, Top: 6}},
		s.u.logoElement(38),
		giode.Stack(giode.Styles{Gap: 0},
			giode.Text("EasySend", giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 19, Color: "theme.text"}),
			giode.Text(tr("Share anything, anywhere"), stSmall),
		),
	)
	kids := append(append([]giode.Element{brand}, navs...), spacer())
	// On short windows (a phone held sideways) the device card would be
	// cut off; the Receive screen shows the same.
	if gtx.Metric.PxToDp(gtx.Constraints.Max.Y) >= sidebarCardMinHeight {
		kids = append(kids, s.self)
	}
	safe := giode.SafeArea()
	return giode.Stack(giode.Styles{Width: dpW(252 + safe.Left), Height: -1, Gap: 4, Background: "theme.surface",
		Padding: giode.Inset{Top: 16 + safe.Top, Bottom: 16 + safe.Bottom, Left: 16 + safe.Left, Right: 16}},
		kids...,
	).Layout(gtx)
}

// sidebarCardMinHeight is the window height, in dp, from which the
// sidebar has room for the device card.
const sidebarCardMinHeight = 520

// logoElement renders the app icon at size dp.
func (u *UI) logoElement(size int) giode.Element {
	if u.logo == nil {
		return orb("send", size)
	}
	return giode.Image(u.logo, giode.Styles{Width: dpW(size), Height: dpH(size)})
}
