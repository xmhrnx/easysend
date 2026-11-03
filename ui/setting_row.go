package ui

import "github.com/xmhrnx/giode"

// settingRow is a labeled control in a settings card.
func settingRow(title, desc string, control giode.Element) giode.Element {
	text := []giode.Element{giode.Text(title, stStrong)}
	if desc != "" {
		text = append(text, giode.Text(desc, stSmall))
	}
	return sideBySide(giode.Styles{Width: -1, Gap: 16, Align: giode.AlignCenter, Justify: giode.JustifySpaceBetween, Padding: giode.SymmetricInset(0, 4)},
		giode.Stack(giode.Styles{Gap: 3, FlexGrow: 1, MinWidth: 200}, text...),
		control,
	)
}

// settingsCard groups related settings under a title.
func settingsCard(title, iconName string, rows ...giode.Element) giode.Element {
	kids := []giode.Element{giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Padding: giode.Inset{Bottom: 4}},
		tile(iconName, 32, colPrimarySoft, "theme.primary"), giode.Text(title, stHeading))}
	for i, r := range rows {
		if i > 0 {
			kids = append(kids, giode.Divider(giode.Styles{Color: "theme.border"}))
		}
		kids = append(kids, r)
	}
	return card(giode.Styles{Gap: 12}, kids...)
}
