package wallpaper

import (
	"math"

	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

var spokes = []point{{math.Cos(math.Pi / 6), -0.5}, {-math.Cos(math.Pi / 6), -0.5}, {0, 1}}

type tilePaints struct {
	bright                    [3]rgb
	second, light, deep, dusk rgb
}

func (paints tilePaints) face(ground rgb, kind, chance float64, which int) rgb {
	switch {
	case kind > 0.66 || chance < 0.012:
		return paints.bright[which]
	case kind < 0.33:
		if which == 2 {
			return paints.dusk
		}
		return mix(ground, paints.second, [2]float64{0.24, 0.13}[which])
	}
	if which == 2 {
		return mix(ground, paints.deep, 0.32)
	}
	return mix(ground, paints.light, [2]float64{0.07, 0.025}[which])
}

func faceOf(dx, dy float64) int {
	angle := math.Atan2(-dy, dx) * 180 / math.Pi
	switch {
	case angle >= 30 && angle < 150:
		return 0
	case angle >= 150 || angle < -90:
		return 1
	}
	return 2
}

func spokeEdge(edge, dx, dy float64) float64 {
	for _, s := range spokes {
		if dx*s.x+dy*s.y > 0 {
			edge = math.Min(edge, math.Abs(dx*s.y-dy*s.x))
		}
	}
	return edge
}

func tiles(colors theme.Colors, n noise, at area) painter {
	base, deep := depths(colors)
	accent, second := pick(colors, "accent", "blue"), analogous(colors)
	paints := tilePaints{[3]rgb{mix(base, accent, 0.62), mix(base, accent, 0.40), mix(deep, accent, 0.22)}, second, pick(colors, "foreground"), deep, mix(deep, second, 0.10)}
	field := wash(colors)
	size, line := 84*at.unit, 0.8*at.unit
	inner := size * math.Sqrt(3) / 2
	left, top, right, bottom := hexBounds(at, size)
	kinds := remembered(left, top, right, bottom, func(q, r int) float64 { return n.fbm(float64(q)*0.11+5.1, float64(r)*0.11+2.3, 3) })
	chances := remembered(left, top, right, bottom, func(q, r int) float64 { return n.toss(q, r, 2) })
	return supersampled(func(x, y float64) rgb {
		q, r, dx, dy := hexagon(x, y, size)
		ground := field(x/at.width, y/at.height)
		face := paints.face(ground, kinds.at(q, r), chances.at(q, r), faceOf(dx, dy))
		if spokeEdge(hexEdge(dx, dy, inner), dx, dy) < line {
			return mix(face, deep, 0.35)
		}
		return face
	}, at)
}
