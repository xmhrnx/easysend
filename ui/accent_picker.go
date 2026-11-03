package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"

	"easysend/internal/settings"
)

// accentPicker shows the accent colors as swatches.
type accentPicker struct {
	u      *UI
	swatch []*giode.PressableElement
}

func newAccentPicker(u *UI) *accentPicker {
	a := &accentPicker{u: u}
	for i, ac := range accents {
		i, ac := i, ac
		p := giode.Pressable(func(ps giode.PressState) giode.Element {
			selected := u.eng.Settings.Get().Accent == i
			ring := "transparent"
			if selected {
				ring = "theme.text"
			} else if ps.Hovered {
				ring = "theme.border"
			}
			var mark giode.Element = giode.Box(giode.Styles{})
			if selected {
				mark = icon("check", 16, ac.Ink)
			}
			return giode.Box(giode.Styles{Padding: giode.UniformInset(3), BorderRadius: 20, BorderWidth: 2, BorderColor: giode.Color(ring)},
				giode.Box(giode.Styles{Width: 30, Height: 30, BorderRadius: 15, BackgroundGradient: giode.Linear(135, giode.Color(ac.From), giode.Color(ac.To)),
					Align: giode.AlignCenter, Justify: giode.JustifyCenter}, mark))
		}).Label(ac.Name).OnClick(func(giode.ClickEvent) {
			_ = u.eng.UpdateSettings(func(s *settings.Settings) { s.Accent = i })
			u.ApplyTheme()
			u.restyle()
		})
		a.swatch = append(a.swatch, p)
	}
	return a
}

func (a *accentPicker) Layout(gtx layout.Context) layout.Dimensions {
	kids := make([]giode.Element, len(a.swatch))
	for i, p := range a.swatch {
		kids[i] = p
	}
	return giode.Row(giode.Styles{Gap: 6}, kids...).Layout(gtx)
}
