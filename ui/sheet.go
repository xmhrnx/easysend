package ui

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/sheet"
)

// newSheet is a modal panel: a bottom sheet on phones, a centered dialog
// from 600dp wide.
func newSheet(open *bool, title string, body, footer func() giode.Element) *sheet.Sheet {
	s := giode.Sheet(open, body, giode.Styles{Background: "theme.surface", Color: "theme.text"}).Adaptive(true)
	if title != "" {
		s.Title(title)
	}
	if footer != nil {
		s.Footer(footer)
	}
	return s
}

// sheetActions lays out a sheet's buttons side by side, equally wide.
func sheetActions(buttons ...*giode.ButtonElement) giode.Element {
	kids := make([]giode.Element, len(buttons))
	for i, b := range buttons {
		kids[i] = giode.Box(giode.Styles{FlexGrow: 1}, b)
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 10, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Top: 4, Bottom: 8}}, kids...)
}

// wide makes a button fill its slot (sheet footers).
func wide(b *giode.ButtonElement, st giode.Styles) *giode.ButtonElement {
	st.Width = -1
	st.Padding = giode.SymmetricInset(16, 13)
	return b.Styles(st)
}

// sheetPrimary and sheetSecondary are the full-width sheet buttons.
func sheetPrimary(label, iconName string, fn func()) *giode.ButtonElement {
	return wide(primaryButton(label, iconName, fn), primaryStyles)
}

func sheetSecondary(label string, fn func()) *giode.ButtonElement {
	return wide(softButton(label, "", fn), giode.Styles{BorderColor: "theme.border", Background: "theme.text", Color: "theme.text", BorderWidth: 1, BorderRadius: 12, FontSize: 14})
}
