package wallpaper

import "math"

type side struct{ ax, ay, ex, ey, length float64 }

type plate struct {
	sides                  []side
	turn                   float64
	c                      rgb
	left, right, top, foot int
}

type plates struct {
	list   []plate
	width  int
	height int
}

func (all *plates) fill(corners []point, c rgb) {
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	area := 0.0
	sides := make([]side, len(corners))
	for index, a := range corners {
		b := corners[(index+1)%len(corners)]
		area += a.x*b.y - b.x*a.y
		minX, minY, maxX, maxY = math.Min(minX, a.x), math.Min(minY, a.y), math.Max(maxX, a.x), math.Max(maxY, a.y)
		ex, ey := b.x-a.x, b.y-a.y
		sides[index] = side{a.x, a.y, ex, ey, math.Hypot(ex, ey)}
	}
	all.list = append(all.list, plate{sides, math.Copysign(1, area), c,
		max(0, int(minX)-1), min(all.width-1, int(maxX)+1), max(0, int(minY)-1), min(all.height-1, int(maxY)+1)})
}

func (all *plates) line(a, b point, thickness float64, c rgb) {
	length := math.Hypot(b.x-a.x, b.y-a.y)
	if length == 0 {
		return
	}
	nx, ny := -(b.y-a.y)/length*thickness/2, (b.x-a.x)/length*thickness/2
	all.fill([]point{{a.x + nx, a.y + ny}, {b.x + nx, b.y + ny}, {b.x - nx, b.y - ny}, {a.x - nx, a.y - ny}}, c)
}

func (one plate) cover(row []rgb, y int) {
	py := float64(y) + 0.5
	for x := one.left; x <= one.right; x++ {
		px, outside := float64(x)+0.5, math.Inf(-1)
		for _, s := range one.sides {
			outside = math.Max(outside, one.turn*((px-s.ax)*s.ey-(py-s.ay)*s.ex)/s.length)
		}
		if covered := math.Max(0, math.Min(1, 0.85-outside)); covered > 0 {
			row[x] = mix(row[x], one.c, covered)
		}
	}
}

func (all *plates) painter(ground painter) painter {
	return func(y int, row []rgb) {
		ground(y, row)
		for _, one := range all.list {
			if y >= one.top && y <= one.foot {
				one.cover(row, y)
			}
		}
	}
}
