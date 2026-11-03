package ui

import prop "github.com/xmhrnx/giode/styles/properties"

func dpW(v int) prop.Width  { return prop.Width(v) }
func dpH(v int) prop.Height { return prop.Height(v) }

// pct converts a fraction to a WidthPct (at least 1, so the bar shows).
func pct(v float32) prop.WidthPct {
	p := int(v*100 + 0.5)
	if p < 1 {
		p = 1
	}
	if p > 100 {
		p = 100
	}
	return prop.WidthPct(p)
}
