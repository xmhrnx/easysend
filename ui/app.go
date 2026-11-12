// Package ui is EasySend's interface, built with Giode.
package ui

import (
	"bytes"
	"gioui.org/layout"
	"image"
	_ "image/png" // logo decoding

	"github.com/xmhrnx/giode"
	prop "github.com/xmhrnx/giode/styles/properties"
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
	u.toasts = giode.Toaster().Placement(giode.ToastTopCenter).Max(3)
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
	// The app is drawn edge to edge (Config.EdgeToEdge): the page runs
	// under the status bar and home indicator, padded clear of them, and
	// on phones the tab bar floats over it.
	safe := giode.SafeArea()
	body := giode.Responsive(
		giode.Breakpoint{MaxWidth: 760, Name: "compact", View: func() giode.Element {
			return giode.Box(giode.Styles{Width: -1, Height: -1, Background: "theme.background"},
				toastInset(u, safe.Top),
				giode.Box(giode.Styles{Width: -1, Height: -1, Padding: giode.Inset{Top: safe.Top, Left: safe.Left, Right: safe.Right}}, page),
				giode.Box(giode.Styles{Position: giode.PositionAbsolute, Left: prop.Left(navMargin + safe.Left), Right: prop.Right(navMargin + safe.Right), Bottom: prop.Bottom(navBottom())}, u.bottomNav),
			)
		}},
		giode.Breakpoint{View: func() giode.Element {
			// The sidebar runs to the window's edges, keeping its content
			// clear of the system bars itself (see sidebar.Layout).
			return giode.Row(giode.Styles{Width: -1, Height: -1, Background: "theme.background", Wrap: giode.WrapNoWrap},
				toastInset(u, safe.Top+u.frameHeight()),
				u.sidebar,
				giode.Box(giode.Styles{FlexGrow: 1, Height: -1, Padding: giode.Inset{Top: safe.Top, Right: safe.Right, Bottom: safe.Bottom}}, page),
			)
		}},
	)
	main := giode.Box(giode.Styles{Width: -1, Height: -1}, body)
	if u.frame != nil {
		main = giode.Stack(giode.Styles{Width: -1, Height: -1, Wrap: giode.WrapNoWrap},
			u.frame,
			giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, body))
	}
	// Overlays (sheets, toasts) place themselves relative to where they
	// sit, so each sits at the window's top-left, over the page: in the
	// column after a full-height page they started at the bottom edge,
	// and toasts asked to go to the top landed at the bottom.
	kids := []giode.Element{main}
	for _, o := range overlays {
		kids = append(kids, giode.Box(giode.Styles{Position: giode.PositionAbsolute, Top: giode.Flush, Left: giode.Flush, Width: -1, Height: -1}, o))
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1}, kids...)
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

// toastInset keeps toasts top dp below the window's top edge: clear of
// the status bar on phones and the title bar on desktops. It's an
// element, so it applies only for the layout actually drawn: Responsive
// builds every breakpoint's view, and setting it in a view let the wide
// layout's inset win on phones too.
func toastInset(u *UI, top int) giode.Element {
	return giode.RawFunc(func(gtx layout.Context) layout.Dimensions {
		u.toasts.Inset(giode.Inset{Top: top})
		return layout.Dimensions{}
	})
}

// frameHeight is the custom title bar's height on desktops, 0 elsewhere.
func (u *UI) frameHeight() int {
	if u.frame == nil {
		return 0
	}
	return titleBarHeight
}
