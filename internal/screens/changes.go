package screens

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	positionPattern = regexp.MustCompile(`^-?[0-9]{1,6}x-?[0-9]{1,6}$`)
	modePattern     = regexp.MustCompile(`^([0-9]{1,5}x[0-9]{1,5})(@([0-9]{1,3}(\.[0-9]+)?))?$`)
	rotations       = []string{"0", "90", "180", "270"}
)

func yesNo(value string) bool {
	return value == "yes" || value == "no"
}

func validScale(value string) bool {
	scale, err := strconv.ParseFloat(value, 64)
	return err == nil && scale >= 0.25 && scale <= 4
}

var simple = map[string]struct {
	valid func(string) bool
	takes string
}{
	"position": {positionPattern.MatchString, "a position like 1920x0"},
	"scale":    {validScale, "a scale from 0.25 to 4"},
	"rotation": {func(value string) bool { return slices.Contains(rotations, value) }, "0, 90, 180 or 270"},
	"on":       {yesNo, "yes or no"},
	"main":     {yesNo, "yes or no"},
}

func refresh(mode string) (string, float64) {
	size, rate, _ := strings.Cut(mode, "@")
	hertz, _ := strconv.ParseFloat(rate, 64)
	return size, hertz
}

func MatchMode(modes []string, value string) (string, bool) {
	if value == Preferred {
		return value, true
	}
	parts := modePattern.FindStringSubmatch(value)
	if parts == nil {
		return "", false
	}
	for _, mode := range modes {
		size, hertz := refresh(mode)
		wanted, _ := strconv.ParseFloat(parts[3], 64)
		if size == parts[1] && (parts[3] == "" || math.Abs(hertz-wanted) < 0.05) {
			return mode, true
		}
	}
	return "", false
}

func (profile *Profile) mirror(at int, value string) error {
	if value == "none" {
		value = ""
	}
	if value != "" && (value == profile.Screens[at].Output || profile.Screen(value) < 0) {
		return fmt.Errorf("%s can mirror another screen of these: %s", profile.Screens[at].Output, strings.Join(profile.outputs(), ", "))
	}
	profile.Screens[at].Mirror = value
	return nil
}

func (profile *Profile) change(monitor Monitor, at int, name, value string) error {
	screen := &profile.Screens[at]
	rule, known := simple[name]
	switch {
	case name == "mode":
		mode, ok := MatchMode(Modes(monitor), value)
		if !ok {
			return fmt.Errorf("%s shows none of %s; its modes are %s", screen.Output, value, strings.Join(Modes(monitor), ", "))
		}
		value = mode
	case name == "mirror":
		return profile.mirror(at, value)
	case !known:
		return fmt.Errorf("no setting named %s: mode, position, scale, rotation, on, main or mirror", name)
	case !rule.valid(value):
		return fmt.Errorf("%s takes %s", name, rule.takes)
	}
	if name == "main" && value == "yes" {
		for other := range profile.Screens {
			profile.Screens[other].Main = false
		}
	}
	screen.set(name, value)
	return nil
}

func (profile Profile) outputs() []string {
	names := make([]string, 0, len(profile.Screens))
	for _, screen := range profile.Screens {
		names = append(names, screen.Output)
	}
	return names
}

func Change(profile Profile, monitors []Monitor, args []string) (Profile, error) {
	profile.Screens = slices.Clone(profile.Screens)
	at := -1
	for _, arg := range args {
		name, value, found := strings.Cut(arg, "=")
		if !found {
			if at = profile.Screen(arg); at < 0 {
				return profile, fmt.Errorf("no screen named %s; the screens now are %s", arg, strings.Join(profile.outputs(), ", "))
			}
			continue
		}
		if at < 0 {
			return profile, fmt.Errorf("name a screen before %s", arg)
		}
		if err := profile.change(monitors[at], at, name, value); err != nil {
			return profile, err
		}
	}
	return profile, profile.Check()
}

func (profile Profile) shown(output string) bool {
	at := profile.Screen(output)
	return at >= 0 && profile.Screens[at].On && profile.Screens[at].Mirror == ""
}

func (profile Profile) Check() error {
	if !slices.ContainsFunc(profile.Screens, func(screen Screen) bool { return profile.shown(screen.Output) }) {
		return errors.New("at least one screen must stay on, showing its own picture")
	}
	for _, screen := range profile.Screens {
		if screen.Mirror != "" && screen.On && !profile.shown(screen.Mirror) {
			return fmt.Errorf("%s mirrors %s, which must be on and show its own picture", screen.Output, screen.Mirror)
		}
		if screen.Main && !profile.shown(screen.Output) {
			return fmt.Errorf("the main screen, %s, must be on and show its own picture", screen.Output)
		}
	}
	return nil
}
