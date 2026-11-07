package ui

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/xmhrnx/giode"

	"easysend/internal/item"
	"easysend/ui/platform"
)

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
