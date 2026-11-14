package wallpaper

import (
	"math"

	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

const tallest = 5

type cubeField struct {
	base, deep, accent, second, light rgb
	n                                 noise
	width, edge, stepX, stepY, rim    float64
}

func (field cubeField) at(i, j, z float64) point {
	return point{field.width/2 + (i-j)*field.stepX, (i+j)*field.stepY - z*field.edge}
}

func (field cubeField) level(i, j int) int {
	rise := field.n.fbm(float64(i)*0.055+3.7, float64(j)*0.055+1.3, 4)
	return max(0, min(tallest, int(math.Round((rise-0.45)*13))))
}

func (field cubeField) top(i, j, level int) rgb {
	top := mix(field.base, field.light, 0.07+0.035*float64(level))
	switch {
	case level == 0:
		top = mix(field.base, field.second, 0.10)
	case level >= tallest-1:
		top = mix(top, field.accent, 0.55)
	}
	if field.n.toss(i, j, 1) < 0.008 && level > 0 {
		top = mix(top, field.accent, 0.75)
	}
	return top
}

func (field cubeField) cube(all *plates, i, j int) {
	level := field.level(i, j)
	fi, fj, z := float64(i), float64(j), float64(level)
	top := field.top(i, j, level)
	if level > 0 {
		all.fill([]point{field.at(fi, fj+1, z), field.at(fi+1, fj+1, z), field.at(fi+1, fj+1, 0), field.at(fi, fj+1, 0)}, mix(top, field.deep, 0.35))
		all.fill([]point{field.at(fi+1, fj, z), field.at(fi+1, fj+1, z), field.at(fi+1, fj+1, 0), field.at(fi+1, fj, 0)}, mix(top, field.deep, 0.62))
	}
	all.fill([]point{field.at(fi, fj, z), field.at(fi+1, fj, z), field.at(fi+1, fj+1, z), field.at(fi, fj+1, z)}, top)
	if level > 0 {
		rim := mix(top, field.light, 0.22)
		all.line(field.at(fi, fj+1, z), field.at(fi+1, fj+1, z), field.rim, rim)
		all.line(field.at(fi+1, fj, z), field.at(fi+1, fj+1, z), field.rim, rim)
	}
}

func cubes(colors theme.Colors, n noise, at area) painter {
	base, deep := depths(colors)
	edge := 60 * at.unit
	field := cubeField{base, deep, pick(colors, "accent", "blue"), analogous(colors), pick(colors, "foreground"), n,
		at.width, edge, edge * math.Sqrt(3) / 2, edge / 2, 1.2 * at.unit}
	all := &plates{width: int(at.width), height: int(at.height)}
	reach := int(at.width/2/field.stepX) + 2
	for v := -2; v <= int((at.height+tallest*edge)/field.stepY)+2; v++ {
		for u := -reach; u <= reach; u++ {
			if (u+v)%2 == 0 {
				field.cube(all, (u+v)/2, (v-u)/2)
			}
		}
	}
	return all.painter(shaded(wash(colors), at))
}
