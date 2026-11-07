// Package wormhole sends and receives over the internet with the Magic
// Wormhole protocol: a short code pairs two devices anywhere, the
// transfer is end-to-end encrypted and goes directly between them when
// the networks allow, through a relay otherwise. It interoperates with
// the wormhole command line tool and the other Magic Wormhole apps.
package wormhole

import (
	"strings"

	ww "github.com/psanford/wormhole-william/wormhole"
)

// Options select the servers; empty fields use the public defaults.
type Options struct {
	RendezvousURL string
	TransitRelay  string
}

// Offer is what a sender offered.
type Offer struct {
	Name  string
	Size  int64
	Files int
	// IsText is set for a text message.
	IsText bool
	// IsFolder is set for several files.
	IsFolder bool
	msg      *ww.IncomingMessage
}

func (o Options) client() *ww.Client {
	return &ww.Client{
		AppID:                     ww.WormholeCLIAppID,
		RendezvousURL:             o.RendezvousURL,
		TransitRelayAddress:       o.TransitRelay,
		PassPhraseComponentLength: 2,
	}
}

// NormalizeCode accepts codes typed with spaces or pasted as a
// wormhole-transfer: link.
func NormalizeCode(code string) string {
	code = strings.TrimSpace(code)
	code = strings.TrimPrefix(code, "wormhole-transfer:")
	code = strings.Join(strings.Fields(code), "-")
	return strings.ToLower(code)
}
