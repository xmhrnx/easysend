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

// ReceiveInternet connects with a code and saves what was sent; it
// returns the transfer ID.
func (e *Engine) ReceiveInternet(code string) string {
	ctx, cancel := context.WithCancel(context.Background())
	dir := e.Settings.Get().DownloadDir
	id := e.Transfers.Start(transfer.Transfer{Direction: transfer.Incoming, Method: transfer.Internet, Peer: "Wormhole", Code: wormhole.NormalizeCode(code), Status: transfer.Waiting, Folder: dir}, cancel)
	go func() {
		defer cancel()
		of, err := wormhole.Receive(ctx, e.wormholeOptions(), code)
		if err != nil {
			e.finishErr(ctx, id, err)
			return
		}
		if of.IsText {
			text, err := of.ReadText()
			if err != nil {
				e.finishErr(ctx, id, err)
				return
			}
			e.Transfers.Update(id, func(t *transfer.Transfer) {
				t.Text = text
				t.Files = []transfer.File{{Name: "Text message", Size: int64(len(text))}}
				t.Total = int64(len(text))
			})
			e.Transfers.Finish(id, transfer.Done, nil)
			e.notify("Message received", text)
			return
		}
		e.Transfers.Update(id, func(t *transfer.Transfer) {
			t.Files = []transfer.File{{Name: of.Name, Size: of.Size}}
			t.Total = of.Size
			t.Status = transfer.Active
		})
		path, err := of.Save(ctx, dir, func(n int64) { e.Transfers.AddBytes(id, 0, n) })
		if err != nil {
			e.finishErr(ctx, id, err)
			return
		}
		e.Transfers.Update(id, func(t *transfer.Transfer) { t.Files[0].Saved = path })
		e.Transfers.Finish(id, transfer.Done, nil)
		if t, ok := e.Transfers.Get(id); ok {
			e.onReceived(t)
		}
	}()
	return id
}
