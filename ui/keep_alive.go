package ui

import (
	"log"
	"runtime"
	"time"
)

// ReceivesInBackground says whether the app should keep receiving after
// it leaves the screen: visible to nearby devices, and not turned off
// in Settings.
func (u *UI) ReceivesInBackground() bool {
	s := u.eng.Settings.Get()
	return s.Discoverable && !s.NoBackground
}

// KeepAliveWhileTransferring keeps the app running (phones: a foreground
// service; desktops: no idle sleep) while transfers are in progress. It
// returns when stop is closed.
func (u *UI) KeepAliveWhileTransferring(stop <-chan struct{}) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	var release, ready func()
	screenOn := false
	for {
		select {
		case <-stop:
			if release != nil {
				release()
			}
			if ready != nil {
				ready()
			}
			return
		case <-t.C:
		}
		// iOS freezes EasySend soon after it leaves the screen - the
		// iPhone locking itself included - and a frozen app can't serve
		// its links: keep the screen on while a link is shared or browser
		// uploads are on, so the other device can still reach it.
		if runtime.GOOS == "ios" {
			serving := len(u.eng.Shares()) > 0 || u.eng.Settings.Get().BrowserDrop
			if serving != screenOn {
				if err := u.app.KeepScreenOn(serving); err != nil {
					log.Printf("keep screen on: %v", err)
				} else {
					screenOn = serving
				}
			}
		}
		// On Android a foreground service keeps EasySend reachable after
		// it leaves the screen, so requests arrive (and notify) without
		// opening the app.
		want := runtime.GOOS == "android" && u.ReceivesInBackground()
		switch {
		case want && ready == nil:
			r, err := u.app.KeepAlive("Ready to receive from nearby devices")
			if err != nil {
				log.Printf("keep alive: %v", err)
			} else {
				ready = r
			}
		case !want && ready != nil:
			ready()
			ready = nil
		}
		active := false
		for _, tr := range u.eng.Transfers.List() {
			if !tr.Status.Finished() {
				active = true
				break
			}
		}
		switch {
		case active && release == nil:
			r, err := u.app.KeepAlive("Transferring files")
			if err != nil {
				log.Printf("keep alive: %v", err)
				continue
			}
			release = r
		case !active && release != nil:
			release()
			release = nil
		}
	}
}
