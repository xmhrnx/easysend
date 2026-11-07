package localsend

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"easysend/internal/item"
)

// SendEvents reports a send's progress.
type SendEvents struct {
	// Accepted runs once the receiver accepted.
	Accepted func()
	// Progress reports n more bytes of items[i].
	Progress func(i int, n int64)
}

// Send offers items to dev and uploads them once accepted. pin is sent
// when the receiver asks for one (ErrPINRequired otherwise).
func (n *Node) Send(ctx context.Context, dev Device, items []item.Item, pin string, ev SendEvents) error {
	req := prepareRequest{Info: n.Self(), Files: map[string]FileMeta{}}
	for _, it := range items {
		fm := FileMeta{ID: it.ID, FileName: it.Name, Size: it.Size, FileType: it.MIME}
		if it.Kind == item.Text {
			fm.Preview = it.Text
		}
		req.Files[it.ID] = fm
	}
	u := dev.url("/prepare-upload")
	if pin != "" {
		u += "?pin=" + url.QueryEscape(pin)
	}
	// The receiver may take a while to answer: the user decides.
	var resp prepareResponse
	if err := n.postJSON(ctx, u, req, &resp); err != nil {
		return err
	}
	if ev.Accepted != nil {
		ev.Accepted()
	}
	if resp.SessionID == "" {
		// 204: delivered with the offer (a text message).
		if ev.Progress != nil {
			for i, it := range items {
				ev.Progress(i, it.Size)
			}
		}
		return nil
	}
	for i, it := range items {
		tok, ok := resp.Files[it.ID]
		if !ok {
			continue // the receiver skipped this file
		}
		if err := n.upload(ctx, dev, resp.SessionID, it, tok, func(k int64) {
			if ev.Progress != nil {
				ev.Progress(i, k)
			}
		}); err != nil {
			if ctx.Err() != nil {
				n.cancelRemote(dev, resp.SessionID)
				return ErrCanceled
			}
			return err
		}
	}
	return nil
}

func (n *Node) cancelRemote(dev Device, sid string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, dev.url("/cancel")+"?sessionId="+url.QueryEscape(sid), nil)
	if err != nil {
		return
	}
	if resp, err := n.client.Do(req); err == nil {
		resp.Body.Close()
	}
}
