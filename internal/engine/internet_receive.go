package engine

import (
	"context"
	"github.com/xmhrnx/giode/i18n"

	"easysend/internal/transfer"
	"easysend/internal/wormhole"
)

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
				t.Files = []transfer.File{{Name: i18n.Intl("Text message"), Size: int64(len(text))}}
				t.Total = int64(len(text))
			})
			e.Transfers.Finish(id, transfer.Done, nil)
			e.notify(i18n.Intl("Message received"), text)
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
