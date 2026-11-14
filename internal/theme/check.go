package theme

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	TextRatio   = 4.5
	AccentRatio = 3.0
	liftStep    = 5
)

var (
	ErrNoColours = errors.New("its colors.toml has no colours")
	neededHues   = []string{"red", "green", "yellow", "blue", "magenta", "cyan"}
	neededKeys   = append([]string{"background", "foreground", "accent"}, neededHues...)
	defaultLooks = map[bool]Colors{
		true:  {"background": "#1e1e2e", "foreground": "#cdd6f4"},
		false: {"background": "#eff1f5", "foreground": "#4c4f69"},
	}
)

type checking struct {
	colors Colors
	bad    map[string]string
	fixes  []string
}

func Checked(text string, light bool) (Colors, []string, error) {
	check := checking{colors: Parse(text), bad: map[string]string{}}
	check.noteBad(text)
	if len(check.colors.hexes()) == 0 {
		return nil, check.fixes, ErrNoColours
	}
	for _, pair := range aliases {
		check.colors.fallback(pair[0], pair[1], "")
	}
	check.workBase()
	check.workMode(light)
	check.workHues()
	check.workAccent()
	check.fixes = append(check.fixes, check.colors.Readable()...)
	check.colors.Resolve()
	return check.colors, check.fixes, nil
}

func (colors Colors) hexes() []string {
	var found []string
	for key, value := range colors {
		if IsHex(value) {
			found = append(found, key)
		}
	}
	return found
}

func (check *checking) noteBad(text string) {
	for _, line := range strings.Split(strings.TrimPrefix(text, "\ufeff"), "\n") {
		key, value, ok := assignment(line)
		if !ok || check.colors[key] != "" || !(strings.HasPrefix(value, "#") || colourKey(key)) {
			continue
		}
		check.bad[key] = value
		if !slices.Contains(neededKeys, key) {
			check.fixes = append(check.fixes, fmt.Sprintf("%s %q isn't a colour; it was left out.", key, value))
		}
	}
}

func (check *checking) worked(key, value, how string) {
	check.colors[key] = value
	problem := key + " was missing"
	if bad, found := check.bad[key]; found {
		problem = fmt.Sprintf("%s %q isn't a colour", key, bad)
	}
	check.fixes = append(check.fixes, fmt.Sprintf("%s; it's %s now, %s.", problem, value, how))
}

func (check *checking) workBase() {
	colors := check.colors
	dark := !strings.EqualFold(colors["mode"], "light")
	switch {
	case colors["background"] == "" && colors["foreground"] != "" && colors["mode"] == "":
		dark = !Dark(colors["foreground"])
	case colors["foreground"] == "" && colors["background"] != "":
		dark = Dark(colors["background"])
	}
	for _, key := range []string{"background", "foreground"} {
		if colors[key] == "" {
			check.worked(key, defaultLooks[dark][key], "the default theme's")
		}
	}
}

func (check *checking) workMode(light bool) {
	if check.colors["mode"] != "" {
		return
	}
	if light {
		check.colors["mode"] = "light"
		return
	}
	check.colors["mode"] = modeOf(check.colors["background"])
	if check.colors["mode"] == "light" {
		check.fixes = append(check.fixes, "mode was missing; it's light now, from the background.")
	}
}

func (check *checking) workHues() {
	for _, hue := range hueAngles {
		if slices.Contains(neededHues, hue.key) && check.colors[hue.key] == "" {
			base, _ := hueTargets(hue.key)
			check.worked(hue.key, hueFor(check.colors["background"], hue.at, hueChroma, base), "worked out from the background")
		}
	}
}

func (check *checking) workAccent() {
	if check.colors["accent"] == "" {
		check.worked("accent", check.colors["blue"], "the theme's blue")
	}
}

func (colors Colors) Readable() []string {
	fixes := colors.lift("foreground", TextRatio)
	fixes = append(fixes, colors.ReadableAccent()...)
	return append(fixes, colors.clearSelection()...)
}

func (colors Colors) ReadableAccent() []string {
	return colors.lift("accent", AccentRatio)
}

func (colors Colors) lift(key string, ratio float64) []string {
	value, background := colors[key], colors["background"]
	before := Contrast(value, background)
	if !IsHex(value) || !IsHex(background) || before >= ratio {
		return nil
	}
	target := black
	if Dark(background) {
		target = white
	}
	moved := value
	for percent := liftStep; percent <= 100 && Contrast(moved, background) < ratio; percent += liftStep {
		moved = Mix(value, target, percent)
	}
	colors[key] = moved
	return []string{fmt.Sprintf("The %s %s was too faint on the background (%.1f:1); it's %s now (%.1f:1).", key, value, before, moved, Contrast(moved, background))}
}

func (colors Colors) clearSelection() []string {
	foreground, background := colors["foreground"], colors["background"]
	selection := cmp.Or(colors["selection"], colors["selection_background"], colors["muted"], colors["color8"], colors["dark_foreground"], colors.shade(20))
	before := Contrast(foreground, selection)
	if !IsHex(selection) || before >= TextRatio {
		return nil
	}
	moved := selection
	for percent := liftStep; percent <= 100 && Contrast(foreground, moved) < TextRatio; percent += liftStep {
		moved = Mix(selection, background, percent)
	}
	if colors["selection_background"] == selection {
		colors["selection_background"] = moved
	}
	colors["selection"] = moved
	return []string{fmt.Sprintf("Text on the selection %s was too faint (%.1f:1); the selection is %s now (%.1f:1).", selection, before, moved, Contrast(foreground, moved))}
}
