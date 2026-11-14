package wallpaper

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/burst"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/hypr"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

const largest = 8192

func Size() (int, int) {
	width, height, ok := hypr.LargestScreen()
	if !ok || width > largest || height > largest {
		return Width, Height
	}
	return width, height
}

func CurrentColors() theme.Colors {
	data, _ := os.ReadFile(filepath.Join(env.ThemeState(), "theme", "colors.toml"))
	colors := theme.Parse(string(data))
	colors.Resolve()
	return colors
}

func folder(name string) string {
	return filepath.Join(env.ThemeState(), name)
}

func keep(path string, data []byte) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return "", err
	}
	return path, atomic.Write(path, data, 0o666)
}

func Another(colors theme.Colors) (string, error) {
	name, seed := Random()
	width, height := Size()
	data, err := JPEG(Draw(name, colors, seed, width, height))
	if err != nil {
		return "", err
	}
	return keep(filepath.Join(folder("drawings"), fmt.Sprintf("%s-%d.jpg", name, seed)), data)
}

func IsLogo(path string) bool {
	logos, err := fsutil.RealPath(folder("logo"))
	if err != nil {
		return false
	}
	dir, err := fsutil.RealPath(filepath.Dir(path))
	return err == nil && dir == logos
}

func logoImage(colors theme.Colors, width, height int) *image.Paletted {
	paints := [3]rgb{pick(colors, "background"), pick(colors, "foreground"), pick(colors, "accent", "blue")}
	palette := make(color.Palette, len(paints))
	for index, c := range paints {
		palette[index] = color.RGBA{clamp(c.r), clamp(c.g), clamp(c.b), 255}
	}
	img := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	place, grid := logoAt(width, height), logoGrid(burst.Still())
	size := place.scale / logoSub
	for y := max(0, place.top); y < min(height, place.top+logoCells*size); y++ {
		for x := max(0, place.left); x < min(width, place.left+logoCells*size); x++ {
			img.Pix[y*img.Stride+x] = grid[(y-place.top)/size*logoCells+(x-place.left)/size]
		}
	}
	return img
}

func LogoStill(colors theme.Colors) (string, error) {
	width, height := Size()
	name := fmt.Sprintf("%08x-%dx%d.png", Seed(colors["background"]+colors["foreground"]+colors["accent"]), width, height)
	path := filepath.Join(folder("logo"), name)
	if fsutil.IsFile(path) {
		return path, nil
	}
	var out bytes.Buffer
	if err := png.Encode(&out, logoImage(colors, width, height)); err != nil {
		return "", err
	}
	return keep(path, out.Bytes())
}

func Tidy(kept string) {
	for _, entry := range fsutil.Entries(filepath.Dir(kept)) {
		if path := filepath.Join(filepath.Dir(kept), entry.Name()); path != kept {
			os.Remove(path)
		}
	}
}
