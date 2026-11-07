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
			Label(e.label).
			OnClick(func(giode.ClickEvent) { u.Go(e.screen) })
		s.items = append(s.items, p)
	}
	s.self = giode.Pressable(func(ps giode.PressState) giode.Element { return s.deviceCard(ps) }).
		Label("This device").
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
		giode.Text(e.label, giode.Styles{FontFamily: fontText, FontSize: 14.5, FontWeight: weight, Color: giode.Color(fg)}),
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
			giode.Text("Share anything, anywhere", stSmall),
		),
	)
	return giode.Stack(giode.Styles{Width: 252, Height: -1, Padding: giode.UniformInset(16), Gap: 4, Background: "theme.surface"},
		append(append([]giode.Element{brand}, navs...), spacer(), s.self)...,
	).Layout(gtx)
}

// logoElement renders the app icon at size dp.
func (u *UI) logoElement(size int) giode.Element {
	if u.logo == nil {
		return orb("send", size)
	}
	return giode.Image(u.logo, giode.Styles{Width: dpW(size), Height: dpH(size)})
}
