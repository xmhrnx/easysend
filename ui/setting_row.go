package ui

import (
	"github.com/xmhrnx/giode"
	prop "github.com/xmhrnx/giode/styles/properties"
)

func settingText(title, desc string) giode.Element {
	text := []giode.Element{giode.Text(title, stStrong)}
	if desc != "" {
		text = append(text, giode.Text(desc, stSmall))
	}
	return giode.Stack(giode.Styles{Gap: 3, FlexGrow: 1}, text...)
}

// settingRow is a setting with a small control (a switch) on its right,
// on every screen size.
func settingRow(title, desc string, control giode.Element) giode.Element {
	return giode.Row(giode.Styles{Width: -1, Gap: 16, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(0, 4)},
		settingText(title, desc), control)
}

// settingBlock is a setting with a wide control (a field, a picker): on
// its right on wide windows, under it at full width on phones.
func settingBlock(title, desc string, control giode.Element, width int) giode.Element {
	if giode.Compact() {
		return giode.Stack(giode.Styles{Width: -1, Gap: 10, Padding: giode.SymmetricInset(0, 4)},
			settingText(title, desc), giode.Box(giode.Styles{Width: -1}, control))
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 24, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(0, 4)},
		settingText(title, desc), giode.Box(giode.Styles{Width: prop.Width(width)}, control))
}

// settingsCard groups related settings under a title.
func settingsCard(title, iconName string, rows ...giode.Element) giode.Element {
	kids := []giode.Element{giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Padding: giode.Inset{Bottom: 4}},
		tile(iconName, 32, "theme.primary"), giode.Text(title, stHeading))}
	for i, r := range rows {
		if i > 0 {
			kids = append(kids, giode.Divider(giode.Styles{Color: "theme.border"}))
		}
		kids = append(kids, r)
	}
	return card(giode.Styles{Gap: 12}, kids...)
}
