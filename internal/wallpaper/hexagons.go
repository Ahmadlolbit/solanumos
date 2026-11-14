package wallpaper

import (
	"math"

	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

var hexLevels = []float64{0.02, 0.06, 0.12, 0.22, 0.40}

func hexagons(colors theme.Colors, n noise, at area) painter {
	_, deep := depths(colors)
	accent := pick(colors, "accent", "blue")
	second := analogous(colors)
	field := wash(colors)
	size, line := 52*at.unit, 1.8*at.unit
	inner := size * math.Sqrt(3) / 2
	left, top, right, bottom := hexBounds(at, size)
	levels := remembered(left, top, right, bottom, func(q, r int) float64 {
		cx, cy := size*math.Sqrt(3)*(float64(q)+float64(r)/2)/at.height, size*1.5*float64(r)/at.height
		return math.Floor(smooth(0.30, 0.74, n.fbm(cx*2.0+1.1, cy*2.0+4.4, 4)) * 4.99)
	})
	return supersampled(func(x, y float64) rgb {
		q, r, dx, dy := hexagon(x, y, size)
		ground := field(x/at.width, y/at.height)
		level := levels.at(q, r)
		tint := mix(second, accent, level/4)
		if hexEdge(dx, dy, inner) < line {
			return mix(ground, deep, 0.35)
		}
		return mix(ground, tint, hexLevels[int(level)])
	}, at)
}
