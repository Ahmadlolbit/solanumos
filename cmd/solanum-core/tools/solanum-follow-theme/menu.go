package followtheme

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/burst"
	"github.com/Ahmadlolbit/solanumos/internal/follow"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
)

func opaque(hex string) color.NRGBA {
	channels := follow.Channels(hex)
	return color.NRGBA{R: uint8(channels[0]), G: uint8(channels[1]), B: uint8(channels[2]), A: 255}
}

func solid(paint color.Color) ([]byte, error) {
	picture := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	draw.Draw(picture, picture.Bounds(), image.NewUniform(paint), image.Point{}, draw.Src)
	var out bytes.Buffer
	err := png.Encode(&out, picture)
	return out.Bytes(), err
}

func menuFace() string {
	for _, size := range []string{"16", "20", "28"} {
		if !fsutil.IsFile(filepath.Join(grubDir, "jetbrains-"+size+".pf2")) {
			return "DejaVu Sans Mono"
		}
	}
	return "JetBrains Mono"
}

func picture(name string, data []byte, err error) error {
	if err == nil {
		_, err = refresh(filepath.Join(grubDir, name), string(data))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err
}

func menu(current theme) bool {
	logo, err := burst.PNG(2, opaque(current.accent))
	drawn := picture("logo.png", logo, err) == nil
	selected := current.background
	square, err := solid(opaque(current.accent))
	if picture("select_c.png", square, err) != nil {
		drawn, selected = false, current.accent
	}
	_, err = render(filepath.Join(templates, "theme.txt"), filepath.Join(grubDir, "theme.txt"), "BACKGROUND="+current.background,
		"FOREGROUND="+current.foreground, "ACCENT="+current.accent, "MUTED="+current.muted, "FONT="+menuFace(), "SELECTED="+selected)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return drawn && err == nil
}
