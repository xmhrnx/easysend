package ui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"easysend/internal/engine"
	"easysend/ui/platform"
)

// syncScreen lists the folders kept in sync with other devices.
type syncScreen struct {
	u      *UI
	scroll *scroll.Scroll
	add    *giode.ButtonElement
	empty  *giode.ButtonElement
}

func newSyncScreen(u *UI) *syncScreen {
	s := &syncScreen{u: u}
	s.add = primaryButton("Add folder", "plus", s.addFolder)
	s.empty = primaryButton("Add folder", "plus", s.addFolder)
	s.scroll = giode.Scroll(s.content, scrollStyles)
	return s
}

// addFolder starts syncing a folder: on desktops one the user picks; on
// phones, where apps can't use the user's folders, a new one in
// EasySend's, named in a sheet.
func (s *syncScreen) addFolder() {
	if !platform.IsDesktop {
		s.u.newFolder.open()
		return
	}
	go func() {
		dir, err := s.u.app.PickFolder(context.Background(), giode.FolderOptions{Title: "Choose a folder to sync"})
		if errors.Is(err, giode.ErrCanceled) {
			return
		}
		s.u.app.Post(func() {
			if err != nil {
				s.u.toast(giode.ToastError, "Couldn't choose the folder: "+err.Error())
				return
			}
			s.u.startSyncing(filepath.Base(dir), dir)
		})
	}()
}

// startSyncing adds a folder and goes on to choose its devices.
func (u *UI) startSyncing(name, root string) {
	f, err := u.eng.AddSyncFolder(name, root)
	if err != nil {
		u.toast(giode.ToastError, "Couldn't sync the folder: "+err.Error())
		return
	}
	u.syncDevices.open(f.Folder)
}

func (s *syncScreen) content() giode.Element {
	folders := s.u.eng.SyncFolders()
	var body []giode.Element
	for _, f := range folders {
		body = append(body, s.u.syncCardFor(f))
	}
	var trailing giode.Element
	if len(folders) == 0 {
		where := "Pick a folder on this computer"
		if !platform.IsDesktop {
			where = "Make a folder in EasySend"
		}
		body = append(body, card(giode.Styles{Align: giode.AlignCenter, Padding: giode.UniformInset(40), Gap: 10},
			orb("sync", 64),
			giode.Text("No synced folders yet", stHeading),
			giode.Text(where+", choose the devices to share it with, and it stays the same on all of them: add, change or delete a file on one, and the others follow.",
				giode.Styles{FontFamily: fontText, FontSize: 13, Color: "theme.muted", TextAlign: giode.TextAlignCenter, MaxWidth: 460}),
			giode.Box(giode.Styles{Padding: giode.Inset{Top: 6}}, s.empty),
		))
		if !platform.IsDesktop {
			body = append(body, giode.Text("Folders sync while EasySend is open, directly between your devices on the same network.", stSmall))
		}
	} else {
		trailing = s.add
		body = append(body, giode.Text("Files go straight between your devices on the network, never through a server. When two devices change the same file, the later change wins and the other is kept as a conflict copy.", stSmall))
	}
	return pageFrame("Sync", "Folders kept the same on your devices.", trailing, body...)
}

func (s *syncScreen) Layout(gtx layout.Context) layout.Dimensions { return s.scroll.Layout(gtx) }

// syncDeviceName is a folder's device as the UI names it: its current
// name from discovery, else the one remembered.
func (u *UI) syncDeviceName(p engine.SyncPeer) string {
	for _, d := range u.eng.Peers() {
		if d.Fingerprint == p.FP {
			return d.Alias
		}
	}
	if p.Name != "" {
		return p.Name
	}
	return "A device"
}

// displayPath is a folder's path as people know it: from the app's own
// folder on phones ("EasySend/Synced/Photos", as the Files app shows
// it), from "~" on desktops.
func (u *UI) displayPath(p string) string {
	base := filepath.Dir(u.eng.Settings.Get().DownloadDir)
	if !platform.IsDesktop {
		if rel, err := filepath.Rel(base, p); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
		return p
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if rel, err := filepath.Rel(home, p); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.Join("~", rel)
		}
	}
	return p
}
