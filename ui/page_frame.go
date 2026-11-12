package ui

import (
	"github.com/xmhrnx/giode"
)

// pageFrame lays out a screen: a header and a centered column of content.
func pageFrame(title, subtitle string, trailing giode.Element, body ...giode.Element) giode.Element {
	pad := 32
	titleSt := stTitle
	if giode.Compact() {
		pad = 16
		titleSt.FontSize = 26
	}
	head := []giode.Element{
		giode.Stack(giode.Styles{Gap: 4, FlexGrow: 1},
			giode.Text(title, titleSt),
			giode.Text(subtitle, giode.Styles{FontFamily: fontText, FontSize: 14.5, Color: "theme.muted"}),
		),
	}
	if trailing != nil {
		head = append(head, trailing)
	}
	kids := append([]giode.Element{giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 12, Padding: giode.Inset{Bottom: 8}}, head...)}, body...)
	bottom := 48
	if giode.Compact() {
		bottom = navClearance() // out from under the floating tab bar
	}
	return giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyCenter, Padding: giode.Inset{Top: pad, Left: pad, Right: pad, Bottom: bottom}},
		giode.Stack(giode.Styles{Width: -1, MaxWidth: 980, Gap: 20}, kids...),
	)
}

// sectionTitle is a heading inside a screen, with optional trailing items.
func sectionTitle(title string, trailing ...giode.Element) giode.Element {
	kids := append([]giode.Element{giode.Text(title, stHeading), spacer()}, trailing...)
	return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 8, Wrap: giode.WrapNoWrap}, kids...)
}

// scrollStyles is the scrollbar look shared by the screens.
var scrollStyles = giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, ThumbColor: colSurface3, TrackColor: "transparent", Inset: 4, AutoHide: true}}

// compactOr picks a size for phones or for wide windows.
func compactOr(phone, wide int) int {
	if giode.Compact() {
		return phone
	}
	return wide
}

// sideBySide lays kids in a row on wide windows and lets them wrap onto
// several lines on phones.
func sideBySide(st giode.Styles, kids ...giode.Element) giode.Element {
	// Wide windows can still be narrow (a phone held sideways, a split
	// iPad): wrap whenever the kids don't fit rather than overflow.
	st.Wrap = giode.WrapWrap
	return giode.Row(st, kids...)
}
