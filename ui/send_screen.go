package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"
	"github.com/xmhrnx/giode/components/segmented"
)

// sendScreen picks what to send and how.
type sendScreen struct {
	u         *UI
	scroll    *scroll.Scroll
	drop      *dropZone
	selection *selectionPanel
	method    *segmented.Control
	nearby    *nearbyPanel
	internet  *internetPanel
	link      *linkPanel
}

func newSendScreen(u *UI) *sendScreen {
	s := &sendScreen{u: u, drop: newDropZone(u), selection: newSelectionPanel(u),
		nearby: newNearbyPanel(u), internet: newInternetPanel(u), link: newLinkPanel(u)}
	s.method = giode.Input.Segmented("Nearby", "Internet", "Link & QR").
		Icons(icon("wifi", 16, ""), icon("globe", 16, ""), icon("qr", 16, ""))
	s.method.Styles(giode.Styles{Width: -1, FontSize: 14})
	s.scroll = giode.Scroll(s.content, scrollStyles)
	return s
}

func (s *sendScreen) content() giode.Element {
	body := []giode.Element{s.drop}
	if len(s.u.items) > 0 {
		body = append(body, s.selection)
	}
	body = append(body, giode.Box(giode.Styles{Width: -1, MaxWidth: 520}, s.method))
	switch s.method.Selected() {
	case 1:
		body = append(body, s.internet)
	case 2:
		body = append(body, s.link)
	default:
		body = append(body, s.nearby)
	}
	return pageFrame("Send", "Share files, folders and text — nearby, over the internet, or with a link.", nil, body...)
}

func (s *sendScreen) Layout(gtx layout.Context) layout.Dimensions { return s.scroll.Layout(gtx) }
