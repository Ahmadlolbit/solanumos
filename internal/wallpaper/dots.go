package wallpaper

import (
	"math"

	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

type dotGrid struct {
	columns int
	radius  []float64
	tint    []rgb
}

func dotsLaid(colors theme.Colors, n noise, at area, cell float64) dotGrid {
	base, _ := depths(colors)
	accent, second, light := pick(colors, "accent", "blue"), analogous(colors), pick(colors, "foreground")
	columns, rows := int(at.width/cell)+1, int(at.height/cell)+1
	grid := dotGrid{columns, make([]float64, columns*rows), make([]rgb, columns*rows)}
	for j := range rows {
		for i := range columns {
			cx, cy := (float64(i)+0.5)*cell/at.height, (float64(j)+0.5)*cell/at.height
			swell := smooth(0.42, 0.78, n.fbm(cx*1.6+7.3, cy*1.6+2.9, 4))
			grid.radius[j*columns+i] = cell * (0.07 + 0.32*swell)
			grid.tint[j*columns+i] = mix(mix(base, light, 0.20), mix(accent, second, n.value(cx*2.2+1.7, cy*2.2+4.1)), smooth(0.05, 0.6, swell)*0.85)
		}
	}
	return grid
}

func dots(colors theme.Colors, n noise, at area) painter {
	field := wash(colors)
	cell := 40 * at.unit
	grid := dotsLaid(colors, n, at, cell)
	return supersampled(func(x, y float64) rgb {
		i, j := int(x/cell), int(y/cell)
		if math.Hypot(x-(float64(i)+0.5)*cell, y-(float64(j)+0.5)*cell) < grid.radius[j*grid.columns+i] {
			return grid.tint[j*grid.columns+i]
		}
		return field(x/at.width, y/at.height)
	}, at)
}
