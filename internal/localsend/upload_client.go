package localsend

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"easysend/internal/item"
)

type progressReader struct {
	r  io.Reader
	fn func(int64)
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

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 && p.fn != nil {
		p.fn(int64(n))
	}
	return n, err
}
