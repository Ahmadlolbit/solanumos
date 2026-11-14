package wallpaper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
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

func Folder(name string) string {
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
	return keep(filepath.Join(Folder("drawings"), fmt.Sprintf("%s-%d.jpg", name, seed)), data)
}

func Tidy(kept string) {
	for _, entry := range fsutil.Entries(filepath.Dir(kept)) {
		if path := filepath.Join(filepath.Dir(kept), entry.Name()); path != kept {
			os.Remove(path)
		}
	}
}
