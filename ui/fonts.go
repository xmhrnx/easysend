package ui

import (
	"fmt"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"

	"github.com/xmhrnx/giode"

	"easysend/assets"
)

// Font families used by the styles.
const (
	fontText    = "Inter"
	fontDisplay = "Inter Display"
)

// LoadFonts registers Inter (and the Go fonts as a fallback for glyphs
// Inter lacks).
func LoadFonts() error {
	faces := []struct {
		data   []byte
		family string
		weight font.Weight
	}{
		{assets.InterRegular, fontText, font.Normal},
		{assets.InterMedium, fontText, font.Medium},
		{assets.InterSemiBold, fontText, font.SemiBold},
		{assets.InterBold, fontText, font.Bold},
		{assets.InterDisplayBold, fontDisplay, font.Bold},
	}
	var all []font.FontFace
	for _, f := range faces {
		parsed, err := opentype.ParseCollection(f.data)
		if err != nil {
			return fmt.Errorf("font %s: %w", f.family, err)
		}
		for _, p := range parsed {
			p.Font.Typeface = font.Typeface(f.family)
			p.Font.Weight = f.weight
			all = append(all, p)
		}
	}
	scripts, err := scriptFaces()
	if err != nil {
		return err
	}
	all = append(all, scripts...)
	all = append(all, gofont.Collection()...)
	giode.RegisterFonts(all)
	return nil
}
