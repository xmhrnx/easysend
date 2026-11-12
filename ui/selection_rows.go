package ui

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/iconbutton"

	"easysend/internal/item"
)

func (p *selectionPanel) row(it item.Item) giode.Element {
	remove := giode.Keyed(scope("sel-remove"), it.ID, func() *iconbutton.Button {
		id := it.ID
		return iconAction("x", tr("Remove"), func() { p.u.removeItem(id) })
	})
	name, sub, ic := it.Name, item.HumanSize(it.Size), fileIcon(it.Name)
	if it.Kind == item.Text {
		name, sub, ic = firstLine(it.Text, 60), tr("Text · {size}", "size", item.HumanSize(it.Size)), "text"
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		tile(ic, 40, "theme.accent"),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
			giode.Text(name, giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightMedium, FontSize: 14, Color: "theme.text", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
			giode.Text(sub, stSmall)),
		remove,
	)
}

// firstLine shortens text to its first line, at most n runes.
func firstLine(s string, n int) string {
	for i, r := range s {
		if r == '\n' {
			s = s[:i] + "…"
			break
		}
	}
	rs := []rune(s)
	if len(rs) > n {
		return string(rs[:n]) + "…"
	}
	return s
}
