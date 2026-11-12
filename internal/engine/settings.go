package engine

import (
	"context"
	"errors"
	"github.com/xmhrnx/giode/i18n"

	"easysend/internal/localsend"
	"easysend/internal/settings"
)

// UpdateSettings changes the settings and applies what changed.
func (e *Engine) UpdateSettings(fn func(*settings.Settings)) error {
	before := e.Settings.Get()
	err := e.Settings.Update(fn)
	after := e.Settings.Get()
	if before.UPnP != after.UPnP {
		e.applyPublic()
	}
	if before.DeviceName != after.DeviceName || (!before.Discoverable && after.Discoverable) {
		go e.node.Refresh(context.Background())
	}
	e.changed()
	return err
}

// errorText turns errors into short messages for people.
func errorText(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, localsend.ErrPINRequired):
		return i18n.Intl("PIN required")
	case errors.Is(err, localsend.ErrDeclined):
		return i18n.Intl("Declined by the receiver")
	case errors.Is(err, localsend.ErrBusy):
		return i18n.Intl("The receiver is busy")
	case errors.Is(err, context.DeadlineExceeded):
		return i18n.Intl("Timed out")
	}
	return err.Error()
}
