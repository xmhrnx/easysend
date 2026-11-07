package engine

import (
	"time"

	"easysend/internal/item"
	"easysend/internal/webshare"
)

// ShareOptions configure a browser link.
type ShareOptions struct {
	Expires      time.Duration // 0: never
	MaxDownloads int
	PIN          string
}

// ShareView is a share with its links.
type ShareView struct {
	webshare.Share
	LANURL    string
	PublicURL string
	// OtherURLs are the link on this device's other local addresses.
	OtherURLs []string
}

// CreateShare publishes items as a browser link.
func (e *Engine) CreateShare(items []item.Item, o ShareOptions) ShareView {
	sh := webshare.Share{Items: items, MaxDownloads: o.MaxDownloads, PIN: o.PIN}
	if o.Expires > 0 {
		sh.Expires = time.Now().Add(o.Expires)
	}
	sh = e.web.Add(sh)
	e.changed()
	return e.view(sh)
}

func (e *Engine) view(sh webshare.Share) ShareView {
	v := ShareView{Share: sh, LANURL: e.LANBase() + "/s/" + sh.Token}
	for _, b := range e.OtherLANBases() {
		v.OtherURLs = append(v.OtherURLs, b+"/s/"+sh.Token)
	}
	if pub := e.PublicBase(); pub != "" {
		v.PublicURL = pub + "/s/" + sh.Token
	}
	return v
}

// Shares lists the browser links, newest first.
func (e *Engine) Shares() []ShareView {
	list := e.web.Shares()
	out := make([]ShareView, len(list))
	for i, sh := range list {
		out[i] = e.view(sh)
	}
	return out
}

// StopShare ends a link.
func (e *Engine) StopShare(token string) {
	e.web.Remove(token)
	e.changed()
}

// DropOtherURLs is the upload link on this device's other local addresses.
func (e *Engine) DropOtherURLs() []string {
	var out []string
	for _, b := range e.OtherLANBases() {
		out = append(out, b+"/r/"+e.Settings.Get().DropToken)
	}
	return out
}

// DropURLs returns the links where browsers upload to this device.
func (e *Engine) DropURLs() (lan, public string) {
	tok := e.Settings.Get().DropToken
	lan = e.LANBase() + "/r/" + tok
	if pub := e.PublicBase(); pub != "" {
		public = pub + "/r/" + tok
	}
	return lan, public
}
