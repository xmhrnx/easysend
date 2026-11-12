package ui

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"
	"github.com/xmhrnx/giode/components/sheet"

	"easysend/internal/settings"
)

// languageListHeight caps the language list, in dp; it scrolls past it.
const languageListHeight = 340

// languageDialog lets the user pick the interface language from a list
// of limited height that scrolls.
type languageDialog struct {
	u      *UI
	isOpen bool
	sheet  *sheet.Sheet
	list   *scroll.Scroll
	rows   []*giode.PressableElement // the system choice, then every language
}

func newLanguageDialog(u *UI) *languageDialog {
	d := &languageDialog{u: u}
	names := languageOptions()
	for i, name := range names {
		code := languageCode(i)
		d.rows = append(d.rows, giode.Pressable(func(ps giode.PressState) giode.Element { return d.row(name, code, ps) }).
			Label(name).OnClick(func(giode.ClickEvent) { d.choose(code) }))
	}
	d.list = giode.Scroll(d.content, scrollStyles)
	d.sheet = newSheet(&d.isOpen, tr("Language"), func() giode.Element {
		return giode.Box(giode.Styles{Width: -1, Height: languageListHeight, MaxHeight: languageListHeight}, d.list)
	}, nil)
	return d
}

func (d *languageDialog) open() { d.isOpen = true }

func (d *languageDialog) choose(code string) {
	d.isOpen = false
	_ = d.u.eng.UpdateSettings(func(s *settings.Settings) { s.Language = code })
	d.u.SetLanguage(code)
}

func (d *languageDialog) row(name, code string, ps giode.PressState) giode.Element {
	selected := d.u.eng.Settings.Get().Language == code
	bg, fg := "transparent", "theme.text"
	switch {
	case ps.Pressed:
		bg = colPressed
	case ps.Hovered:
		bg = colHover
	}
	var mark giode.Element = giode.Box(giode.Styles{Width: 18, Height: 18})
	if selected {
		fg = "theme.primary"
		mark = icon("check", 18, "theme.primary")
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(14, 12),
		BorderRadius: 12, Background: giode.Color(bg)},
		giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(name, giode.Styles{FontFamily: fontText, FontSize: 15, Color: giode.Color(fg)})),
		mark,
	)
}

func (d *languageDialog) content() giode.Element {
	rows := make([]giode.Element, len(d.rows))
	for i, r := range d.rows {
		rows[i] = r
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 2}, rows...)
}

func (d *languageDialog) Layout(gtx layout.Context) layout.Dimensions { return d.sheet.Layout(gtx) }

// languageName is the picker's name for a saved language code.
func languageName(code string) string { return languageOptions()[languageIndex(code)] }
