package ui

import (
	"context"
	"errors"
	"strconv"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/scroll"
	"github.com/xmhrnx/giode/components/segmented"
	"github.com/xmhrnx/giode/components/toggle"

	"easysend/internal/engine"
	"easysend/internal/settings"
	"easysend/ui/platform"
)

// settingsScreen edits the preferences.
type settingsScreen struct {
	u      *UI
	scroll *scroll.Scroll

	name   *input.Input
	theme  *segmented.Control
	accent *accentPicker

	folder     *giode.ButtonElement
	openFolder *giode.ButtonElement
	accept     *segmented.Control
	requirePIN *toggle.Switch
	pin        *input.Input
	notify     *toggle.Switch

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
	field := func(placeholder, value string, w int, save func(string)) *input.Input {
		in := giode.Input.Text(placeholder, giode.Styles{Width: dpW(w), Background: colSurface2, BorderColor: "theme.border", BorderRadius: 10, Color: "theme.text"})
		in.SetText(value)
		in.Submit(save)
		in.OnBlur(func() { save(in.Text()) })
		return in
	}
	sw := func(on bool, fn func(bool)) *toggle.Switch {
		t := giode.Input.Switch("").OnChange(fn)
		t.SetChecked(on)
		return t
	}

	s.name = field("Device name", cur.DeviceName, 240, func(v string) {
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

	s.folder = softButton("Change…", "folder", s.pickFolder)
	s.openFolder = softButton("Open", "external", func() { u.reveal(u.eng.Settings.Get().DownloadDir) })
	s.accept = giode.Input.Segmented("Ask", "Favorites", "Everyone").OnChange(func(i int) {
		update(func(st *settings.Settings) { st.AutoAccept = settings.AutoAccept(i) })
	})
	s.accept.SetSelected(int(cur.AutoAccept))
	s.pin = field("PIN", cur.PIN, 110, func(v string) { update(func(st *settings.Settings) { st.PIN = v }) })
	s.pin.Keyboard(giode.KeyboardNumeric)
	s.requirePIN = sw(cur.RequirePIN, func(on bool) { update(func(st *settings.Settings) { st.RequirePIN = on }) })
	s.notify = sw(cur.Notifications, func(on bool) { update(func(st *settings.Settings) { st.Notifications = on }) })

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
	s.publicURL = field("https://files.example.com", cur.PublicURL, 260, func(v string) { update(func(st *settings.Settings) { st.PublicURL = v }) })
	s.rendezvous = field("Default (relay.magic-wormhole.io)", cur.RendezvousURL, 260, func(v string) { update(func(st *settings.Settings) { st.RendezvousURL = v }) })
	s.relay = field("Default (transit.magic-wormhole.io)", cur.TransitRelay, 260, func(v string) { update(func(st *settings.Settings) { st.TransitRelay = v }) })

	s.scroll = giode.Scroll(s.content, scrollStyles)
	return s
}

func (s *settingsScreen) pickFolder() {
	go func() {
		dir, err := s.u.app.PickFolder(context.Background(), giode.FolderOptions{Title: "Where should received files go?", Directory: s.u.eng.Settings.Get().DownloadDir})
		if errors.Is(err, giode.ErrCanceled) {
			return
		}
		s.u.app.Post(func() {
			if err != nil {
				s.u.toast(giode.ToastError, "Couldn't choose the folder: "+err.Error())
				return
			}
			_ = s.u.eng.UpdateSettings(func(st *settings.Settings) { st.DownloadDir = dir })
			s.u.toast(giode.ToastSuccess, "Received files now go to "+dir)
		})
	}()
}

func (s *settingsScreen) syncSwitches(cur settings.Settings) {
	s.discoverable.SetChecked(cur.Discoverable)
	s.login.SetChecked(s.u.app.LaunchAtLogin())
}

func (s *settingsScreen) content() giode.Element {
	cur := s.u.eng.Settings.Get()
	s.syncSwitches(cur)
	self := s.u.eng.Self()

	folderControls := []giode.Element{s.openFolder}
	if platform.IsDesktop {
		folderControls = append(folderControls, s.folder)
	}
	pinControls := []giode.Element{s.requirePIN}
	if cur.RequirePIN {
		pinControls = append([]giode.Element{s.pin}, pinControls...)
	}

	body := []giode.Element{
		settingsCard("Profile & look", "sparkles",
			settingRow("Device name", "How you appear to nearby devices and on browser pages.", s.name),
			settingRow("Appearance", "", giode.Box(giode.Styles{Width: 280}, s.theme)),
			settingRow("Accent color", "", s.accent),
		),
		settingsCard("Receiving", "inbox",
			settingRow("Save files to", cur.DownloadDir, giode.Row(giode.Styles{Gap: 8}, folderControls...)),
			settingRow("Accept automatically", "Save without asking from favorites (star a device) or from everyone.", giode.Box(giode.Styles{Width: 300}, s.accept)),
			settingRow("Require a PIN", "Senders must type this PIN before you get a request.", giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter}, pinControls...)),
			settingRow("Notifications", "Tell me when something arrives.", s.notify),
		),
	}
	if platform.IsDesktop {
		body = append(body, settingsCard("Desktop", "monitor",
			settingRow("Start at login", "Open EasySend quietly in the tray when you sign in.", s.login),
			settingRow("Minimize to tray", "Minimizing hides the window; the tray icon brings it back.", s.minTray),
			settingRow("Close to tray", "Keep receiving when the window is closed. Quit from the tray menu.", s.closeTray),
			settingRow("Always start in the tray", "Start hidden even when you open the app yourself.", s.hidden),
		))
	}
	body = append(body,
		settingsCard("Network", "wifi",
			settingRow("Visible to nearby devices", "Answer EasySend and LocalSend discovery on this network.", s.discoverable),
			settingRow("Open browser links to the internet", s.upnpStatus(), s.upnp),
			settingRow("Public address", "Optional: your own domain or tunnel that forwards to port "+strconv.Itoa(self.WebPort)+".", s.publicURL),
			settingRow("Ports", "Nearby: "+strconv.Itoa(self.Port)+" (LocalSend protocol, HTTPS) · Browser: "+strconv.Itoa(self.WebPort), giode.Box(giode.Styles{})),
		),
		settingsCard("Internet transfers", "globe",
			settingRow("Rendezvous server", "Leave empty for the public Magic Wormhole server.", s.rendezvous),
			settingRow("Transit relay", "Used only when a direct connection isn't possible.", s.relay),
		),
		settingsCard("About", "info",
			settingRow("EasySend "+version(), "Nearby transfers speak the open LocalSend protocol; internet transfers speak Magic Wormhole. Both are end-to-end encrypted.", giode.Box(giode.Styles{})),
		),
	)
	return pageFrame("Settings", "Make EasySend yours.", nil, body...)
}

func (s *settingsScreen) upnpStatus() string {
	st := s.u.eng.Public()
	switch st.State {
	case engine.PublicReady:
		return "Working: reachable at " + st.IP + ":" + strconv.Itoa(st.Port) + " (UPnP)."
	case engine.PublicWorking:
		return "Asking your router…"
	case engine.PublicUnavailable:
		return st.Err
	}
	return "Asks your router (UPnP) to open a port so links work from anywhere."
}

func version() string {
	if giode.AppVersion != "" {
		return giode.AppVersion
	}
	return "1.0.0"
}

func (s *settingsScreen) Layout(gtx layout.Context) layout.Dimensions { return s.scroll.Layout(gtx) }
