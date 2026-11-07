package ui

import (
	"github.com/xmhrnx/giode"
)

func (s *sidebar) deviceCard(ps giode.PressState) giode.Element {
	self := s.u.eng.Self()
	st := s.u.eng.Settings.Get()
	bg := colSurface2
	if ps.Hovered {
		bg = colSurface3
	}
	status, dot := "Visible nearby", "theme.success"
	if !st.Discoverable {
		status, dot = "Hidden", "theme.muted"
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10, Padding: giode.UniformInset(14), BorderRadius: 16, Background: giode.Color(bg)},
		giode.Text("THIS DEVICE", stOverline),
		giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			orb("laptop", 34),
			giode.Stack(giode.Styles{Gap: 2, FlexGrow: 1},
				giode.Text(self.Name, giode.Styles{FontFamily: fontText, FontWeight: giode.FontWeightSemiBold, FontSize: 13.5, Color: "theme.text", MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}),
				giode.Text(self.IP, stSmall),
			),
		),
		giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter},
			giode.Box(giode.Styles{Width: 7, Height: 7, BorderRadius: 4, Background: giode.Color(dot)}),
			giode.Text(status, stSmall),
		),
	)
}
