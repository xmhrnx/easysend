package ui

import (
	"strings"

	"github.com/xmhrnx/giode"

	"easysend/ui/platform"
)

// linkHelp is what to try when a link doesn't open on the other device:
// the link on this device's other local addresses, and on desktops the
// firewall, which blocks phones until EasySend is allowed through it.
func linkHelp(urls []string, key string) giode.Element {
	var kids []giode.Element
	if len(urls) > 0 {
		kids = append(kids,
			giode.Text("Not opening? Try:", stSmall),
			giode.Text(strings.Join(urls, "\n"), giode.Styles{FontFamily: fontText, FontSize: 12, Color: "theme.muted"}).Selectable(key))
	}
	if platform.IsDesktop {
		kids = append(kids, giode.Text("Still not opening? Allow EasySend in this computer's firewall for this network.", stSmall))
	}
	return giode.Stack(giode.Styles{Gap: 2}, kids...)
}
