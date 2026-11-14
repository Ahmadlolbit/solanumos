package wallpaper

import (
	"math"

	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

type point struct{ x, y float64 }

var (
	bayer = [16]float64{0, 8, 2, 10, 12, 4, 14, 6, 3, 11, 1, 9, 15, 7, 13, 5}
	hues  = []string{"red", "orange", "yellow", "green", "cyan", "blue", "magenta"}
)

func wash(colors theme.Colors) func(x, y float64) rgb {
	base, deep := depths(colors)
	return func(x, y float64) rgb {
		return mix(base, deep, smooth(0, 1, (x+y)/2))
	}
}

func (n noise) toss(i, j, salt int) float64 {
	return noise{n.seed ^ uint32(salt)*0x9e3779b9}.lattice(i, j)
}

func supersampled(shade func(x, y float64) rgb, at area) painter {
	k := 3
	if at.unit > 1 {
		k = 2
	}
	side, count := float64(k), float64(k*k)
	return shaded(func(x, y float64) rgb {
		var sum rgb
		for sy := range k {
			for sx := range k {
				c := shade(x*at.width+(float64(sx)+0.5)/side, y*at.height+(float64(sy)+0.5)/side)
				sum = rgb{sum.r + c.r, sum.g + c.g, sum.b + c.b}
			}
		}
		return rgb{sum.r / count, sum.g / count, sum.b / count}
	}, at)
}

func hue(c rgb) (float64, float64) {
	high, low := math.Max(c.r, math.Max(c.g, c.b)), math.Min(c.r, math.Min(c.g, c.b))
	if high == low {
		return 0, 0
	}
	spread, turn := high-low, 0.0
	switch high {
	case c.r:
		turn = math.Mod((c.g-c.b)/spread+6, 6)
	case c.g:
		turn = (c.b-c.r)/spread + 2
	default:
		turn = (c.r-c.g)/spread + 4
	}
	return turn * 60, spread / high
}

func analogous(colors theme.Colors) rgb {
	accent := pick(colors, "accent", "blue")
	angle, strength := hue(accent)
	if strength < 0.2 {
		return mix(accent, pick(colors, "foreground"), 0.5)
	}
	best, gap := pick(colors, "cyan", "green", "accent"), 360.0
	for _, name := range hues {
		c, ok := hex(colors[name])
		if !ok {
			continue
		}
		other, vivid := hue(c)
		distance := math.Abs(math.Mod(other-angle+540, 360) - 180)
		if vivid >= 0.2 && distance >= 20 && distance < gap {
			best, gap = c, distance
		}
	}
	return best
}

func hexagon(x, y, size float64) (int, int, float64, float64) {
	q, r := (math.Sqrt(3)/3*x-y/3)/size, 2.0/3*y/size
	s := -q - r
	rq, rr, rs := math.Round(q), math.Round(r), math.Round(s)
	dq, dr, ds := math.Abs(rq-q), math.Abs(rr-r), math.Abs(rs-s)
	if dq > dr && dq > ds {
		rq = -rr - rs
	} else if dr > ds {
		rr = -rq - rs
	}
	return int(rq), int(rr), x - size*math.Sqrt(3)*(rq+rr/2), y - size*1.5*rr
}

func hexBounds(at area, size float64) (int, int, int, int) {
	return int(-(at.height+2)/3/size) - 2, -2, int(math.Sqrt(3)/3*(at.width+2)/size) + 2, int(2.0/3*(at.height+2)/size) + 2
}

func hexEdge(dx, dy, inner float64) float64 {
	return inner - math.Max(math.Abs(dx), math.Max(math.Abs(dx*0.5+dy*math.Sqrt(3)/2), math.Abs(dx*0.5-dy*math.Sqrt(3)/2)))
}
