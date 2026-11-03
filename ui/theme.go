package ui

import (
	"github.com/xmhrnx/giode"
	prop "github.com/xmhrnx/giode/styles/properties"

	"easysend/internal/settings"
)

// accent is a metallic tone the user can pick. Solid draws lines, text and
// icons on dark surfaces, To on light ones; From and To paint the swatch in
// the picker.
type accent struct {
	Name     string
	From, To string
	Solid    string
}

var accents = []accent{
	{"Gold", "#E0C894", "#A88A52", "#C8A96B"},
	{"Rose gold", "#E3C3B4", "#A47665", "#C79A87"},
	{"Amber", "#E6C595", "#AE8152", "#CDA06A"},
	{"Champagne", "#E8DEC8", "#AE9D7E", "#CBBB98"},
	{"Bronze", "#D5B394", "#86654A", "#B08A67"},
}

func currentAccent(s settings.Settings) accent {
	if s.Accent >= 0 && s.Accent < len(accents) {
		return accents[s.Accent]
	}
	return accents[0]
}

// applyTheme sets the palettes and the mode from the settings.
func applyTheme(s settings.Settings, systemDark bool) {
	a := currentAccent(s)
	giode.SetPalette(giode.ModeDark, giode.Palette{
		Primary: giode.Color(a.Solid), Accent: giode.Color(a.Solid), Secondary: giode.Color(a.Solid),
		Background: "#0A0A0B", Surface: "#141416", Text: "#EDEDEF", Muted: "#8E8E96", Border: "#26262B",
		OnPrimary: "#0A0A0B", Success: "#8CC08A", Warning: "#E0A85A", Danger: "#E07A66", Info: "#A1A1AA",
	})
	giode.SetPalette(giode.ModeLight, giode.Palette{
		Primary: giode.Color(a.To), Accent: giode.Color(a.To), Secondary: giode.Color(a.To),
		Background: "#F4F4F6", Surface: "#FFFFFF", Text: "#18181B", Muted: "#6B6B74", Border: "#E4E4E8",
		OnPrimary: "#FFFFFF", Success: "#4C8A4A", Warning: "#A8681A", Danger: "#B9442E", Info: "#52525B",
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
// Surfaces are neutral dark (or white); the accent only draws lines, text
// and icons.
var (
	colSurface2 = giode.Mix("theme.text", "theme.surface", 0.035)
	colSurface3 = giode.Mix("theme.text", "theme.surface", 0.07)
	colHover    = giode.Mix("theme.text", "theme.surface", 0.06)
	colPressed  = giode.Mix("theme.text", "theme.surface", 0.1)
	colRing     = giode.Mix("theme.primary", "theme.surface", 0.45)
)

// line is a color token softened into a hairline outline.
func line(token string) string { return giode.Mix(token, "theme.surface", 0.45) }

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

// orb is a round badge: a dark disc with an accent ring and icon.
func orb(iconName string, size int) giode.Element {
	return giode.Box(giode.Styles{
		Width: prop.Width(size), Height: prop.Height(size), BorderRadius: prop.BorderRadius(size / 2),
		Background: colSurface2, BorderWidth: 1, BorderColor: colRing,
		Align: giode.AlignCenter, Justify: giode.JustifyCenter,
	}, icon(iconName, size*9/20, "theme.primary"))
}

// tile is a rounded square outlined in fg, holding an icon in fg.
func tile(iconName string, size int, fg string) giode.Element {
	return giode.Box(giode.Styles{
		Width: prop.Width(size), Height: prop.Height(size), BorderRadius: prop.BorderRadius(size * 3 / 10),
		BorderWidth: 1, BorderColor: giode.Color(line(fg)),
		Align: giode.AlignCenter, Justify: giode.JustifyCenter,
	}, icon(iconName, size/2, fg))
}

// chip is a small outlined label in fg.
func chip(text, fg string) giode.Element {
	return giode.Box(giode.Styles{BorderRadius: 99, BorderWidth: 1, BorderColor: giode.Color(line(fg)), Padding: giode.SymmetricInset(10, 3)},
		giode.Text(text, giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightMedium, FontSize: 12, Color: giode.Color(fg)}))
}

// spacer grows to fill a row.
func spacer() giode.Element { return giode.Box(giode.Styles{FlexGrow: 1}) }
