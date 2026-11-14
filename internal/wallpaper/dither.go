package wallpaper

import (
	"math"

	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

func dither(colors theme.Colors, n noise, at area) painter {
	accent := pick(colors, "accent", "blue")
	second := analogous(colors)
	field := wash(colors)
	dot := max(2, int(math.Round(10*at.unit)))
	center := func(px, py int) (float64, float64) {
		return (float64(px) + 0.5) * float64(dot) / at.height, (float64(py) + 0.5) * float64(dot) / at.height
	}
	levels := remembered(0, 0, int(at.width)/dot+1, int(at.height)/dot+1, func(px, py int) float64 {
		cx, cy := center(px, py)
		return smooth(0.40, 0.80, n.fbm(cx*1.4+3.3, cy*1.4+6.6, 5)) * 0.55
	})
	return shaded(func(x, y float64) rgb {
		px, py := int(math.Round(x*at.width))/dot, int(math.Round(y*at.height))/dot
		cx, cy := center(px, py)
		ground := field(x, y)
		if levels.at(px, py)*16 > bayer[(py&3)*4+px&3]+0.5 {
			return mix(ground, mix(accent, second, smooth(0.25, 0.75, n.value(cx*1.2+9.1, cy*1.2+0.7))), 0.7)
		}
		return ground
	}, at)
}
