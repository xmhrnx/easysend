// EasySend sends files, folders and text between devices: nearby over the
// local network (LocalSend protocol), anywhere over the internet (Magic
// Wormhole), and to anyone with a browser through a link or QR code.
package main

import (
	"image"
	"log"

	"github.com/xmhrnx/giode"

	"easysend/assets"
	"easysend/internal/engine"
	"easysend/internal/settings"
	"easysend/ui"
	"easysend/ui/platform"
)

func main() {
	if err := ui.LoadFonts(); err != nil {
		log.Fatal(err)
	}
	if err := ui.LoadIcons(); err != nil {
		log.Fatal(err)
	}
	prefs, err := settings.Open(platform.DataDir(), platform.DownloadDir())
	if err != nil {
		log.Printf("settings: %v", err)
	}
	s := prefs.Get()
	startInTray := giode.StartInTrayAtLogin
	if s.StartHidden {
		startInTray = giode.StartInTrayAlways
	}

	// The services start in OnStartup, which only the first running
	// instance reaches; a second launch hands its files to it and exits.
	stop := make(chan struct{})
	var (
		app *giode.App
		eng *engine.Engine
		u   *ui.UI
	)
	app = giode.New(giode.Config{
		Title:          "EasySend",
		Size:           image.Pt(1180, 800),
		MinSize:        image.Pt(380, 560),
		Resizable:      true,
		Center:         true,
		Icon:           assets.Icon,
		Background:     "theme.background",
		StatusBar:      "theme.background",
		NavigationBar:  "theme.surface",
		MinimizeToTray: s.MinimizeToTray,
		CloseToTray:    s.CloseToTray,
		StartInTray:    startInTray,
		SingleInstanceLock: &giode.SingleInstanceLock{
			UniqueID:     "app.easysend" + platform.Profile(),
			OnLaunchArgs: func(args []string) { u.AddPaths(args) },
		},
		ContextActions: []giode.ContextAction{
			{ID: "send", Label: "Send with EasySend", Files: true, Folders: true},
		},
		OnContextAction:     func(_ string, paths []string) { u.AddPaths(paths) },
		OnFileDrop:          func(paths []string) { u.AddPaths(paths) },
		OnSystemThemeChange: func(bool) { u.ApplyTheme() },
		OnStartup: func() {
			eng, err = engine.Start(engine.Config{
				DataDir:     platform.DataDir(),
				DownloadDir: platform.DownloadDir(),
				DeviceModel: deviceModel(),
				DeviceType:  platform.DeviceType,
				OnChange:    app.Invalidate,
				Notify: func(title, body string) {
					if err := app.Notify(title, body); err != nil {
						log.Printf("notify: %v", err)
					}
				},
			})
			if err != nil {
				log.Fatal(err)
			}
			u = ui.New(app, eng)
			u.SetupTray()
			go u.KeepAliveWhileTransferring(stop)
			if platform.IsDesktop && platform.Profile() == "" {
				if err := app.InstallContextActions(); err != nil {
					log.Printf("context menu: %v", err)
				}
			}
		},
		OnShutdown: func() {
			close(stop)
			if eng != nil {
				eng.Close()
			}
		},
	})
	if err := app.Run(func() giode.Element { return u.View() }); err != nil {
		log.Fatal(err)
	}
}

// deviceModel names the hardware for other devices' lists.
func deviceModel() string {
	d := giode.Device()
	if d.Model != "" {
		return d.Model
	}
	switch d.OS {
	case "darwin":
		return "Mac"
	case "windows":
		return "Windows PC"
	case "linux":
		return "Linux PC"
	case "android":
		return "Android"
	case "ios":
		return "iPhone"
	}
	return "Computer"
}
