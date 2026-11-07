package wormhole

import (
	"context"
	"io"

	ww "github.com/psanford/wormhole-william/wormhole"
)

// Receive connects with code and returns the offer; Save or Reject it.
func Receive(ctx context.Context, o Options, code string) (*Offer, error) {
	code = NormalizeCode(code)
	msg, err := o.client().Receive(ctx, code)
	if err != nil {
		return nil, err
	}
	of := &Offer{Name: msg.Name, Size: msg.UncompressedBytes64, Files: msg.FileCount, msg: msg}
	switch msg.Type {
	case ww.TransferText:
		of.IsText = true
	case ww.TransferDirectory:
		of.IsFolder = true
	default:
		of.Files = 1
	}
	return of, nil
}

// Reject declines the offer.
func (o *Offer) Reject() error { return o.msg.Reject() }

// ReadText returns a text message.
func (o *Offer) ReadText() (string, error) {
	b, err := io.ReadAll(io.LimitReader(o.msg, 16<<20))
	return string(b), err
}
