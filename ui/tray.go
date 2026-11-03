package ui

import (
	"runtime"

	"github.com/xmhrnx/giode"

	"easysend/assets"
	"easysend/internal/settings"
)

// SetupTray puts EasySend in the system tray (menu bar on macOS).
func (u *UI) SetupTray() {
	if !giode.TraySupported() {
		return
	}
	t := u.app.Tray().SetTooltip("EasySend")
	if runtime.GOOS == "darwin" {
		t.SetTemplateIcon(assets.TrayTemplate)
	} else {
		t.SetIcon(assets.Tray)
	}
	u.refreshTrayMenu()
}

// refreshTrayMenu rebuilds the tray menu (check marks follow settings).
func (u *UI) refreshTrayMenu() {
	if !giode.TraySupported() {
		return
	}
	s := u.eng.Settings.Get()
	show := func(sc screen) func() {
		return func() {
			u.Go(sc)
			u.app.ShowWindow()
		}
	}
	u.app.Tray().SetMenu(
		giode.TrayAction("Open EasySend", u.app.ShowWindow),
		giode.TraySeparator(),
		giode.TrayAction("Send files…", func() {
			show(screenSend)()
			u.pickFiles()
		}),
		giode.TrayAction("Send clipboard text", func() {
			show(screenSend)()
			u.pasteText()
		}),
		giode.TrayAction("Receive…", show(screenReceive)),
		giode.TrayAction("Activity", show(screenActivity)),
		giode.TraySeparator(),
		giode.TrayCheckbox("Visible to nearby devices", s.Discoverable, func() {
			_ = u.eng.UpdateSettings(func(st *settings.Settings) { st.Discoverable = !st.Discoverable })
			u.refreshTrayMenu()
		}),
		giode.TrayCheckbox("Receive from browsers", s.BrowserDrop, func() {
			_ = u.eng.UpdateSettings(func(st *settings.Settings) { st.BrowserDrop = !st.BrowserDrop })
			u.refreshTrayMenu()
		}),
		giode.TraySeparator(),
		giode.TrayAction("Quit EasySend", u.app.Quit),
	)
}
