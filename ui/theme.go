package ui

import (
	"github.com/xmhrnx/giode"
	prop "github.com/xmhrnx/giode/styles/properties"

	"easysend/internal/settings"
)

// accent is a soft metallic tone the user can pick: a gradient for the
// call-to-action buttons, a solid tone for highlights, and the ink drawn
// on top of the gradient.
type accent struct {
	Name     string
	From, To string
	Solid    string
	Ink      string
}

var accents = []accent{
	{"Gold", "#E0C894", "#A88A52", "#C8A96B", "#17130C"},
	{"Rose gold", "#E3C3B4", "#A47665", "#C79A87", "#1A110D"},
	{"Amber", "#E6C595", "#AE8152", "#CDA06A", "#19110A"},
	{"Champagne", "#E8DEC8", "#AE9D7E", "#CBBB98", "#18150F"},
	{"Bronze", "#D5B394", "#86654A", "#B08A67", "#170F09"},
}

func currentAccent(s settings.Settings) accent {
	if s.Accent >= 0 && s.Accent < len(accents) {
		return accents[s.Accent]
	}
	return accents[0]
}

// applyTheme sets the warm palettes and the mode from the settings.
func applyTheme(s settings.Settings, systemDark bool) {
	a := currentAccent(s)
	giode.SetPalette(giode.ModeDark, giode.Palette{
		Primary: giode.Color(a.Solid), Accent: giode.Color(a.Solid), Secondary: giode.Color(a.To),
		Background: "#0D0B09", Surface: "#17140F", Text: "#F5EEE3", Muted: "#A3978A", Border: "#2B251D",
		OnPrimary: giode.Color(a.Ink), Success: "#8CC08A", Warning: "#E9A84C", Danger: "#E5735F", Info: "#B8A68A",
	})
	giode.SetPalette(giode.ModeLight, giode.Palette{
		Primary: giode.Color(a.To), Accent: giode.Color(a.To), Secondary: giode.Color(a.Solid),
		Background: "#F8F4EC", Surface: "#FFFDF9", Text: "#1E1912", Muted: "#7B6F60", Border: "#E9E0D1",
		OnPrimary: giode.Color(a.Ink), Success: "#4C8A4A", Warning: "#B9741B", Danger: "#C2462F", Info: "#8C6A2A",
	})
	dark := systemDark
	switch s.Theme {
	case settings.ThemeDark:
		dark = true
	case settings.ThemeLight:
		dark = false
	}
	if dark {
		giode.SetMode(giode.ModeDark)
	} else {
		giode.SetMode(giode.ModeLight)
	}
}

// Shared colors. Mixed with theme tokens, they follow light/dark switches.
var (
	colSurface2    = giode.Mix("theme.text", "theme.surface", 0.035)
	colSurface3    = giode.Mix("theme.text", "theme.surface", 0.07)
	colPrimarySoft = giode.Mix("theme.primary", "theme.surface", 0.1)
	colAccentSoft  = giode.Mix("theme.accent", "theme.surface", 0.09)
	colOrbFill     = giode.Mix("theme.primary", "theme.surface", 0.08)
	colOrbRing     = giode.Mix("theme.primary", "theme.surface", 0.38)
	colSuccessSoft = giode.Mix("theme.success", "theme.surface", 0.15)
	colDangerSoft  = giode.Mix("theme.danger", "theme.surface", 0.14)
	colWarningSoft = giode.Mix("theme.warning", "theme.surface", 0.15)
	colHover       = giode.Mix("theme.text", "theme.surface", 0.06)
	colPressed     = giode.Mix("theme.text", "theme.surface", 0.1)
)

// brandGradient is the accent gradient at angle degrees.
func brandGradient(s settings.Settings, angle float32) *giode.LinearGradient {
	a := currentAccent(s)
	return giode.Linear(angle, giode.Color(a.From), giode.Color(a.To))
}

// Text styles.
var (
	stTitle    = giode.Styles{FontFamily: fontDisplay, FontWeight: giode.FontWeightBold, FontSize: 30, Color: "theme.text"}
	stHeading  = giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightSemiBold, FontSize: 17, Color: "theme.text"}
	stBody     = giode.Styles{FontFamily: fontText, FontSize: 14, Color: "theme.text"}
	stStrong   = giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightSemiBold, FontSize: 14, Color: "theme.text"}
	stMuted    = giode.Styles{FontFamily: fontText, FontSize: 13, Color: "theme.muted"}
	stSmall    = giode.Styles{FontFamily: fontText, FontSize: 12, Color: "theme.muted"}
	stOverline = giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightSemiBold, FontSize: 11, Color: "theme.muted", TextTransform: giode.TextTransformUppercase}
)

// card is the standard panel.
func card(st giode.Styles, kids ...giode.Element) giode.Element {
	base := giode.Styles{
		FlexDirection: giode.FlexDirectionColumn, Gap: 14, Padding: giode.UniformInset(20),
		Background: "theme.surface", BorderRadius: 22, BorderWidth: 1, BorderColor: "theme.border",
		Width: -1,
	}
	return giode.Box(base.With(st), kids...)
}

// orb is a round badge with a soft accent ring holding an icon.
func orb(iconName string, size int) giode.Element {
	return giode.Box(giode.Styles{
		Width: prop.Width(size), Height: prop.Height(size), BorderRadius: prop.BorderRadius(size / 2),
		Background: colOrbFill, BorderWidth: 1, BorderColor: colOrbRing,
		Align: giode.AlignCenter, Justify: giode.JustifyCenter,
	}, icon(iconName, size*9/20, "theme.primary"))
}

// tile is a rounded square with a tinted background and an icon.
func tile(iconName string, size int, bg, fg string) giode.Element {
	return giode.Box(giode.Styles{
		Width: prop.Width(size), Height: prop.Height(size), BorderRadius: prop.BorderRadius(size * 3 / 10), Background: giode.Color(bg),
		Align: giode.AlignCenter, Justify: giode.JustifyCenter,
	}, icon(iconName, size/2, fg))
}

// chip is a small rounded label.
func chip(text, bg, fg string) giode.Element {
	return giode.Box(giode.Styles{Background: giode.Color(bg), BorderRadius: 99, Padding: giode.SymmetricInset(10, 4)},
		giode.Text(text, giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightMedium, FontSize: 12, Color: giode.Color(fg)}))
}

// spacer grows to fill a row.
func spacer() giode.Element { return giode.Box(giode.Styles{FlexGrow: 1}) }
