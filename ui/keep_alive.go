package ui

import (
	"log"
	"time"
)

// KeepAliveWhileTransferring keeps the app running (phones: a foreground
// service; desktops: no idle sleep) while transfers are in progress. It
// returns when stop is closed.
func (u *UI) KeepAliveWhileTransferring(stop <-chan struct{}) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	var release func()
	for {
		select {
		case <-stop:
			if release != nil {
				release()
			}
			return
		case <-t.C:
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
