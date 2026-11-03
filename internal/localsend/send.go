package localsend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"easysend/internal/item"
)

func (n *Node) postJSON(ctx context.Context, u string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := statusError(resp); err != nil {
		return err
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(out)
}

func statusError(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusUnauthorized:
		return ErrPINRequired
	case http.StatusForbidden:
		return ErrDeclined
	case http.StatusConflict, http.StatusTooManyRequests:
		return ErrBusy
	}
	msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Errorf("receiver answered %s %s", resp.Status, bytes.TrimSpace(msg))
}

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

func (n *Node) upload(ctx context.Context, dev Device, sid string, it item.Item, tok string, progress func(int64)) error {
	rc, err := it.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	q := url.Values{"sessionId": {sid}, "fileId": {it.ID}, "token": {tok}}
	body := &progressReader{r: rc, fn: progress}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, dev.url("/upload")+"?"+q.Encode(), body)
	if err != nil {
		return err
	}
	req.ContentLength = it.Size
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return statusError(resp)
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

type progressReader struct {
	r  io.Reader
	fn func(int64)
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 && p.fn != nil {
		p.fn(int64(n))
	}
	return n, err
}

// IsCanceled reports whether err means the transfer was canceled.
func IsCanceled(err error) bool {
	return errors.Is(err, ErrCanceled) || errors.Is(err, context.Canceled)
}
