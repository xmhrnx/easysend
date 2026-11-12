package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"

	"easysend/internal/item"
)

// selectionPanel lists what will be sent.
type selectionPanel struct {
	u     *UI
	clear *giode.ButtonElement
}

const selectionShown = 6

func newSelectionPanel(u *UI) *selectionPanel {
	return &selectionPanel{u: u, clear: giode.Button(tr("Clear"), giode.Styles{Color: "theme.muted", Background: colSurface3, BorderRadius: 10, FontSize: 13, Padding: giode.SymmetricInset(12, 6)}).
		Variant(giode.VariantGhost).OnClick(u.clearItems)}
}

func (p *selectionPanel) Layout(gtx layout.Context) layout.Dimensions {
	items := p.u.items
	rows := []giode.Element{
		sectionTitle(count("items", len(items))+" · "+item.HumanSize(item.TotalSize(items)), p.clear),
	}
	for i, it := range items {
		if i == selectionShown {
			rows = append(rows, giode.Text(count("more_items", len(items)-selectionShown), stMuted))
			break
		}
		rows = append(rows, p.row(it))
	}
	return card(giode.Styles{Gap: 10}, rows...).Layout(gtx)
}
