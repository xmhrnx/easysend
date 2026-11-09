// Package ui is EasySend's interface, built with Giode.
package ui

import (
	"bytes"
	"image"
	_ "image/png" // logo decoding

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/support"

	"easysend/assets"
	"easysend/internal/engine"
	"easysend/internal/item"
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
	help     *support.Support // report form, donations
}

// New builds the interface for a running engine.
func New(app *giode.App, eng *engine.Engine) *UI {
	u := &UI{app: app, eng: eng}
	if img, _, err := image.Decode(bytes.NewReader(assets.Logo)); err == nil {
		u.logo = img
	}
	u.toasts = giode.Toaster().Placement(giode.ToastBottomCenter).Max(3)
	u.help = newSupport(u)
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
	overlays := []giode.Element{u.incoming, u.pin, u.compose, u.address, u.help, u.toasts}
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
