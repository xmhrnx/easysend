package localsend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// IsCanceled reports whether err means the transfer was canceled.
func IsCanceled(err error) bool {
	return errors.Is(err, ErrCanceled) || errors.Is(err, context.Canceled)
}
