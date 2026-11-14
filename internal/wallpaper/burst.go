package wallpaper

import (
	"math"

	"github.com/Ahmadlolbit/solanumos/internal/burst"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

const (
	logoMargin = 6
	logoSub    = 4
	logoSpan   = burst.Size + 2*logoMargin
	logoCells  = logoSpan * logoSub
)

type logoPlace struct{ scale, left, top int }

func logoAt(width, height int) logoPlace {
	scale := logoSub * max(1, int(math.Round(logoSub*float64(height)/Height)))
	return logoPlace{scale, (width-burst.Size*scale)/2 - logoMargin*scale, (height-burst.Size*scale)/2 - logoMargin*scale}
}

func lit(frame burst.Frame, x, y int) bool {
	return x >= 0 && y >= 0 && x < burst.Size && y < burst.Size && frame[y]>>(burst.Size-1-x)&1 == 1
}

func blurred(glow []float64, across bool) []float64 {
	pass := make([]float64, len(glow))
	for at := range glow {
		x, y, sum := at%logoSpan, at/logoSpan, 0.0
		for k := -2; k <= 2; k++ {
			xx, yy := x, y+k
			if across {
				xx, yy = x+k, y
			}
			if xx >= 0 && yy >= 0 && xx < logoSpan && yy < logoSpan {
				sum += glow[yy*logoSpan+xx]
			}
		}
		pass[at] = sum / 5
	}
	return pass
}

func glowOf(frame burst.Frame) []float64 {
	glow := make([]float64, logoSpan*logoSpan)
	for at := range glow {
		if lit(frame, at%logoSpan-logoMargin, at/logoSpan-logoMargin) {
			glow[at] = 1
		}
	}
	for range 2 {
		glow = blurred(blurred(glow, true), false)
	}
	return glow
}

func haloAt(glow []float64, gx, gy, sx, sy int) bool {
	value := func(x, y int) float64 {
		if x < logoSpan && y < logoSpan {
			return glow[y*logoSpan+x]
		}
		return 0
	}
	fx, fy := (float64(sx)+0.5)/logoSub, (float64(sy)+0.5)/logoSub
	g := (value(gx, gy)*(1-fx)+value(gx+1, gy)*fx)*(1-fy) + (value(gx, gy+1)*(1-fx)+value(gx+1, gy+1)*fx)*fy
	px, py := gx*logoSub+sx, gy*logoSub+sy
	return g*1.35*16 > bayer[(py&3)*4+px&3]+0.5
}

func logoGrid(frame burst.Frame) []uint8 {
	glow := glowOf(frame)
	grid := make([]uint8, logoCells*logoCells)
	for at := range grid {
		x, y := at%logoCells, at/logoCells
		gx, gy := x/logoSub, y/logoSub
		switch {
		case lit(frame, gx-logoMargin, gy-logoMargin):
			grid[at] = 1
		case haloAt(glow, gx, gy, x%logoSub, y%logoSub):
			grid[at] = 2
		}
	}
	return grid
}

func logoPainter(colors theme.Colors, frame burst.Frame, place logoPlace) painter {
	paints := [3]rgb{pick(colors, "background"), pick(colors, "foreground"), pick(colors, "accent", "blue")}
	grid, size := logoGrid(frame), place.scale/logoSub
	return func(y int, row []rgb) {
		cy := (y - place.top) / size
		for x := range row {
			cx := (x - place.left) / size
			row[x] = paints[0]
			if y >= place.top && x >= place.left && cx < logoCells && cy < logoCells {
				row[x] = paints[grid[cy*logoCells+cx]]
			}
		}
	}
}

func burstStyle(colors theme.Colors, _ noise, at area) painter {
	return logoPainter(colors, burst.Still(), logoAt(int(at.width), int(at.height)))
}
