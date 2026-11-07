// Package transfer tracks transfers in progress and keeps their history.
package transfer

import (
	"context"
	"time"
)

// Direction says whether a transfer goes out or comes in.
type Direction int

// Method is the channel a transfer uses.
type Method string

// Status is where a transfer stands.
type Status int

// File is one file of a transfer.
type File struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Done  int64  `json:"done"`
	Saved string `json:"saved,omitempty"`
}

// Transfer is one transfer's state.
type Transfer struct {
	ID        string    `json:"id"`
	Direction Direction `json:"direction"`
	Method    Method    `json:"method"`
	Peer      string    `json:"peer"`
	Files     []File    `json:"files"`
	Text      string    `json:"text,omitempty"`
	Total     int64     `json:"total"`
	Done      int64     `json:"done"`
	Status    Status    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Code      string    `json:"code,omitempty"`
	Started   time.Time `json:"started"`
	Finished  time.Time `json:"finished,omitempty"`
	Folder    string    `json:"folder,omitempty"`
	// Speed in bytes per second, smoothed; not persisted.
	Speed float64 `json:"-"`

	cancel   context.CancelFunc
	lastTick time.Time
	lastDone int64
}

const (
	// Outgoing transfers send from this device.
	Outgoing Direction = iota
	// Incoming transfers arrive at this device.
	Incoming
)

const (
	// Nearby is the LocalSend-compatible local network protocol.
	Nearby Method = "nearby"
	// Internet is Magic Wormhole.
	Internet Method = "internet"
	// Browser is the built-in web page.
	Browser Method = "browser"
)

const (
	// Waiting means the peer hasn't accepted or connected yet.
	Waiting Status = iota
	// Active means bytes are flowing.
	Active
	// Done means it finished successfully.
	Done
	// Failed means it stopped with an error.
	Failed
	// Canceled means a user stopped it.
	Canceled
	// Declined means the receiver refused it.
	Declined
)

// Finished reports whether the status is final.
func (s Status) Finished() bool { return s >= Done }
