package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/follow"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
	"github.com/Ahmadlolbit/solanumos/internal/wallpaper"
)

const (
	previewWidth  = 1280
	previewHeight = 800
	firstPicture  = "1-contours.jpg"
)

var (
	ownCommands = map[string]func([]string) int{"--make": makeTheme, "--copy": copyTheme, "--change": changeTheme, "--palette": showPalette}
	surfaces    = []string{"selection", "dark_background", "darker_background", "lighter_background"}
	hueIcons    = map[string]string{"red": "Yaru-red", "orange": "Yaru", "yellow": "Yaru-yellow", "green": "Yaru-sage", "cyan": "Yaru-prussiangreen", "blue": "Yaru-blue", "magenta": "Yaru-magenta"}
	errExists   = errors.New("is already there")
)

type swatch struct {
	Key      string  `json:"key"`
	Value    string  `json:"value"`
	Contrast float64 `json:"contrast"`
	Against  string  `json:"against"`
	Target   float64 `json:"target,omitempty"`
}

type paletteShown struct {
	Name   string   `json:"name"`
	Mode   string   `json:"mode"`
	Colors []swatch `json:"colors"`
	Fixes  []string `json:"fixes"`
}

func colour(value string) (string, bool) {
	hex, ok := follow.Hex(value)
	return strings.ToLower(hex), ok && len(strings.TrimPrefix(value, "#")) == 6
}

func ownName(argument string) (string, error) {
	name := theme.NameOf(argument)
	switch {
	case !theme.ValidName(name):
		return "", fmt.Errorf("%q can't name a theme; use letters, digits and dashes", argument)
	case fsutil.Exists(filepath.Join(userThemes(), name)) || fsutil.IsLink(filepath.Join(userThemes(), name)) || fsutil.IsDir(filepath.Join(systemRoot, "themes", name)):
		return "", fmt.Errorf("a theme named %s %w; pick another name", theme.Pretty(name), errExists)
	}
	return name, nil
}

func changesFrom(args []string) (map[string]string, error) {
	changes := map[string]string{}
	for _, pair := range args {
		key, value, found := strings.Cut(pair, "=")
		hex, ok := colour(value)
		if !found || !ok || !slices.Contains(theme.FileKeys(), key) {
			return nil, fmt.Errorf("%q isn't a colour change like accent=#ff8800", pair)
		}
		changes[key] = hex
	}
	return changes, nil
}

func iconsFor(colors theme.Colors) string {
	if icons, found := hueIcons[theme.NearestHue(colors["accent"])]; found {
		return icons
	}
	if colors["mode"] == "light" {
		return "Yaru"
	}
	return "Yaru-dark"
}

func contoursOf(name string, colors theme.Colors, width, height int) ([]byte, error) {
	return wallpaper.JPEG(wallpaper.Contours(colors, wallpaper.Seed(name), width, height))
}

func drawPictures(dir, name string, colors theme.Colors) error {
	preview, err := contoursOf(name, colors, previewWidth, previewHeight)
	if err == nil {
		err = atomic.Write(filepath.Join(dir, "preview.jpg"), preview, 0o644)
	}
	if err != nil {
		return err
	}
	width, height := wallpaper.Size()
	first, err := contoursOf(name, colors, width, height)
	if err == nil {
		err = os.MkdirAll(filepath.Join(dir, "backgrounds"), 0o777)
	}
	if err == nil {
		err = atomic.Write(filepath.Join(dir, "backgrounds", firstPicture), first, 0o644)
	}
	return err
}

func writeOwn(name, text string, light, draw bool) error {
	dir := filepath.Join(userThemes(), name)
	colors, _, err := theme.Checked(text, light)
	if err == nil {
		err = os.MkdirAll(dir, 0o777)
	}
	if err == nil {
		err = atomic.Write(filepath.Join(dir, "colors.toml"), []byte(text), 0o644)
	}
	if err == nil && !fsutil.IsFile(filepath.Join(dir, "icons.theme")) {
		err = atomic.Write(filepath.Join(dir, "icons.theme"), []byte(iconsFor(colors)+"\n"), 0o644)
	}
	if err != nil || !draw {
		return err
	}
	return drawPictures(dir, name, colors)
}

