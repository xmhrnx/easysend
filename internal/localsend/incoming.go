package localsend

import (
	"context"
	"io"
	"time"
)

type session struct {
	id      string
	sender  string
	tokens  map[string]string
	pending map[string]bool
	sink    Sink
	ctx     context.Context
	cancel  context.CancelFunc
	last    time.Time
}

// Incoming is a transfer offered by a peer.
type Incoming struct {
	Sender Device
	Files  []FileMeta
	// Cancel stops the transfer from this side.
	Cancel func()
}

// Sink receives the files of an accepted transfer.
type Sink interface {
	// Write stores the content of file id, read from r.
	Write(ctx context.Context, id string, r io.Reader) error
	// Close ends the transfer: err is nil when every file arrived.
	Close(err error)
}
