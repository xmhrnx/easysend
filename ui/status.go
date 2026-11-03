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
				return "Creating code…"
			}
			return "Waiting for the receiver…"
		}
		if t.Direction == transfer.Outgoing {
			return "Waiting for " + t.Peer + " to accept…"
		}
		return "Connecting…"
	case transfer.Active:
		s := item.HumanSize(t.Done) + " of " + item.HumanSize(t.Total)
		if t.Speed > 0 {
			s += " · " + item.HumanSize(int64(t.Speed)) + "/s"
			if eta := t.ETA(); eta > 0 {
				s += " · " + item.HumanDuration(eta) + " left"
			}
		}
		return s
	case transfer.Done:
		if t.Direction == transfer.Incoming {
			return "Received · " + item.HumanSize(t.Total)
		}
		return "Sent · " + item.HumanSize(t.Total)
	case transfer.Canceled:
		return "Canceled"
	case transfer.Declined:
		return "Declined"
	}
	if t.Error != "" {
		return t.Error
	}
	return "Failed"
}

// statusChip is a colored label for a finished transfer.
func statusChip(t transfer.Transfer) giode.Element {
	switch t.Status {
	case transfer.Done:
		return chip("Done", "theme.success")
	case transfer.Canceled:
		return chip("Canceled", "theme.muted")
	case transfer.Declined:
		return chip("Declined", "theme.warning")
	case transfer.Failed:
		return chip("Failed", "theme.danger")
	}
	return chip("In progress", "theme.primary")
}

// progressBar is a slim accent line.
func (u *UI) progressBar(v float32) giode.Element {
	if v < 0 {
		v = 0
	}
	return giode.Box(giode.Styles{Width: -1, Height: 4, BorderRadius: 2, Background: colSurface3, Overflow: giode.OverflowHidden},
		giode.Box(giode.Styles{WidthPct: pct(v), Height: 4, BorderRadius: 2, Background: "theme.primary"}))
}
