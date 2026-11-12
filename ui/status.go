package ui

import (
	"github.com/xmhrnx/giode"

	"easysend/internal/item"
	"easysend/internal/transfer"
)

// statusText describes a transfer's state in a few words.
func statusText(t transfer.Transfer) string {
	switch t.Status {
	case transfer.Waiting:
		if t.Method == transfer.Internet && t.Direction == transfer.Outgoing {
			if t.Code == "" {
				return tr("Creating code…")
			}
			return tr("Waiting for the receiver…")
		}
		if t.Direction == transfer.Outgoing {
			return tr("Waiting for {name} to accept…", "name", t.Peer)
		}
		return tr("Connecting…")
	case transfer.Active:
		s := tr("{done} of {total}", "done", item.HumanSize(t.Done), "total", item.HumanSize(t.Total))
		if t.Speed > 0 {
			s += " · " + item.HumanSize(int64(t.Speed)) + "/s"
			if eta := t.ETA(); eta > 0 {
				s += " · " + tr("{time} left", "time", item.HumanDuration(eta))
			}
		}
		return s
	case transfer.Done:
		if t.Direction == transfer.Incoming {
			return tr("Received · {size}", "size", item.HumanSize(t.Total))
		}
		return tr("Sent · {size}", "size", item.HumanSize(t.Total))
	case transfer.Canceled:
		return tr("Canceled")
	case transfer.Declined:
		return tr("Declined")
	}
	if t.Error != "" {
		return tr(t.Error)
	}
	return tr("Failed")
}

// statusChip is a colored label for a finished transfer.
func statusChip(t transfer.Transfer) giode.Element {
	switch t.Status {
	case transfer.Done:
		return chip(tr("Done"), "theme.success")
	case transfer.Canceled:
		return chip(tr("Canceled"), "theme.muted")
	case transfer.Declined:
		return chip(tr("Declined"), "theme.warning")
	case transfer.Failed:
		return chip(tr("Failed"), "theme.danger")
	}
	return chip(tr("In progress"), "theme.primary")
}

// progressBar is a slim accent line.
func (u *UI) progressBar(v float32) giode.Element {
	if v < 0 {
		v = 0
	}
	return giode.Box(giode.Styles{Width: -1, Height: 4, BorderRadius: 2, Background: colSurface3, Overflow: giode.OverflowHidden},
		giode.Box(giode.Styles{WidthPct: pct(v), Height: 4, BorderRadius: 2, Background: "theme.primary"}))
}
