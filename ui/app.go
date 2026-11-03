// Package ui is EasySend's interface, built with Giode.
package ui

import (
	"bytes"
	"context"
	"errors"
	"image"
	_ "image/png" // logo decoding
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmhrnx/giode"

	"easysend/assets"
	"easysend/internal/engine"
	"easysend/internal/item"
	"easysend/internal/transfer"
	"easysend/ui/platform"
)

// screen identifies a page of the app.
type screen int

const (
	screenSend screen = iota
	screenReceive
	screenActivity
	screenSettings
)

// UI is the root of the interface and the state it owns.
type UI struct {
	app  *giode.App
	eng  *engine.Engine
	logo image.Image

	current screen
	// items is what the user picked to send.
	items []item.Item

	frame     *windowFrame // nil on phones
	sidebar   *sidebar
	bottomNav *bottomNav
	send      *sendScreen
	receive   *receiveScreen
	activity  *activityScreen
	settings  *settingsScreen

	incoming *incomingDialog
	pin      *pinDialog
	compose  *composeDialog
	address  *addressDialog
	toasts   *giode.ToasterElement
}

// New builds the interface for a running engine.
func New(app *giode.App, eng *engine.Engine) *UI {
	u := &UI{app: app, eng: eng}
	if img, _, err := image.Decode(bytes.NewReader(assets.Logo)); err == nil {
		u.logo = img
	}
	u.toasts = giode.Toaster().Placement(giode.ToastBottomCenter).Max(3)
	if platform.IsDesktop {
		u.frame = newWindowFrame(u)
	}
	u.sidebar = newSidebar(u)
	u.bottomNav = newBottomNav(u)
	u.send = newSendScreen(u)
	u.receive = newReceiveScreen(u)
	u.activity = newActivityScreen(u)
	u.settings = newSettingsScreen(u)
	u.incoming = newIncomingDialog(u)
	u.pin = newPINDialog(u)
	u.compose = newComposeDialog(u)
	u.address = newAddressDialog(u)
	u.ApplyTheme()
	return u
}

// ApplyTheme applies the theme settings.
func (u *UI) ApplyTheme() {
	dark, err := giode.SystemDarkMode()
	if err != nil {
		dark = true
	}
	applyTheme(u.eng.Settings.Get(), dark)
}

// View is the root view, run every frame.
func (u *UI) View() giode.Element {
	page := u.page()
	overlays := []giode.Element{u.incoming, u.pin, u.compose, u.address, u.toasts}
	body := giode.Responsive(
		giode.Breakpoint{MaxWidth: 760, Name: "compact", View: func() giode.Element {
			return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: "theme.background"},
				giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, page),
				u.bottomNav,
			)
		}},
		giode.Breakpoint{View: func() giode.Element {
			return giode.Row(giode.Styles{Width: -1, Height: -1, Background: "theme.background", Wrap: giode.WrapNoWrap},
				u.sidebar,
				giode.Box(giode.Styles{FlexGrow: 1, Height: -1}, page),
			)
		}},
	)
	main := giode.Box(giode.Styles{Width: -1, Height: -1}, body)
	if u.frame != nil {
		main = giode.Stack(giode.Styles{Width: -1, Height: -1, Wrap: giode.WrapNoWrap},
			u.frame,
			giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, body))
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1}, append([]giode.Element{main}, overlays...)...)
}

func (u *UI) page() giode.Element {
	switch u.current {
	case screenReceive:
		return u.receive
	case screenActivity:
		return u.activity
	case screenSettings:
		return u.settings
	}
	return u.send
}

// Go switches to a screen.
func (u *UI) Go(s screen) { u.current = s }

// toast shows a short message.
func (u *UI) toast(kind giode.ToastKind, msg string) { u.toasts.Show(kind, msg) }

// AddPaths adds files or folders to the selection (drag and drop, "Open
// with", the file manager's context menu) and shows the Send screen.
func (u *UI) AddPaths(paths []string) {
	go func() {
		var added []item.Item
		var firstErr error
		for _, p := range paths {
			its, err := item.FromPath(p)
			if err != nil && firstErr == nil {
				firstErr = err
			}
			added = append(added, its...)
		}
		u.app.Post(func() {
			u.addItems(added)
			if firstErr != nil {
				u.toast(giode.ToastError, "Couldn't add: "+firstErr.Error())
			}
		})
	}()
}

func (u *UI) addItems(its []item.Item) {
	if len(its) == 0 {
		return
	}
	u.items = append(u.items, its...)
	u.current = screenSend
	u.app.ShowWindow()
}

func (u *UI) removeItem(id string) {
	giode.AfterLayout(func() {
		for i, it := range u.items {
			if it.ID == id {
				u.items = append(u.items[:i], u.items[i+1:]...)
				return
			}
		}
	})
}

func (u *UI) clearItems() { giode.AfterLayout(func() { u.items = nil }) }

// pickFiles shows the file picker and adds the chosen files.
func (u *UI) pickFiles() {
	go func() {
		files, err := u.app.PickFiles(context.Background(), giode.PickOptions{Multiple: true})
		if errors.Is(err, giode.ErrCanceled) {
			return
		}
		if err != nil {
			u.app.Post(func() { u.toast(giode.ToastError, "Couldn't open files: "+err.Error()) })
			return
		}
		var added []item.Item
		for _, f := range files {
			it, err := pickedItem(f)
			f.Close()
			if err != nil {
				u.app.Post(func() { u.toast(giode.ToastError, "Couldn't read "+f.Name+": "+err.Error()) })
				continue
			}
			added = append(added, it)
		}
		u.app.Post(func() { u.addItems(added) })
	}()
}

