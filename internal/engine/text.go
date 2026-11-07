package engine

import (
	"easysend/internal/localsend"
	"easysend/internal/transfer"
)

func (e *Engine) recordText(method transfer.Method, from, text string) {
	id := e.Transfers.Start(transfer.Transfer{Direction: transfer.Incoming, Method: method, Peer: from, Text: text,
		Files: []transfer.File{{Name: "Text message", Size: int64(len(text))}}, Status: transfer.Active}, nil)
	e.Transfers.Finish(id, transfer.Done, nil)
	preview := text
	if len(preview) > 120 {
		preview = preview[:120] + "…"
	}
	e.notify("Message from "+from, preview)
}

func (e *Engine) onNearbyText(from localsend.Device, text string) {
	e.recordText(transfer.Nearby, from.Alias, text)
}

func (e *Engine) onBrowserText(from, text string) {
	e.recordText(transfer.Browser, from, text)
}
