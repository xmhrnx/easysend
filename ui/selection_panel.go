package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/iconbutton"

	"easysend/internal/item"
)

const selectionShown = 6

// selectionPanel lists what will be sent.
type selectionPanel struct {
	u     *UI
	clear *giode.ButtonElement
}

func newSelectionPanel(u *UI) *selectionPanel {
	return &selectionPanel{u: u, clear: giode.Button("Clear", giode.Styles{Color: "theme.muted", Background: colSurface3, BorderRadius: 10, FontSize: 13, Padding: giode.SymmetricInset(12, 6)}).
		Variant(giode.VariantGhost).OnClick(u.clearItems)}
}

func (p *selectionPanel) Layout(gtx layout.Context) layout.Dimensions {
	items := p.u.items
	rows := []giode.Element{
		sectionTitle(plural(len(items), "item", "items")+" · "+item.HumanSize(item.TotalSize(items)), p.clear),
	}
	for i, it := range items {
		if i == selectionShown {
			rows = append(rows, giode.Text("+ "+plural(len(items)-selectionShown, "more item", "more items"), stMuted))
			break
		}
		rows = append(rows, p.row(it))
	}
	return card(giode.Styles{Gap: 10}, rows...).Layout(gtx)
}

func (p *selectionPanel) row(it item.Item) giode.Element {
	remove := giode.Keyed("sel-remove", it.ID, func() *iconbutton.Button {
		id := it.ID
		return iconAction("x", "Remove", func() { p.u.removeItem(id) })
	})
	name, sub, ic := it.Name, item.HumanSize(it.Size), fileIcon(it.Name)
	if it.Kind == item.Text {
		name, sub, ic = firstLine(it.Text, 60), "Text · "+item.HumanSize(it.Size), "text"
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
