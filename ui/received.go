package ui

import (
	"context"
	"os"
	"path/filepath"

	"github.com/xmhrnx/giode"

	"easysend/internal/transfer"
	"easysend/ui/platform"
)

// showReceivedLabel names the action that shows received files.
func showReceivedLabel() string {
	switch {
	case platform.ReceivesToDownloads:
		return tr("Open Downloads")
	case platform.IsDesktop:
		return tr("Show in folder")
	}
	return tr("Show in Files")
}

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
			u.app.Post(func() { u.toast(giode.ToastError, tr("Couldn't open: {error}", "error", err.Error())) })
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
			u.app.Post(func() { u.toast(giode.ToastError, tr("Couldn't open the folder: {error}", "error", err.Error())) })
		}
	}()
}

func isFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
