// EasySend sends files, folders and text between devices: nearby over the
// local network (LocalSend protocol), anywhere over the internet (Magic
// Wormhole), and to anyone with a browser through a link or QR code.
package main

import (
	"image"
	"log"
	"net/url"
	"runtime"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/crash"

	"easysend/assets"
	"easysend/internal/engine"
	"easysend/internal/settings"
	"easysend/ui"
	"easysend/ui/platform"
)

func main() {
	if err := crash.Init(crash.Options{Version: giode.AppVersion}); err != nil {
		log.Printf("crash reports: %v", err)
	}
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
	discovery := &multicastLock{}
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
		Frameless:      platform.IsDesktop, // the UI draws its own title bar
		Icon:           assets.Icon,
		Background:     "theme.background",
		StatusBar:      "theme.background",
		NavigationBar:  "theme.background",
		MinimizeToTray: s.MinimizeToTray,
		CloseToTray:    s.CloseToTray,
		StartInTray:    startInTray,
		SingleInstanceLock: &giode.SingleInstanceLock{
			UniqueID:     "app.easysend" + platform.Profile(),
			OnLaunchArgs: func(args []string) { u.AddPaths(args) },
		},
		OnShareReceived: func(s giode.SharedContent) { u.ReceiveShare(s) },
		OnOpenURL:       func(link *url.URL) { u.OpenLink(link) },
		OnBackground:    discovery.stop,
		OnForeground: func() {
			discovery.start()
			// iOS takes back a suspended app's sockets; reopen them.
			if runtime.GOOS == "ios" && eng != nil {
				eng.Resume()
			}
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
				Publish: publisher(app),
			})
			if err != nil {
				log.Fatal(err)
			}
			u = ui.New(app, eng)
			u.SetupTray()
			go u.KeepAliveWhileTransferring(stop)
			discovery.start()
			if platform.IsDesktop && platform.Profile() == "" {
				if err := app.InstallContextActions(); err != nil {
					log.Printf("context menu: %v", err)
				}
			}
		},
		OnShutdown: func() {
			close(stop)
			discovery.stop()
			if eng != nil {
				eng.Close()
			}
		},
	})
	if err := app.Run(func() giode.Element { return u.View() }); err != nil {
		log.Fatal(err)
	}
}
