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
	type face struct {
		data   []byte
		family string
		weight font.Weight
	}
	faces := []face{
		{assets.NotoSansGeorgianRegular, "Noto Sans Georgian", font.Normal},
		{assets.NotoSansGeorgianBold, "Noto Sans Georgian", font.Bold},
		{assets.NotoSansArmenianRegular, "Noto Sans Armenian", font.Normal},
		{assets.NotoSansArmenianBold, "Noto Sans Armenian", font.Bold},
		{assets.NotoSansHebrewRegular, "Noto Sans Hebrew", font.Normal},
		{assets.NotoSansHebrewBold, "Noto Sans Hebrew", font.Bold},
		{assets.NotoSansArabicRegular, "Noto Sans Arabic", font.Normal},
		{assets.NotoSansArabicBold, "Noto Sans Arabic", font.Bold},
		{assets.NotoSansDevanagariRegular, "Noto Sans Devanagari", font.Normal},
		{assets.NotoSansDevanagariBold, "Noto Sans Devanagari", font.Bold},
		{assets.NotoSansThaiRegular, "Noto Sans Thai", font.Normal},
		{assets.NotoSansThaiBold, "Noto Sans Thai", font.Bold},
		{assets.NotoSansSCRegular, "Noto Sans SC", font.Normal},
		{assets.NotoSansSCRegular, "Noto Sans SC", font.Bold}, // Regular-only subset; the shaper skips a family with no Bold
		{assets.NotoSansJPRegular, "Noto Sans JP", font.Normal},
		{assets.NotoSansJPRegular, "Noto Sans JP", font.Bold}, // Regular-only subset; the shaper skips a family with no Bold
		{assets.NotoSansKRRegular, "Noto Sans KR", font.Normal},
		{assets.NotoSansKRRegular, "Noto Sans KR", font.Bold}, // Regular-only subset; the shaper skips a family with no Bold
	}
	var all []font.FontFace
	for _, f := range faces {
		parsed, err := opentype.ParseCollection(f.data)
		if err != nil {
			return nil, fmt.Errorf("font %s: %w", f.family, err)
		}
		for _, p := range parsed {
			p.Font.Typeface = font.Typeface(f.family)
			p.Font.Weight = f.weight
			all = append(all, p)
		}
	}
	return all, nil
}
