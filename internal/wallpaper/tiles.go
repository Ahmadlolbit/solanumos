package wallpaper

import (
	"math"

	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

var spokes = []point{{math.Cos(math.Pi / 6), -0.5}, {-math.Cos(math.Pi / 6), -0.5}, {0, 1}}

type tilePaints struct{ base, deep, accent, second, light rgb }

func (paints tilePaints) face(ground rgb, kind, chance float64, which int) rgb {
	from, to, amount := [3]rgb{ground, ground, ground}, [3]rgb{paints.light, paints.light, paints.deep}, [3]float64{0.07, 0.025, 0.32}
	switch {
	case kind > 0.66 || chance < 0.012:
		from, to, amount = [3]rgb{paints.base, paints.base, paints.deep}, [3]rgb{paints.accent, paints.accent, paints.accent}, [3]float64{0.62, 0.40, 0.22}
	case kind < 0.33:
		from, to, amount = [3]rgb{ground, ground, paints.deep}, [3]rgb{paints.second, paints.second, paints.second}, [3]float64{0.24, 0.13, 0.10}
	}
	return mix(from[which], to[which], amount[which])
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
	paints := tilePaints{base, deep, pick(colors, "accent", "blue"), analogous(colors), pick(colors, "foreground")}
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
