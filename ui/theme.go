package ui

import (
	"github.com/xmhrnx/giode"
	prop "github.com/xmhrnx/giode/styles/properties"

	"easysend/internal/settings"
)

// accent is a pair of brand colors the user can pick.
type accent struct {
	Name     string
	From, To string
}

var accents = []accent{
	{"Aurora", "#7C5CFF", "#22D3EE"},
	{"Sunset", "#F97316", "#EC4899"},
	{"Emerald", "#10B981", "#0EA5E9"},
	{"Orchid", "#D946EF", "#6366F1"},
	{"Ember", "#EF4444", "#F59E0B"},
}

func currentAccent(s settings.Settings) accent {
	if s.Accent >= 0 && s.Accent < len(accents) {
		return accents[s.Accent]
	}
	return accents[0]
}

// applyTheme sets the palettes and mode from the settings.
func applyTheme(s settings.Settings, systemDark bool) {
	a := currentAccent(s)
	giode.SetPalette(giode.ModeDark, giode.Palette{
		Primary: giode.Color(a.From), Accent: giode.Color(a.To), Secondary: giode.Color(a.To),
		Background: "#08090F", Surface: "#11121C", Text: "#ECEEFF", Muted: "#8A8FAD", Border: "#23253A",
		OnPrimary: "#FFFFFF", Success: "#34D399", Warning: "#FBBF24", Danger: "#F87171", Info: "#60A5FA",
	})
	giode.SetPalette(giode.ModeLight, giode.Palette{
		Primary: giode.Color(a.From), Accent: giode.Color(a.To), Secondary: giode.Color(a.To),
		Background: "#F2F3F9", Surface: "#FFFFFF", Text: "#12142A", Muted: "#646A88", Border: "#E3E5F0",
		OnPrimary: "#FFFFFF", Success: "#059669", Warning: "#D97706", Danger: "#DC2626", Info: "#2563EB",
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
	colPrimarySoft = giode.Mix("theme.primary", "theme.surface", 0.14)
	colAccentSoft  = giode.Mix("theme.accent", "theme.surface", 0.14)
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

// orb is a round, gradient-filled badge holding an icon.
func orb(s settings.Settings, iconName string, size int) giode.Element {
	return giode.Box(giode.Styles{
		Width: prop.Width(size), Height: prop.Height(size), BorderRadius: prop.BorderRadius(size / 2), BackgroundGradient: brandGradient(s, 135),
		Align: giode.AlignCenter, Justify: giode.JustifyCenter,
		BoxShadow: giode.BoxShadow{Y: size / 6, Blur: size / 2, Color: giode.Color(currentAccent(s).From + "66")},
	}, icon(iconName, size*9/20, "#FFFFFF"))
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
