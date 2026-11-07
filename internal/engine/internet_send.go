package engine

import (
	"context"
	"errors"

	"easysend/internal/item"
	"easysend/internal/transfer"
	"easysend/internal/wormhole"
)

func (e *Engine) wormholeOptions() wormhole.Options {
	s := e.Settings.Get()
	return wormhole.Options{RendezvousURL: s.RendezvousURL, TransitRelay: s.TransitRelay}
}

// SendInternet offers items over the internet and returns the transfer
// ID; the code to share appears on the transfer (Code) once the
// rendezvous server gave one.
func (e *Engine) SendInternet(items []item.Item) string {
	ctx, cancel := context.WithCancel(context.Background())
	files := make([]transfer.File, len(items))
	for i, it := range items {
		files[i] = transfer.File{Name: it.Name, Size: it.Size}
	}
	t := transfer.Transfer{Direction: transfer.Outgoing, Method: transfer.Internet, Peer: "Wormhole", Files: files, Status: transfer.Waiting}
	if len(items) == 1 && items[0].Kind == item.Text {
		t.Text = items[0].Text
	}
	id := e.Transfers.Start(t, cancel)
	go func() {
		defer cancel()
		var last int64
		code, res, err := wormhole.Send(ctx, e.wormholeOptions(), items, func(sent, total int64) {
			e.Transfers.Update(id, func(t *transfer.Transfer) { t.Total = total })
			e.Transfers.AddBytes(id, -1, sent-last)
			last = sent
		})
		if err != nil {
			e.finishErr(ctx, id, err)
			return
		}
		e.Transfers.Update(id, func(t *transfer.Transfer) { t.Code = code })
		if err := <-res; err != nil {
			e.finishErr(ctx, id, err)
			return
		}
		e.Transfers.Finish(id, transfer.Done, nil)
	}()
	return id
}

func (e *Engine) finishErr(ctx context.Context, id string, err error) {
	if ctx.Err() != nil || errors.Is(err, context.Canceled) {
		e.Transfers.Finish(id, transfer.Canceled, nil)
		return
	}
	e.Transfers.Finish(id, transfer.Failed, err)
}
