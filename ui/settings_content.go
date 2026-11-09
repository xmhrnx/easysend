package ui

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"easysend/internal/settings"
	"easysend/ui/platform"
)

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
	saveTo := cur.DownloadDir
	if platform.ReceivesToDownloads {
		saveTo = "Downloads › EasySend"
	}

	var pinRow giode.Element = giode.Box(giode.Styles{Display: giode.DisplayNone})
	if cur.RequirePIN {
		pinRow = settingBlock("PIN", "Tell it to people who send to you.", s.pin, 160)
	}
	body := []giode.Element{
		settingsCard("Profile & look", "sparkles",
			settingBlock("Device name", "How you appear to nearby devices and on browser pages.", s.name, 260),
			settingBlock("Appearance", "", s.theme, 280),
			settingBlock("Accent color", "", s.accent, 240),
		),
		settingsCard("Receiving", "inbox",
			settingBlock("Save files to", saveTo, giode.Row(giode.Styles{Gap: 8}, folderControls...), 220),
			settingBlock("Accept automatically", "Save without asking from favorites (star a device) or from everyone.", s.accept, 300),
			settingRow("Require a PIN", "Senders must type this PIN before you get a request.", s.requirePIN),
			pinRow,
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
			settingBlock("Public address", "Optional: your own domain or tunnel that forwards to port "+strconv.Itoa(self.WebPort)+".", s.publicURL, 260),
			settingRow("Ports", "Nearby: "+strconv.Itoa(self.Port)+" (LocalSend protocol, HTTPS) · Browser: "+strconv.Itoa(self.WebPort), giode.Box(giode.Styles{})),
		),
		settingsCard("Internet transfers", "globe",
			settingBlock("Rendezvous server", "Leave empty for the public Magic Wormhole server.", s.rendezvous, 260),
			settingBlock("Transit relay", "Used only when a direct connection isn't possible.", s.relay, 260),
		),
		s.supportCard(),
		settingsCard("About", "info",
			settingRow("EasySend "+version(), "Nearby transfers speak the open LocalSend protocol; internet transfers speak Magic Wormhole. Both are end-to-end encrypted.", giode.Box(giode.Styles{})),
		),
	)
	return pageFrame("Settings", "Make EasySend yours.", nil, body...)
}
