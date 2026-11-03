package ui

import (
	"fmt"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/kit"
	prop "github.com/xmhrnx/giode/styles/properties"

	"easysend/internal/settings"
)

// radar draws expanding rings around an orb: the "looking for devices"
// and "ready to receive" animation.
func radar(s settings.Settings, size int, iconName string) giode.Element {
	a := currentAccent(s)
	r, g, b := hexRGB(a.From)
	rings := giode.Canvas(giode.Styles{Width: prop.Width(size), Height: prop.Height(size)}, func(p *giode.Painter) {
		c := p.Width() / 2
		maxR := c
		const period = 2.6
		now := float64(time.Now().UnixNano()) / 1e9
		for i := 0; i < 3; i++ {
			phase := float32((now/period + float64(i)/3) - float64(int(now/period+float64(i)/3)))
			rad := maxR*0.28 + phase*(maxR*0.72)
			alpha := (1 - phase) * 0.55
			p.Arc(c, c, rad, 2, 0, 360, fmt.Sprintf("rgba(%d,%d,%d,%.3f)", r, g, b, alpha))
		}
		p.Circle(c, c, maxR*0.34, fmt.Sprintf("rgba(%d,%d,%d,0.10)", r, g, b))
		kit.RequestFrame(p.Context())
	})
	orbSize := size * 30 / 100
	return giode.Box(giode.Styles{Width: prop.Width(size), Height: prop.Height(size), Align: giode.AlignCenter, Justify: giode.JustifyCenter},
		giode.Box(giode.Styles{Position: giode.PositionAbsolute, Top: giode.Flush, Left: giode.Flush}, rings),
		orb(s, iconName, orbSize),
	)
}

func hexRGB(hex string) (r, g, b int) {
	if len(hex) == 7 && hex[0] == '#' {
		_, _ = fmt.Sscanf(hex[1:], "%02x%02x%02x", &r, &g, &b)
	}
	return r, g, b
}