func makeTheme(args []string) int {
	if len(args) != 4 {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	var picked [3]string
	for index, value := range args[1:] {
		hex, ok := colour(value)
		if !ok {
			fmt.Fprintf(os.Stderr, "%q isn't a colour like #1e1e2e\n", value)
			return 2
		}
		picked[index] = hex
	}
	name, err := ownName(args[0])
	if err == nil {
		err = writeOwn(name, theme.Make(picked[0], picked[1], picked[2]).File(), false, true)
	}
	return made(name, err)
}

func made(name string, err error) int {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("Made %s in %s\n", theme.Pretty(name), filepath.Join(userThemes(), name))
	return 0
}

func sourceDir(name string) string {
	if own := filepath.Join(userThemes(), name); fsutil.IsFile(filepath.Join(own, "colors.toml")) {
		return own
	}
	return filepath.Join(systemRoot, "themes", name)
}

func themeText(name string) (string, bool, error) {
	if !theme.ValidName(name) {
		return "", false, fmt.Errorf("no theme named %s; see solanum-theme-set --list", name)
	}
	dir := sourceDir(name)
	data, err := os.ReadFile(filepath.Join(dir, "colors.toml"))
	if err != nil {
		return "", false, fmt.Errorf("no theme named %s with colours; see solanum-theme-set --list", name)
	}
	return string(data), fsutil.IsFile(filepath.Join(dir, "light.mode")), nil
}

func linkPictures(from, name string) error {
	dir := filepath.Join(userThemes(), name, "backgrounds")
	for _, picture := range fsutil.Images(fsutil.Followed(filepath.Join(sourceDir(from), "backgrounds"))) {
		if filepath.Base(picture) == firstPicture {
			continue
		}
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return err
		}
		if err := atomic.Symlink(fsutil.Followed(picture), filepath.Join(dir, filepath.Base(picture))); err != nil {
			return err
		}
	}
	return nil
}

func copyTheme(args []string) int {
	if len(args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	from := theme.Slug(args[0])
	text, light, err := themeText(from)
	changes, bad := changesFrom(args[2:])
	name, taken := ownName(args[1])
	if err = errors.Join(err, bad, taken); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if icons := filepath.Join(sourceDir(from), "icons.theme"); fsutil.IsFile(icons) {
		err = writeIcons(name, icons)
	}
	if err == nil {
		err = writeOwn(name, theme.WithColours(text, changes), light, true)
	}
	if err == nil {
		err = linkPictures(from, name)
	}
	return made(name, err)
}

func writeIcons(name, icons string) error {
	data, err := os.ReadFile(icons)
	if err == nil {
		err = os.MkdirAll(filepath.Join(userThemes(), name), 0o777)
	}
	if err == nil {
		err = atomic.Write(filepath.Join(userThemes(), name, "icons.theme"), data, 0o644)
	}
	return err
}

func changeTheme(args []string) int {
	if len(args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	name := theme.Slug(args[0])
	own := filepath.Join(userThemes(), name)
	changes, err := changesFrom(args[1:])
	if err == nil && (!theme.ValidName(name) || !fsutil.IsFile(filepath.Join(own, "colors.toml"))) {
		err = fmt.Errorf("%s isn't a theme of your own; make your own copy first: solanum-theme-set --copy %s NAME", theme.Pretty(name), name)
	}
	text, light, readErr := themeText(name)
	if err = errors.Join(err, readErr); err == nil {
		err = writeOwn(name, theme.WithColours(text, changes), light, fsutil.IsFile(filepath.Join(own, "backgrounds", firstPicture)))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("Changed %s\n", theme.Pretty(name))
	if current, _ := fsutil.Text(filepath.Join(state(), "theme.name")); current == name {
		return set(name)
	}
	return 0
}

func paletteText(args []string) (string, string, bool, error) {
	if len(args) == 3 {
		background, one := colour(args[0])
		foreground, two := colour(args[1])
		accent, three := colour(args[2])
		if one && two && three {
			return "", theme.Make(background, foreground, accent).File(), false, nil
		}
	}
	if len(args) == 0 {
		return "", "", false, errors.New("--palette takes a theme or three colours")
	}
	name := theme.Slug(args[0])
	text, light, err := themeText(name)
	changes, bad := changesFrom(args[1:])
	if err = errors.Join(err, bad); err != nil {
		return "", "", false, err
	}
	return name, theme.WithColours(text, changes), light, nil
}

func swatches(colors theme.Colors) []swatch {
	var found []swatch
	for _, key := range theme.FileKeys() {
		if !theme.IsHex(colors[key]) {
			continue
		}
		against, other := "background", colors["background"]
		if key == "background" || slices.Contains(surfaces, key) {
			against, other = "foreground", colors["foreground"]
		}
		found = append(found, swatch{key, colors[key], math.Round(theme.Contrast(colors[key], other)*10) / 10, against, theme.Target(key)})
	}
	return found
}

func showPalette(args []string) int {
	name, text, light, err := paletteText(args)
	colors, fixes, checkErr := theme.Checked(text, light)
	if err = errors.Join(err, checkErr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	data, err := json.Marshal(paletteShown{name, colors["mode"], swatches(colors), append([]string{}, fixes...)})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(string(data))
	return 0
}
