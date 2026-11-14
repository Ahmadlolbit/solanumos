package wallpaper

import (
	"math"

	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

func truchet(colors theme.Colors, n noise, at area) painter {
	accent := pick(colors, "accent", "blue")
	second := analogous(colors)
	field := wash(colors)
	cell, line := 90*at.unit, 1.8*at.unit
	return supersampled(func(x, y float64) rgb {
		i, j := int(x/cell), int(y/cell)
		u, v := x/cell-float64(i), y/cell-float64(j)
		turned := 0
		if n.toss(i, j, 4) < 0.5 {
			u, turned = 1-u, 1
		}
		near, far := math.Hypot(u, v), math.Hypot(1-u, 1-v)
		ground := field(x/at.width, y/at.height)
		if (near < 0.5 || far < 0.5) == ((i+j+turned)%2 == 0) {
			ground = mix(ground, accent, 0.07)
		}
		if math.Min(math.Abs(near-0.5), math.Abs(far-0.5))*cell < line {
			return mix(ground, mix(accent, second, n.value(x/at.height*3+0.5, y/at.height*3+0.5)), 0.7)
		}
		return ground
	}, at)
}
