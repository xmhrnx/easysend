package ui

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/store"
	"github.com/xmhrnx/giode/support"
)

// Support settings. Reports become GitHub issues through the support
// relay (giode/support/relay, a Cloudflare Worker with "easysend" mapped
// to xmhrnx/easysend in its APPS); leave reportURL empty until it's
// deployed and the Report button stays hidden.
const (
	reportURL = ""
	reportApp = "easysend"
	donateURL = "https://ko-fi.com/xmhrnx"
)

// tipProducts are the consumable in-app products offered as tips where
// the store requires its own payments (App Store, Mac App Store, Google
// Play). Create them with these ids in App Store Connect and the Play
// Console.
var tipProducts = []string{"tip_coffee", "tip_lunch", "tip_dinner"}

// newSupport sets up the report form and donations.
func newSupport(u *UI) *support.Support {
	return support.New(u.app, support.Options{
		ReportURL: reportURL,
		App:       reportApp,
		Name:      "EasySend",
		DonateURL: donateURL,
		Tips:      tipProducts,
		Toasts:    u.toasts,
	})
}

// canReport reports whether problem reports can be sent from this build.
func canReport() bool { return reportURL != "" }

// donateText describes donating the way this install takes donations.
func donateText() string {
	if store.BillingRequired() {
		return "EasySend is free. If it helps you, leave a tip — it unlocks nothing, it just says thanks."
	}
	return "EasySend is free. If it helps you, buy me a coffee on Ko-fi."
}

// supportCard is the Help & support section of Settings.
func (s *settingsScreen) supportCard() giode.Element {
	var rows []giode.Element
	if canReport() {
		rows = append(rows, settingRow("Report a problem", "Something not working? Tell us — no account needed.", s.report))
	}
	if s.u.help.CanDonate() {
		rows = append(rows, settingRow("Support EasySend", donateText(), s.donate))
	}
	if len(rows) == 0 {
		return giode.Box(giode.Styles{Display: giode.DisplayNone})
	}
	return settingsCard("Help & support", "heart", rows...)
}
