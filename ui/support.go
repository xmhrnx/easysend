package ui

import (
	"context"
	"strconv"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/store"
	"github.com/xmhrnx/giode/support"

	"easysend/internal/diag"
	"easysend/internal/netutil"
)

// Support settings. Reports become GitHub issues through the support
// relay (giode/support/relay, a Cloudflare Worker with "easysend" mapped
// to xmhrnx/easysend in its APPS). With reportURL empty the Report button
// stays hidden.
const (
	reportURL = "https://giode-support.zahiroddinimehran.workers.dev/report"
	reportApp = "easysend"
	donateURL = "https://ko-fi.com/xmhrnx"
	// privacyURL is the privacy policy (PRIVACY.md), linked from Settings
	// and the store listings.
	privacyURL = "https://github.com/xmhrnx/easysend/blob/main/PRIVACY.md"
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
		Logs:      u.networkLog,
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

// networkLog is what "Include recent logs" attaches to a report: this
// device's adapters and addresses now, then what the networking did.
func (u *UI) networkLog() string {
	self := u.eng.Self()
	return "links: " + u.eng.LANBase() + "\nnearby port: " + strconv.Itoa(self.Port) + ", browser port: " + strconv.Itoa(self.WebPort) +
		"\n\n" + netutil.Describe() + "\n" + diag.Lines()
}

// openPrivacyPolicy opens the privacy policy in the browser.
func (u *UI) openPrivacyPolicy() {
	go func() {
		if err := u.app.OpenURL(context.Background(), privacyURL); err != nil {
			u.app.Post(func() { u.toast(giode.ToastError, "Couldn't open the privacy policy: "+err.Error()) })
		}
	}()
}
