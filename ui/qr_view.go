package ui

import (
	"image"
	"image/color"
	"sync"

	"github.com/skip2/go-qrcode"
	"github.com/xmhrnx/giode"
	prop "github.com/xmhrnx/giode/styles/properties"
)

// qrPixels is about how wide a QR code is rendered: wider than it's ever
// shown (168 dp at 3x), so drawing it is a plain downscale and stays
// sharp.
const qrPixels = 600

// qrCache keeps the QR codes of the last few texts: links come and go,
// and each image is about 1.4 MB.
var qrCache struct {
	sync.Mutex
	imgs  map[string]image.Image
	order []string
}

const qrCached = 8

// qrImage renders text as a QR code once, dark modules on white. The
// image is kept, so the GPU uploads it once and each frame just draws
// it: a vector path of every module, rebuilt per frame, made screens
// with a QR code sluggish (most of all on Metal, on Apple devices).
func qrImage(text string) image.Image {
	qrCache.Lock()
	img, ok := qrCache.imgs[text]
	qrCache.Unlock()
	if ok {
		return img
	}
	q, err := qrcode.New(text, qrcode.Medium)
	if err != nil {
		return nil
	}
	q.DisableBorder = true
	bm := q.Bitmap()
	n := len(bm)
	if n == 0 {
		return nil
	}
	scale := (qrPixels + n - 1) / n
	rgba := image.NewRGBA(image.Rect(0, 0, n*scale, n*scale))
	dark, light := color.RGBA{0x0B, 0x0C, 0x14, 0xFF}, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	for y, row := range bm {
		for x, on := range row {
			c := light
			if on {
				c = dark
			}
			for py := y * scale; py < (y+1)*scale; py++ {
				for px := x * scale; px < (x+1)*scale; px++ {
					rgba.SetRGBA(px, py, c)
				}
			}
		}
	}
	qrCache.Lock()
	defer qrCache.Unlock()
	if qrCache.imgs == nil {
		qrCache.imgs = map[string]image.Image{}
	}
	if old, ok := qrCache.imgs[text]; ok {
		return old
	}
	if len(qrCache.order) == qrCached {
		// Let go of its GPU upload too, which Giode keeps per image.
		prop.ForgetImage(qrCache.imgs[qrCache.order[0]])
		delete(qrCache.imgs, qrCache.order[0])
		qrCache.order = qrCache.order[1:]
	}
	qrCache.imgs[text] = rgba
	qrCache.order = append(qrCache.order, text)
	return rgba
}

// qrView draws text as a QR code, dark on white so every scanner reads it.
func qrView(text string, size int) giode.Element {
	pad := size / 14
	box := giode.Styles{Background: "#FFFFFF", BorderRadius: 18, Padding: giode.UniformInset(pad), Width: prop.Width(size), Height: prop.Height(size)}
	img := qrImage(text)
	if img == nil {
		return giode.Box(box)
	}
	return giode.Box(box, giode.Image(img, giode.Styles{Width: -1, Height: -1}))
}
