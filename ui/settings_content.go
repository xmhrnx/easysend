package ui

import (
	"runtime"

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
		saveTo = tr("Downloads › EasySend")
	}

	receiving := []giode.Element{
		settingBlock(tr("Save files to"), saveTo, giode.Row(giode.Styles{Gap: 8}, folderControls...), 220),
		settingBlock(tr("Accept automatically"), tr("Save without asking from favorites (star a device) or from everyone."), s.accept, 300),
		settingRow(tr("Require a PIN"), tr("Senders must type this PIN before you get a request."), s.requirePIN),
	}
	// Only while a PIN is required: a hidden row would still get a
	// divider from the card.
	if cur.RequirePIN {
		receiving = append(receiving, settingBlock(tr("PIN"), tr("Tell it to people who send to you."), s.pin, 160))
	}
	receiving = append(receiving,
		settingRow(tr("Notifications"), tr("Tell me when something arrives while EasySend isn't open."), s.notify),
	)
	if !platform.IsDesktop {
		hint := tr("Stay reachable after you leave EasySend, with a notification showing it's ready.")
		if runtime.GOOS == "ios" {
			hint = tr("Stay reachable for the short time iOS allows (about 30 seconds) after you leave EasySend.")
		}
		receiving = append(receiving, settingRow(tr("Keep receiving in the background"), hint, s.background))
	}
	body := []giode.Element{
		settingsCard(tr("Profile & look"), "sparkles",
			settingBlock(tr("Language"), "", s.language, 260),
			settingBlock(tr("Device name"), tr("How you appear to nearby devices and on browser pages."), s.name, 260),
			settingBlock(tr("Appearance"), "", s.theme, 280),
			settingBlock(tr("Accent color"), "", s.accent, 240),
		),
		settingsCard(tr("Sending"), "send",
			settingRow(tr("Clear the list after sending"), tr("Once files are delivered, take them off the list so the next send starts fresh."), s.clearSent),
		),
		settingsCard(tr("Receiving"), "inbox", receiving...),
	}
	if platform.IsDesktop {
		body = append(body, settingsCard(tr("Desktop"), "monitor",
			settingRow(tr("Start at login"), tr("Open EasySend quietly in the tray when you sign in."), s.login),
			settingRow(tr("Minimize to tray"), tr("Minimizing hides the window; the tray icon brings it back."), s.minTray),
			settingRow(tr("Close to tray"), tr("Keep receiving when the window is closed. Quit from the tray menu."), s.closeTray),
			settingRow(tr("Always start in the tray"), tr("Start hidden even when you open the app yourself."), s.hidden),
		))
	}
	body = append(body,
		settingsCard(tr("Network"), "wifi",
			settingRow(tr("Visible to nearby devices"), tr("Answer EasySend and LocalSend discovery on this network."), s.discoverable),
			settingRow(tr("Open browser links to the internet"), s.upnpStatus(), s.upnp),
			settingBlock("Public address", tr("Optional: your own domain or tunnel that forwards to port {port}.", "port", self.WebPort), s.publicURL, 260),
			settingRow("Ports", tr("Nearby: {nearby} (LocalSend protocol, HTTPS) · Browser: {browser}", "nearby", self.Port, "browser", self.WebPort), giode.Box(giode.Styles{})),
		),
		settingsCard(tr("Internet transfers"), "globe",
			settingBlock(tr("Rendezvous server"), tr("Leave empty for the public Magic Wormhole server."), s.rendezvous, 260),
			settingBlock(tr("Transit relay"), tr("Used only when a direct connection isn't possible."), s.relay, 260),
		),
		s.supportCard(),
		settingsCard(tr("About"), "info",
			settingRow(tr("EasySend {version}", "version", version()), "Nearby transfers speak the open LocalSend protocol; internet transfers speak Magic Wormhole. Both are end-to-end encrypted.", giode.Box(giode.Styles{})),
			settingRow(tr("Privacy"), tr("No accounts, no ads, no tracking. The developer never sees what you send."), s.privacy),
		),
	)
	return pageFrame(tr("Settings"), tr("Make EasySend yours."), nil, body...)
}
