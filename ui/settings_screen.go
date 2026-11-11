package ui

import (
	"context"
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/scroll"
	"github.com/xmhrnx/giode/components/segmented"
	"github.com/xmhrnx/giode/components/toggle"

	"easysend/internal/settings"
)

// settingsScreen edits the preferences.
type settingsScreen struct {
	u      *UI
	scroll *scroll.Scroll

	report, donate *giode.ButtonElement
	privacy        *giode.ButtonElement

	name   *input.Input
	theme  *segmented.Control
	accent *accentPicker

	folder     *giode.ButtonElement
	openFolder *giode.ButtonElement
	accept     *segmented.Control
	requirePIN *toggle.Switch
	pin        *input.Input
	notify     *toggle.Switch
	clearSent  *toggle.Switch
	background *toggle.Switch

	login, minTray, closeTray, hidden *toggle.Switch

	discoverable, upnp *toggle.Switch
	publicURL          *input.Input
	rendezvous, relay  *input.Input
}

func newSettingsScreen(u *UI) *settingsScreen {
	s := &settingsScreen{u: u, accent: newAccentPicker(u)}
	cur := u.eng.Settings.Get()
	update := func(fn func(*settings.Settings)) {
		if err := u.eng.UpdateSettings(fn); err != nil {
			u.toast(giode.ToastError, "Couldn't save settings: "+err.Error())
		}
	}
	field := func(placeholder, value string, save func(string)) *input.Input {
		in := giode.Input.Text(placeholder, giode.Styles{Background: colSurface2, BorderColor: "theme.border", BorderRadius: 10, Color: "theme.text"})
		in.SetText(value)
		in.Submit(save)
		in.OnBlur(func() { save(in.Text()) })
		return in
	}
	sw := func(on bool, fn func(bool)) *toggle.Switch {
		t := newSwitch("", giode.Styles{}).OnChange(fn)
		t.SetChecked(on)
		return t
	}

	s.name = field("Device name", cur.DeviceName, func(v string) {
		if v != "" && v != u.eng.Settings.Get().DeviceName {
			update(func(st *settings.Settings) { st.DeviceName = v })
			u.toast(giode.ToastSuccess, "Name updated")
		}
	})
	s.theme = giode.Input.Segmented("System", "Dark", "Light").OnChange(func(i int) {
		update(func(st *settings.Settings) { st.Theme = settings.Theme(i) })
		u.ApplyTheme()
	})
	s.theme.SetSelected(int(cur.Theme))
	s.theme.Styles(giode.Styles{Width: -1})

	s.folder = softButton("Change…", "folder", s.pickFolder)
	s.openFolder = softButton("Open", "external", u.openReceivedFolder)
	s.accept = giode.Input.Segmented("Ask", "Favorites", "Everyone").OnChange(func(i int) {
		update(func(st *settings.Settings) { st.AutoAccept = settings.AutoAccept(i) })
	})
	s.accept.SetSelected(int(cur.AutoAccept))
	s.accept.Styles(giode.Styles{Width: -1})
	s.pin = field("PIN", cur.PIN, func(v string) { update(func(st *settings.Settings) { st.PIN = v }) })
	s.pin.Keyboard(giode.KeyboardNumeric)
	s.requirePIN = sw(cur.RequirePIN, func(on bool) { update(func(st *settings.Settings) { st.RequirePIN = on }) })
	s.notify = sw(cur.Notifications, func(on bool) {
		update(func(st *settings.Settings) { st.Notifications = on })
		if on {
			go s.allowNotifications()
		}
	})
	s.clearSent = sw(!cur.KeepAfterSend, func(on bool) { update(func(st *settings.Settings) { st.KeepAfterSend = !on }) })
	s.background = sw(!cur.NoBackground, func(on bool) { update(func(st *settings.Settings) { st.NoBackground = !on }) })

	s.login = sw(u.app.LaunchAtLogin(), func(on bool) {
		if err := u.app.SetLaunchAtLogin(on); err != nil {
			u.toast(giode.ToastError, "Couldn't change start at login: "+err.Error())
			s.login.SetChecked(u.app.LaunchAtLogin())
		}
	})
	s.minTray = sw(cur.MinimizeToTray, func(on bool) {
		update(func(st *settings.Settings) { st.MinimizeToTray = on })
		u.app.SetMinimizeToTray(on)
	})
	s.closeTray = sw(cur.CloseToTray, func(on bool) {
		update(func(st *settings.Settings) { st.CloseToTray = on })
		u.app.SetCloseToTray(on)
	})
	s.hidden = sw(cur.StartHidden, func(on bool) { update(func(st *settings.Settings) { st.StartHidden = on }) })

	s.discoverable = sw(cur.Discoverable, func(on bool) { update(func(st *settings.Settings) { st.Discoverable = on }) })
	s.upnp = sw(cur.UPnP, func(on bool) { update(func(st *settings.Settings) { st.UPnP = on }) })
	s.publicURL = field("https://files.example.com", cur.PublicURL, func(v string) { update(func(st *settings.Settings) { st.PublicURL = v }) })
	s.rendezvous = field("Default (relay.magic-wormhole.io)", cur.RendezvousURL, func(v string) { update(func(st *settings.Settings) { st.RendezvousURL = v }) })
	s.relay = field("Default (transit.magic-wormhole.io)", cur.TransitRelay, func(v string) { update(func(st *settings.Settings) { st.TransitRelay = v }) })

	s.report = softButton("Report…", "report", u.help.Report)
	s.donate = primaryButton("Donate", "heart", u.help.Donate)
	s.privacy = softButton("Privacy policy", "shield", u.openPrivacyPolicy)
	s.scroll = giode.Scroll(s.content, scrollStyles)
	return s
}

func (s *settingsScreen) Layout(gtx layout.Context) layout.Dimensions { return s.scroll.Layout(gtx) }

// allowNotifications asks for the system's permission when notifications
// are turned on, and says how to allow them when the user said no before.
func (s *settingsScreen) allowNotifications() {
	ok, err := s.u.app.RequestNotificationPermission(context.Background())
	if err == nil && !ok {
		s.u.app.Post(func() {
			s.u.toast(giode.ToastInfo, "Notifications are off for EasySend in the system settings")
		})
	}
}
