package wallpaper

import (
	"math"

	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

func triangleAt(x, y, side, tall float64) (float64, float64, int, float64) {
	b := y / tall
	a := x/side - b/2
	ia, ib := math.Floor(a), math.Floor(b)
	fa, fb := a-ia, b-ib
	half, edge := 0, math.Min(math.Min(fa, fb), 1-fa-fb)*tall
	if fa+fb > 1 {
		half, edge = 1, math.Min(math.Min(1-fa, 1-fb), fa+fb-1)*tall
	}
	return ia, ib, half, edge
}

func triangles(colors theme.Colors, n noise, at area) painter {
	base, deep := depths(colors)
	accent := pick(colors, "accent", "blue")
	second := analogous(colors)
	side, line := 120*at.unit, 0.9*at.unit
	tall := side * math.Sqrt(3) / 2
	shades := remembered(int(-(at.height+2)/tall/2)-2, -4, int((at.width+2)/side)+2, 2*int((at.height+2)/tall)+4, func(i, j int) float64 {
		ca, cb := float64(i)+float64(1+j&1)/3, float64(j>>1)+float64(1+j&1)/3
		cx, cy := (ca+cb/2)*side/at.height, cb*tall/at.height
		return smooth(0.28, 0.72, n.fbm(cx*2.4+2.2, cy*2.4+8.1, 4)+(n.toss(i, j, 3)-0.5)*0.08)
	})
	return supersampled(func(x, y float64) rgb {
		ia, ib, half, edge := triangleAt(x, y, side, tall)
		t := shades.at(int(ia), int(ib)*2+half)
		face := mix(mix(deep, second, 0.12), base, t*2)
		if t > 0.5 {
			face = mix(base, mix(base, accent, 0.4), (t-0.5)*2)
		}
		if edge < line {
			return mix(face, deep, 0.25)
		}
		return face
	}, at)
}
