package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/look"
	"github.com/Ahmadlolbit/solanumos/internal/wallpaper"
)

var appNames = map[string]string{
	"alacritty.toml": "Alacritty", "btop.theme": "btop", "fastfetch.jsonc": "fastfetch", "firefox.css": "Firefox", "foot.ini": "foot",
	"glass.lua": "Hyprland", "gtk3.css": "GTK apps", "gtk4.css": "GTK apps", "hyprland.lua": "Hyprland", "hyprlock.conf": "the lock screen",
	"kde.colors": "KDE apps", "kitty.conf": "kitty", "look.toml": "the bar", "mako.ini": "notifications", "nvim.lua": "Neovim",
	"walker.css": "the launcher",
}

type leftOut struct {
	file string
	why  string
}

func (left leftOut) String() string {
	return "Left out " + left.file + ": " + left.why + "."
}

func missed(app string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't give %s the theme: %s. The rest changed.\n", app, reason(err))
	}
}

func appOf(file string) string {
	if app, found := appNames[file]; found {
		return app
	}
	return file
}

func nowhere(path string) string {
	switch {
	case fsutil.IsLink(path) && !fsutil.Exists(path):
		return "it links to nowhere"
	case fsutil.IsDir(path):
		return "it's a folder"
	}
	return "it links to something other than a file"
}

func cantUse(notes []leftOut) error {
	for _, note := range notes {
		if note.file == "colors.toml" || note.file == "its folder" {
			return fmt.Errorf("%s can't be used (%s)", strings.Replace(note.file, "colors.toml", "its colors.toml", 1), note.why)
		}
	}
	return errors.New("it has no colors.toml")
}

func stage(name, next string) ([]leftOut, error) {
	if err := os.RemoveAll(next); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := os.MkdirAll(next, 0o777); err != nil {
		return nil, err
	}
	if shipped := filepath.Join(systemRoot, "themes", name); fsutil.IsDir(shipped) {
		if err := fsutil.StageTree(shipped, next); err != nil {
			return nil, err
		}
	}
	own := filepath.Join(userThemes(), name)
	if !fsutil.IsDir(own) {
		return nil, nil
	}
	return stageOwn(own, next)
}

func stageOwn(own, next string) ([]leftOut, error) {
	entries, err := os.ReadDir(own)
	if err != nil {
		return []leftOut{{"its folder", reason(err)}}, nil
	}
	var notes []leftOut
	for _, entry := range entries {
		left, err := stageEntry(own, next, entry.Name())
		if err != nil {
			return nil, err
		}
		notes = append(notes, left...)
	}
	return notes, nil
}

func stageEntry(own, next, base string) ([]leftOut, error) {
	source, target := filepath.Join(own, base), filepath.Join(next, base)
	switch {
	case strings.HasPrefix(base, "."):
		return nil, nil
	case strings.HasSuffix(base, ".lua") || strings.Contains(denied, " "+base+" "):
		fmt.Fprintf(os.Stderr, "ignored %s from %s: your themes can't bring code or settings that run commands\n", base, own)
		return nil, nil
	case base == "backgrounds":
		return stageImages(source, target)
	case !slices.Contains(ownFiles, base):
		return nil, nil
	}
	return stageOne(source, target, base)
}

func stageOne(source, target, name string) ([]leftOut, error) {
	if !fsutil.IsFile(source) {
		return []leftOut{{name, nowhere(source)}}, nil
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return []leftOut{{name, reason(err)}}, nil
	}
	if err := os.Remove(target); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	return nil, atomic.Stage(target, data, 0o644)
}

func stageImages(source, target string) ([]leftOut, error) {
	if !fsutil.IsDir(source) {
		return []leftOut{{"backgrounds", nowhere(source)}}, nil
	}
	var notes []leftOut
	for _, entry := range fsutil.Entries(fsutil.Followed(source)) {
		if !fsutil.ImageName(entry.Name()) {
			continue
		}
		if err := os.MkdirAll(target, 0o777); err != nil {
			return nil, err
		}
		left, err := stageOne(filepath.Join(source, entry.Name()), filepath.Join(target, entry.Name()), "backgrounds/"+entry.Name())
		if err != nil {
			return nil, err
		}
		notes = append(notes, left...)
	}
	return notes, nil
}

func checkPictures(next string) []leftOut {
	dir := filepath.Join(next, "backgrounds")
	var notes []leftOut
	for _, entry := range fsutil.Entries(dir) {
		path := filepath.Join(dir, entry.Name())
		if fsutil.ImageName(entry.Name()) && !wallpaper.IsPicture(path) {
			os.Remove(path)
			notes = append(notes, leftOut{"backgrounds/" + entry.Name(), "it isn't a picture"})
		}
	}
	return notes
}

func defaultIcons(mode string) string {
	shipped, _ := look.Default(mode)
	first, _, _ := strings.Cut(fsutil.Optional(filepath.Join(systemRoot, "themes", shipped, "icons.theme")), "\n")
	if icons := strings.TrimSpace(first); iconName.MatchString(icons) && iconInstalled(icons) {
		return icons
	}
	if mode == "dark" {
		return "Yaru-dark"
	}
	return "Yaru"
}

func checkIcons(next, mode string) []string {
	file := filepath.Join(next, "icons.theme")
	first, _, _ := strings.Cut(fsutil.Optional(file), "\n")
	named := strings.TrimSpace(strings.TrimPrefix(first, "\ufeff"))
	if named == "" || iconName.MatchString(named) && iconInstalled(named) {
		return nil
	}
	fallback := defaultIcons(mode)
	err := os.Remove(file)
	if err == nil {
		err = atomic.Stage(file, []byte(fallback+"\n"), 0o644)
	}
	if err != nil {
		return []string{leftOut{"icons.theme", reason(err)}.String()}
	}
	return []string{fmt.Sprintf("icons.theme names %s, which isn't installed; it's %s now, the default theme's.", named, fallback)}
}

func checkPreview(name string) []string {
	for _, dir := range []string{filepath.Join(userThemes(), name), filepath.Join(systemRoot, "themes", name)} {
		for _, kind := range previewKinds {
			file := filepath.Join(dir, "preview."+kind)
			if !fsutil.IsFile(file) {
				continue
			}
			if wallpaper.IsPicture(file) {
				return nil
			}
			return []string{"preview." + kind + " isn't a picture; the picker shows the first background instead."}
		}
	}
	return nil
}

func checkFiles(name, next, mode string, notes []leftOut) []string {
	var lines []string
	for _, note := range append(notes, checkPictures(next)...) {
		lines = append(lines, note.String())
	}
	lines = append(lines, checkIcons(next, mode)...)
	return append(lines, checkPreview(name)...)
}
