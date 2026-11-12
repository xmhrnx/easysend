package transfer

import (
	"github.com/xmhrnx/giode/i18n"
	"path/filepath"
	"time"
)

// Progress is the completed fraction in [0, 1].
func (t Transfer) Progress() float32 {
	if t.Total <= 0 {
		if t.Status == Done {
			return 1
		}
		return 0
	}
	p := float32(t.Done) / float32(t.Total)
	if p > 1 {
		p = 1
	}
	return p
}

// ETA estimates the remaining time.
func (t Transfer) ETA() time.Duration {
	if t.Speed <= 1 {
		return -1
	}
	return time.Duration(float64(t.Total-t.Done) / t.Speed * float64(time.Second))
}

// Title is a short description: the single file's name or "3 files".
func (t Transfer) Title() string {
	switch {
	case t.Text != "" && len(t.Files) <= 1:
		return i18n.Intl("Text message")
	case len(t.Files) == 1:
		return filepath.Base(t.Files[0].Name)
	case len(t.Files) == 0:
		return i18n.Intl("Transfer")
	}
	return i18n.Intl("files", "count", len(t.Files))
}
