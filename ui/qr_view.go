package ui

import (
	"sync"

	"github.com/skip2/go-qrcode"
	"github.com/xmhrnx/giode"
	prop "github.com/xmhrnx/giode/styles/properties"
)

var qrCache sync.Map // text -> [][]bool

func qrBitmap(text string) [][]bool {
	if v, ok := qrCache.Load(text); ok {
		return v.([][]bool)
	}
	q, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		return nil
	}
	q.DisableBorder = true
	bm := q.Bitmap()
	qrCache.Store(text, bm)
	return bm
}

// qrView draws text as a QR code, dark on white so every scanner reads it.
func qrView(text string, size int) giode.Element {
	bm := qrBitmap(text)
	pad := size / 14
	return giode.Box(giode.Styles{Background: "#FFFFFF", BorderRadius: 18, Padding: giode.UniformInset(pad), Width: prop.Width(size), Height: prop.Height(size)},
		giode.Canvas(giode.Styles{Width: -1, Height: -1}, func(p *giode.Painter) {
			n := len(bm)
			if n == 0 {
				return
			}
			cell := p.Width() / float32(n)
			path := p.Path()
			for y, row := range bm {
				for x := 0; x < n; {
					if !row[x] {
						x++
						continue
					}
					start := x
					for x < n && row[x] {
						x++
					}
					x0, y0 := float32(start)*cell, float32(y)*cell
					x1, y1 := float32(x)*cell, y0+cell
					path.MoveTo(x0, y0)
					path.LineTo(x1, y0)
					path.LineTo(x1, y1+0.3)
					path.LineTo(x0, y1+0.3)
					path.Close()
				}
			}
			path.Fill("#0B0C14")
		}))
}
