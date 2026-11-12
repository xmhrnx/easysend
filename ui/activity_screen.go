package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"
	"github.com/xmhrnx/giode/components/segmented"

	"easysend/internal/transfer"
)

// activityScreen lists transfers in progress and the history.
type activityScreen struct {
	u      *UI
	scroll *scroll.Scroll
	filter *segmented.Control
	clear  *giode.ButtonElement
}

func newActivityScreen(u *UI) *activityScreen {
	a := &activityScreen{u: u}
	a.filter = giode.Input.Segmented(tr("All"), tr("Sent"), tr("Received"))
	a.filter.Styles(giode.Styles{Width: -1})
	a.clear = softButton(tr("Clear history"), "trash", func() { giode.AfterLayout(u.eng.Transfers.ClearFinished) })
	a.scroll = giode.Scroll(a.content, scrollStyles)
	return a
}

func (a *activityScreen) keep(t transfer.Transfer) bool {
	switch a.filter.Selected() {
	case 1:
		return t.Direction == transfer.Outgoing
	case 2:
		return t.Direction == transfer.Incoming
	}
	return true
}

func (a *activityScreen) content() giode.Element {
	var active, done []giode.Element
	for _, t := range a.u.eng.Transfers.List() {
		if !a.keep(t) {
			continue
		}
		if t.Status.Finished() {
			done = append(done, a.u.transferRowFor(t))
		} else {
			active = append(active, a.u.transferRowFor(t))
		}
	}
	body := []giode.Element{a.filter}
	if len(active) > 0 {
		body = append(body, sectionTitle(tr("In progress")))
		body = append(body, active...)
	}
	if len(done) > 0 {
		body = append(body, sectionTitle(tr("History"), a.clear))
		body = append(body, done...)
	}
	if len(active)+len(done) == 0 {
		body = append(body, card(giode.Styles{Align: giode.AlignCenter, Padding: giode.UniformInset(40), Gap: 10},
			orb("history", 64),
			giode.Text(tr("Nothing here yet"), stHeading),
			giode.Text(tr("Transfers you send and receive show up here, with their progress."), stMuted),
		))
	}
	return pageFrame(tr("Activity"), tr("Every transfer, live and past."), nil, body...)
}

func (a *activityScreen) Layout(gtx layout.Context) layout.Dimensions { return a.scroll.Layout(gtx) }
