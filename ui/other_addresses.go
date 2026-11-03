package ui

import (
	"strings"

	"github.com/xmhrnx/giode"
)

// otherAddresses lists a link on the device's other local addresses, for
// when the first one doesn't open (it belongs to an adapter the other
// device can't reach). It's empty when there are none.
func otherAddresses(urls []string, key string) giode.Element {
	if len(urls) == 0 {
		return giode.Box(giode.Styles{})
	}
	return giode.Stack(giode.Styles{Gap: 2},
		giode.Text("Not opening? Try:", stSmall),
		giode.Text(strings.Join(urls, "\n"), giode.Styles{FontFamily: fontText, FontSize: 12, Color: "theme.muted"}).Selectable(key),
	)
}
