package theme

import (
	"maps"
	"math"
	"slices"
	"strings"
)

const (
	brightChroma = 0.09
	accentChroma = 0.04
)

var fileOrder = [][]string{
	{"mode"}, {"accent", "selection", "muted"},
	{"background", "dark_background", "darker_background", "lighter_background"},
	{"foreground", "dark_foreground", "light_foreground", "bright_foreground"},
	{"red", "yellow", "orange", "green", "cyan", "blue", "magenta", "brown"},
	{"bright_red", "bright_yellow", "bright_green", "bright_cyan", "bright_blue", "bright_magenta"},
}

func hueOf(color string) (float64, float64) {
	value := floats(color)
	r, g, b := linear(value[0]), linear(value[1]), linear(value[2])
	l := math.Cbrt(0.4122214708*r + 0.5363325363*g + 0.0514459929*b)
	m := math.Cbrt(0.2119034982*r + 0.6806995451*g + 0.1073969566*b)
	s := math.Cbrt(0.0883024619*r + 0.2817188376*g + 0.6299787005*b)
	x := 1.9779984951*l - 2.4285922050*m + 0.4505937099*s
	y := 0.0259040371*l + 0.7827717662*m - 0.8086757660*s
	return math.Mod(math.Atan2(y, x)/degree+360, 360), math.Hypot(x, y)
}

func NearestHue(accent string) string {
	at, chroma := hueOf(accent)
	if chroma < accentChroma {
		return ""
	}
	nearest, distance := "", 360.0
	for _, hue := range hueAngles {
		if gap := math.Min(math.Abs(at-hue.at), 360-math.Abs(at-hue.at)); gap < distance {
			nearest, distance = hue.key, gap
		}
	}
	return nearest
}

func (colors Colors) workHue(key string, at float64, bright bool) {
	base, lifted := hueTargets(key)
	if colors[key] == "" {
		colors[key] = hueFor(colors["background"], at, hueChroma, base)
	}
	if bright && colors["bright_"+key] == "" {
		colors["bright_"+key] = hueFor(colors["background"], at, brightChroma, lifted)
	}
}

func (colors Colors) workShades() {
	background, foreground := colors["background"], colors["foreground"]
	darker, darkest := 0.05, 0.10
	if Dark(background) {
		darker, darkest = 0.25, 0.46
	}
	colors.fallback("dark_background", "", blend(background, black, darker))
	colors.fallback("darker_background", "", blend(background, black, darkest))
	colors.fallback("lighter_background", "", blend(background, foreground, 0.10))
	colors.fallback("dark_foreground", "", firstMix(background, foreground, TextRatio))
	colors.fallback("light_foreground", "", blend(background, foreground, 0.85))
	colors.fallback("bright_foreground", "", foreground)
}

func Make(background, foreground, accent string) Colors {
	colors := Colors{"mode": modeOf(background), "background": background, "foreground": foreground, "accent": accent}
	colors["selection"] = blend(background, foreground, 0.15)
	colors["muted"] = firstMix(background, foreground, MutedRatio)
	if nearest := NearestHue(accent); nearest != "" {
		colors[nearest] = accent
	}
	for _, hue := range hueAngles {
		colors.workHue(hue.key, hue.at, hue.key != "orange")
	}
	colors["brown"] = blend(colors["orange"], black, 0.5)
	colors.workShades()
	return colors
}

func (colors Colors) File() string {
	var groups []string
	for _, group := range fileOrder {
		var lines []string
		for _, key := range group {
			if colors[key] != "" {
				lines = append(lines, key+` = "`+colors[key]+`"`)
			}
		}
		if len(lines) > 0 {
			groups = append(groups, strings.Join(lines, "\n"))
		}
	}
	return strings.Join(groups, "\n\n") + "\n"
}

func FileKeys() []string {
	var keys []string
	for _, group := range fileOrder[1:] {
		keys = append(keys, group...)
	}
	return keys
}

func WithColours(text string, changes map[string]string) string {
	lines := strings.Split(strings.TrimSuffix(strings.TrimPrefix(text, "\ufeff"), "\n"), "\n")
	done := map[string]bool{}
	for index, line := range lines {
		key, _, ok := assignment(line)
		if value, change := changes[key]; ok && change {
			lines[index], done[key] = key+` = "`+value+`"`, true
		}
	}
	for _, key := range slices.Sorted(maps.Keys(changes)) {
		if !done[key] {
			lines = append(lines, key+` = "`+changes[key]+`"`)
		}
	}
	return strings.TrimPrefix(strings.Join(lines, "\n"), "\n") + "\n"
}
