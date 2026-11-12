package engine

import (
	"context"
	"github.com/xmhrnx/giode/i18n"
	"time"

	"easysend/internal/item"
	"easysend/internal/localsend"
	"easysend/internal/transfer"
)

func (e *Engine) recordText(method transfer.Method, from, text string) {
	id := e.Transfers.Start(transfer.Transfer{Direction: transfer.Incoming, Method: method, Peer: from, Text: text,
		Files: []transfer.File{{Name: i18n.Intl("Text message"), Size: int64(len(text))}}, Status: transfer.Active}, nil)
	e.Transfers.Finish(id, transfer.Done, nil)
	e.notify(i18n.Intl("Message from {name}", "name", from), preview(text))
}

func preview(text string) string {
	if r := []rune(text); len(r) > 120 {
		return string(r[:120]) + "…"
	}
	return text
}

// onNearbyText asks before taking a text message, as for files.
func (e *Engine) onNearbyText(ctx context.Context, from localsend.Device, text string) bool {
	req := &Request{ID: item.NewID(), Sender: from, Text: text, Total: int64(len(text)), Received: time.Now(), answer: make(chan bool, 1)}
	if !e.ask(ctx, req, i18n.Intl("{name} wants to send you a message", "name", from.Alias), preview(text)) {
		return false
	}
	e.recordText(transfer.Nearby, from.Alias, text)
	return true
}

func (e *Engine) onBrowserText(from, text string) {
	e.recordText(transfer.Browser, from, text)
}
