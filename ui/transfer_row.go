package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/iconbutton"

	"easysend/internal/engine"
	"easysend/internal/transfer"
)

// transferRow is one transfer in a list, with its actions.
type transferRow struct {
	u                          *UI
	t                          transfer.Transfer
	cancel, open, copy, remove *iconbutton.Button
	retry                      *giode.ButtonElement
}

func (u *UI) transferRowFor(t transfer.Transfer) *transferRow {
	r := giode.Keyed(scope("transfer"), t.ID, func() *transferRow {
		r := &transferRow{u: u}
		r.cancel = iconAction("x", tr("Cancel"), func() { u.eng.Transfers.Cancel(r.t.ID) })
		r.open = iconAction("folder", showReceivedLabel(), func() { u.showReceived(r.t) })
		r.copy = iconAction("copy", tr("Copy text"), func() { u.copy(r.t.Text, tr("Text")) })
		r.remove = iconAction("trash", tr("Remove from history"), func() {
			id := r.t.ID
			giode.AfterLayout(func() { u.eng.Transfers.Remove(id) })
		})
		r.retry = softButton(tr("Enter PIN"), "lock", func() { u.pin.reopen(r.t) })
		return r
	})
	r.t = t
	return r
}

func (r *transferRow) Layout(gtx layout.Context) layout.Dimensions {
	t := r.t
	dirIcon, fg := "upload", "theme.primary"
	verb := tr("to {name}", "name", t.Peer)
	if t.Direction == transfer.Incoming {
		dirIcon, fg, verb = "download", "theme.primary", tr("from {name}", "name", t.Peer)
	}
	if t.Text != "" {
		dirIcon = "message"
	}
	sub := verb + " · " + methodLabel(t.Method) + " · " + timeAgo(t.Started)
	info := []giode.Element{
		giode.Text(titleOf(t), giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightSemiBold, FontSize: 14.5, Color: "theme.text", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
		giode.Text(sub, stSmall),
	}
	if t.Text != "" && t.Direction == transfer.Incoming {
		info = append(info, giode.Box(giode.Styles{Background: colSurface2, BorderRadius: 10, Padding: giode.SymmetricInset(10, 8), Width: -1},
			giode.Text(firstLine(t.Text, 280), giode.Styles{FontFamily: fontText, FontSize: 13, Color: "theme.text", MaxLines: 4}).Selectable("txt-"+t.ID)))
	}
	if !t.Status.Finished() {
		info = append(info, r.u.progressBar(t.Progress()), giode.Text(statusText(t), stSmall))
	} else if t.Status != transfer.Done {
		info = append(info, giode.Text(statusText(t), giode.Styles{FontFamily: fontText, FontSize: 12, Color: "theme.danger"}))
	}
	var actions []giode.Element
	switch {
	case !t.Status.Finished():
		actions = append(actions, r.cancel)
	default:
		if engine.IsPINError(t) {
			if giode.Compact() {
				// On a phone, under the message: beside the chip and icons
				// it left the text a few letters wide.
				info = append(info, giode.Box(giode.Styles{Padding: giode.Inset{Top: 4}}, r.retry))
			} else {
				actions = append(actions, r.retry)
			}
		}
		if t.Text != "" {
			actions = append(actions, r.copy)
		}
		if t.Direction == transfer.Incoming && t.Status == transfer.Done && len(savedFiles(t)) > 0 && t.Text == "" {
			actions = append(actions, r.open)
		}
		actions = append(actions, statusChip(t), r.remove)
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 14, Align: giode.AlignCenter, Padding: giode.UniformInset(14), Background: "theme.surface",
		BorderRadius: 18, BorderWidth: 1, BorderColor: "theme.border", Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{Align: giode.AlignStart}, tile(dirIcon, 44, fg)),
		giode.Stack(giode.Styles{Gap: 5, FlexGrow: 1}, info...),
		giode.Row(giode.Styles{Gap: 4, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, actions...),
	).Layout(gtx)
}