// pickedItem turns a picked file into an item; files without a path
// (phones) are copied to the outbox first, so they can be read again.
func pickedItem(f giode.PickedFile) (item.Item, error) {
	if f.Path != "" {
		return item.FromFile(f.Path)
	}
	dir := filepath.Join(platform.DataDir(), "outbox", item.NewID())
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return item.Item{}, err
	}
	dst := item.SafeJoin(dir, f.Name)
	out, err := os.Create(dst)
	if err != nil {
		return item.Item{}, err
	}
	if _, err := io.Copy(out, f); err != nil {
		out.Close()
		return item.Item{}, err
	}
	if err := out.Close(); err != nil {
		return item.Item{}, err
	}
	return item.FromFile(dst)
}

// pickFolder shows the folder picker and adds every file inside.
func (u *UI) pickFolder() {
	go func() {
		dir, err := u.app.PickFolder(context.Background(), giode.FolderOptions{Title: "Choose a folder to send"})
		if errors.Is(err, giode.ErrCanceled) {
			return
		}
		if err != nil {
			u.app.Post(func() { u.toast(giode.ToastError, "Couldn't open the folder: "+err.Error()) })
			return
		}
		u.AddPaths([]string{dir})
	}()
}

// pasteText adds the clipboard's text as an item.
func (u *UI) pasteText() {
	u.app.ReadClipboard(func(text string) {
		if text == "" {
			u.toast(giode.ToastInfo, "The clipboard has no text")
			return
		}
		u.addItems([]item.Item{item.FromText(text)})
		u.toast(giode.ToastSuccess, "Added text from the clipboard")
	})
}

// showReceivedLabel names the action that shows received files.
var showReceivedLabel = func() string {
	switch {
	case platform.ReceivesToDownloads:
		return "Open Downloads"
	case platform.IsDesktop:
		return "Show in folder"
	}
	return "Show in Files"
}()

// savedFiles lists where a transfer's files were saved.
func savedFiles(t transfer.Transfer) []string {
	var out []string
	for _, f := range t.Files {
		if f.Saved != "" {
			out = append(out, f.Saved)
		}
	}
	return out
}

// showReceived shows where a received transfer's files went: Android's
// Downloads, the file itself in its folder on desktops, or the folder.
func (u *UI) showReceived(t transfer.Transfer) {
	saved := savedFiles(t)
	go func() {
		ctx := context.Background()
		var err error
		switch {
		case platform.ReceivesToDownloads:
			err = u.app.OpenDownloads(ctx, "EasySend")
		case len(saved) == 1 && isFile(saved[0]) && platform.IsDesktop:
			err = u.app.Share(ctx, giode.ShareFile{Path: saved[0]}) // reveals it in its folder
		case len(saved) == 1 && !isFile(saved[0]):
			err = u.app.OpenFolder(ctx, saved[0]) // a received folder
		case len(saved) > 0:
			err = u.app.OpenFolder(ctx, filepath.Dir(saved[0]))
		default:
			err = u.app.OpenFolder(ctx, t.Folder)
		}
		if err != nil {
			u.app.Post(func() { u.toast(giode.ToastError, "Couldn't open: "+err.Error()) })
		}
	}()
}

// openReceivedFolder opens the folder received files go to.
func (u *UI) openReceivedFolder() {
	go func() {
		ctx := context.Background()
		var err error
		if platform.ReceivesToDownloads {
			err = u.app.OpenDownloads(ctx, "EasySend")
		} else {
			dir := u.eng.Settings.Get().DownloadDir
			if err = os.MkdirAll(dir, 0o755); err == nil {
				err = u.app.OpenFolder(ctx, dir)
			}
		}
		if err != nil {
			u.app.Post(func() { u.toast(giode.ToastError, "Couldn't open the folder: "+err.Error()) })
		}
	}()
}

func isFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// OpenLink handles a link that opened the app: a scanned internet code
// (wormhole-transfer:7-crossover-clockwork) starts receiving it.
func (u *UI) OpenLink(link *url.URL) {
	if link.Scheme != "wormhole-transfer" {
		return
	}
	code := link.Opaque
	if code == "" {
		code = strings.TrimPrefix(link.Host+link.Path, "/")
	}
	if code == "" {
		return
	}
	u.eng.ReceiveInternet(code)
	u.current = screenActivity
	u.app.ShowWindow()
	u.toast(giode.ToastInfo, "Receiving with code "+code+"…")
}

// ReceiveShare adds what another app shared through the system share
// sheet to the selection, and shows the Send screen.
func (u *UI) ReceiveShare(s giode.SharedContent) {
	u.current = screenSend
	if s.Text != "" {
		u.items = append(u.items, item.FromText(s.Text))
	}
	if len(s.Files) == 0 {
		return
	}
	go func() {
		var added []item.Item
		for _, f := range s.Files {
			it, err := pickedItem(f)
			f.Close()
			if err != nil {
				name := f.Name
				u.app.Post(func() { u.toast(giode.ToastError, "Couldn't read "+name+": "+err.Error()) })
				continue
			}
			added = append(added, it)
		}
		u.app.Post(func() { u.addItems(added) })
	}()
}

// copy puts text on the clipboard.
func (u *UI) copy(text, what string) {
	u.app.SetClipboard(text)
	u.toast(giode.ToastSuccess, what+" copied")
}
