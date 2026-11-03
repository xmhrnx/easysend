package ui

import (
	"math"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/kit"
	prop "github.com/xmhrnx/giode/styles/properties"
)

const (
	radarRings  = 3
	radarPeriod = 2.6 // seconds for a ring to travel out
	radarPeak   = 0.55
	radarSteps  = 100
)

// radarShades are the ring colors from invisible to radarPeak: the accent
// mixed into the card's surface, opaque. Translucent strokes blend in
// linear light, where a nearly transparent ring still shows and then
// vanishes at once; an opaque mix fades smoothly into the background.
var radarShades = func() [radarSteps + 1]string {
	var out [radarSteps + 1]string
	for i := range out {
		out[i] = giode.Mix("theme.primary", "theme.surface", radarPeak*float64(i)/radarSteps)
	}
	return out
}()

// radar draws rings expanding from an orb: the "looking for devices" and
// "ready to receive" animation. It sits on a card (theme.surface).
func radar(size int, iconName string) giode.Element {
	rings := giode.Canvas(giode.Styles{Width: prop.Width(size), Height: prop.Height(size)}, func(p *giode.Painter) {
		c := p.Width() / 2
		const stroke = 2
		minR, maxR := c*0.3, c-stroke
		t := float64(time.Now().UnixNano()) / 1e9 / radarPeriod
		for i := 0; i < radarRings; i++ {
			_, phase := math.Modf(t + float64(i)/radarRings)
			// Fade in leaving the orb and out at the edge, so no ring
			// appears or disappears abruptly.
			level := math.Sin(math.Pi * phase)
			shade := radarShades[int(level*radarSteps+0.5)]
			p.Arc(c, c, minR+float32(phase)*(maxR-minR), stroke, 0, 360, shade)
		}
		kit.RequestFrame(p.Context())
	})
	orbSize := size * 30 / 100
	return giode.Box(giode.Styles{Width: prop.Width(size), Height: prop.Height(size), Align: giode.AlignCenter, Justify: giode.JustifyCenter},
		giode.Box(giode.Styles{Position: giode.PositionAbsolute, Top: giode.Flush, Left: giode.Flush}, rings),
		orb(iconName, orbSize),
	)
}
