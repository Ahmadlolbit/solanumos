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

func chosen(current theme) string {
	data, err := solid(opaque(current.accent))
	if err == nil {
		_, err = refresh(filepath.Join(grubDir, "select_c.png"), string(data))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return current.accent
	}
	return current.background
}

func menu(current theme) bool {
	_, err := render(filepath.Join(templates, "theme.txt"), filepath.Join(grubDir, "theme.txt"), "BACKGROUND="+current.background,
		"FOREGROUND="+current.foreground, "ACCENT="+current.accent, "MUTED="+current.muted, "FONT="+menuFace(), "SELECTED="+chosen(current))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err == nil
}
