package wallpaper

import (
	"image"
	"math/rand/v2"
	"slices"

	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

type area struct{ width, height, unit float64 }

type painter func(y int, row []rgb)

type finisher func(img *image.RGBA, row []rgb, y int)

type style func(colors theme.Colors, n noise, at area) painter

var (
	Styles = []string{"contours", "cubes", "tiles", "dots", "hexagons", "triangles", "truchet", "dither", "bauhaus", "burst"}
	styles = map[string]style{
		"cubes": cubes, "tiles": tiles, "dots": dots, "hexagons": hexagons, "triangles": triangles,
		"truchet": truchet, "dither": dither, "bauhaus": bauhaus, "burst": burstStyle,
	}
)

func sized(width, height int) area {
	return area{float64(width), float64(height), float64(height) / Height}
}

func Known(name string) bool {
	return slices.Contains(Styles, name)
}

func Draw(name string, colors theme.Colors, seed uint32, width, height int) *image.RGBA {
	paint, ok := styles[name]
	if !ok {
		return Contours(colors, seed, width, height)
	}
	return render(paint(colors, noise{seed}, sized(width, height)), grained(noise{seed ^ 0x9e3779b9}), width, height)
}

func Random() (string, uint32) {
	return Styles[rand.IntN(len(Styles))], rand.Uint32()
}
