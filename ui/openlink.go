package ui

import (
	"net/url"
	"strings"

	"github.com/xmhrnx/giode"

	"easysend/internal/item"
)

// OpenLink handles a link that opened the app: a scanned internet code
// (wormhole-transfer:7-crossover-clockwork) starts receiving it.
func (u *UI) OpenLink(link *url.URL) {
	if link.Scheme != "wormhole-transfer" {
		return
	}
	code := link.Opaque
	if code == "" {
		code = strings.TrimPrefix(link.Host+link.Path, "/")
	}
	if code == "" {
		return
	}
	u.eng.ReceiveInternet(code)
	u.current = screenActivity
	u.app.ShowWindow()
	u.toast(giode.ToastInfo, tr("Receiving with code {code}…", "code", code))
}

// ReceiveShare adds what another app shared through the system share
// sheet to the selection, and shows the Send screen.
func (u *UI) ReceiveShare(s giode.SharedContent) {
	u.current = screenSend
	if s.Text != "" {
		u.items = append(u.items, item.FromText(s.Text))
	}
	if len(s.Files) == 0 {
		return
	}
	go func() {
		var added []item.Item
		for _, f := range s.Files {
			it, err := pickedItem(f)
			f.Close()
			if err != nil {
				name := f.Name
				u.app.Post(func() {
					u.toast(giode.ToastError, tr("Couldn't read {name}: {error}", "name", name, "error", err.Error()))
				})
				continue
			}
			added = append(added, it)
		}
		u.app.Post(func() { u.addItems(added) })
	}()
}

// copy puts text on the clipboard.
func (u *UI) copy(text, what string) {
	u.app.SetClipboard(text)
	u.toast(giode.ToastSuccess, tr("{what} copied", "what", what))
}
