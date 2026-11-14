package wallpaper

import (
	"math"

	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

func bauhausShape(n noise, i, j int, u, v float64) bool {
	cu, cv := math.Round(n.toss(i, j, 7)), math.Round(n.toss(i, j, 8))
	switch int(n.toss(i, j, 6) * 10) {
	case 0:
		return math.Hypot(u-cu, v-cv) < 1
	case 1:
		return math.Hypot(u-0.5, v-cv) < 0.5
	case 2:
		return math.Hypot(u-0.5, v-0.5) < 0.36
	case 3:
		return (u > v) == (cu > 0.5)
	case 4:
		r := math.Hypot(u-0.5, v-0.5)
		return r > 0.2 && r < 0.4
	case 5:
		return math.Hypot(u-cu, v-cv) < 0.5 || math.Hypot(u-1+cu, v-1+cv) < 0.5
	}
	return false
}

func bauhaus(colors theme.Colors, n noise, at area) painter {
	base, _ := depths(colors)
	paints := []rgb{pick(colors, "accent", "blue"), analogous(colors), pick(colors, "foreground")}
	field := wash(colors)
	cell := 160 * at.unit
	return supersampled(func(x, y float64) rgb {
		i, j := int(x/cell), int(y/cell)
		u, v := x/cell-float64(i), y/cell-float64(j)
		paint := func(salt int) rgb {
			return mix(base, paints[int(n.toss(i, j, salt)*float64(len(paints)))], 0.55)
		}
		back := field(x/at.width, y/at.height)
		if n.toss(i, j, 5) < 0.25 {
			back = mix(back, paint(10), 0.18)
		}
		if bauhausShape(n, i, j, u, v) {
			return paint(11)
		}
		return back
	}, at)
}
