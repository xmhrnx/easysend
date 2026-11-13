package ui

import (
	"fmt"

	"gioui.org/font"
	"gioui.org/font/opentype"

	"easysend/assets"
)

// scriptFaces returns the Noto Sans faces for scripts Inter lacks, each under
// its own family name so they act as per-glyph fallbacks behind Inter.
func scriptFaces() ([]font.FontFace, error) {
	// Each script's Regular and Bold files (Bold nil where Noto ships a
	// Regular-only subset).
	scripts := []struct {
		family        string
		regular, bold []byte
	}{
		{"Noto Sans Georgian", assets.NotoSansGeorgianRegular, assets.NotoSansGeorgianBold},
		{"Noto Sans Armenian", assets.NotoSansArmenianRegular, assets.NotoSansArmenianBold},
		{"Noto Sans Hebrew", assets.NotoSansHebrewRegular, assets.NotoSansHebrewBold},
		{"Noto Sans Arabic", assets.NotoSansArabicRegular, assets.NotoSansArabicBold},
		{"Noto Sans Devanagari", assets.NotoSansDevanagariRegular, assets.NotoSansDevanagariBold},
		{"Noto Sans Thai", assets.NotoSansThaiRegular, assets.NotoSansThaiBold},
		{"Noto Sans SC", assets.NotoSansSCRegular, nil},
		{"Noto Sans JP", assets.NotoSansJPRegular, nil},
		{"Noto Sans KR", assets.NotoSansKRRegular, nil},
	}
	// Every weight the styles use. The shaper falls back to another
	// family for glyphs Inter lacks only among faces of the weight asked
	// for: with just Regular and Bold registered, Medium and SemiBold
	// text (the sidebar, buttons) in these scripts drew as boxes.
	// Medium uses the Regular file and SemiBold the Bold one.
	weights := []struct {
		weight font.Weight
		bold   bool
	}{
		{font.Normal, false},
		{font.Medium, false},
		{font.SemiBold, true},
		{font.Bold, true},
	}
	var all []font.FontFace
	for _, sc := range scripts {
		for _, w := range weights {
			data := sc.regular
			if w.bold && sc.bold != nil {
				data = sc.bold
			}
			parsed, err := opentype.ParseCollection(data)
			if err != nil {
				return nil, fmt.Errorf("font %s: %w", sc.family, err)
			}
			for _, p := range parsed {
				p.Font.Typeface = font.Typeface(sc.family)
				p.Font.Weight = w.weight
				all = append(all, p)
			}
		}
	}
	return all, nil
}
