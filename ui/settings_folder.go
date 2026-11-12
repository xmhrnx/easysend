package ui

import (
	"context"
	"errors"

	"github.com/xmhrnx/giode"

	"easysend/internal/settings"
)

func (s *settingsScreen) pickFolder() {
	go func() {
		dir, err := s.u.app.PickFolder(context.Background(), giode.FolderOptions{Title: tr("Where should received files go?"), Directory: s.u.eng.Settings.Get().DownloadDir})
		if errors.Is(err, giode.ErrCanceled) {
			return
		}
		s.u.app.Post(func() {
			if err != nil {
				s.u.toast(giode.ToastError, tr("Couldn't choose the folder: {error}", "error", err.Error()))
				return
			}
			_ = s.u.eng.UpdateSettings(func(st *settings.Settings) { st.DownloadDir = dir })
			s.u.toast(giode.ToastSuccess, tr("Received files now go to {folder}", "folder", dir))
		})
	}()
}
