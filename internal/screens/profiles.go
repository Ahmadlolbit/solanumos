package screens

import (
	"slices"
	"strings"
)

const Preferred = "preferred"

type Screen struct {
	Output      string
	Description string
	Model       string
	Serial      string
	Mode        string
	Position    string
	Scale       string
	Rotation    string
	On          bool
	Main        bool
	Mirror      string
}

type Profile struct {
	Key     string
	Screens []Screen
}

func yes(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func (screen Screen) text() string {
	pairs := [][2]string{
		{"screen", screen.Output}, {"description", screen.Description}, {"model", screen.Model}, {"serial", screen.Serial},
		{"mode", screen.Mode}, {"position", screen.Position}, {"scale", screen.Scale}, {"rotation", screen.Rotation},
		{"on", yes(screen.On)}, {"main", yes(screen.Main)}, {"mirror", screen.Mirror},
	}
	var text strings.Builder
	for _, pair := range pairs {
		text.WriteString(strings.TrimSpace(pair[0] + " = " + pair[1]))
		text.WriteString("\n")
	}
	return text.String()
}

func Format(profiles []Profile) string {
	parts := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		screens := make([]string, 0, len(profile.Screens))
		for _, screen := range profile.Screens {
			screens = append(screens, screen.text())
		}
		parts = append(parts, "["+profile.Key+"]\n"+strings.Join(screens, "\n"))
	}
	return strings.Join(parts, "\n")
}

func (screen *Screen) set(name, value string) {
	fields := map[string]*string{
		"description": &screen.Description, "model": &screen.Model, "serial": &screen.Serial, "mode": &screen.Mode,
		"position": &screen.Position, "scale": &screen.Scale, "rotation": &screen.Rotation, "mirror": &screen.Mirror,
	}
	switch {
	case fields[name] != nil:
		*fields[name] = value
	case name == "on":
		screen.On = value != "no"
	case name == "main":
		screen.Main = value == "yes"
	}
}

func setting(line string) (string, string, bool) {
	name, value, found := strings.Cut(line, "=")
	return strings.TrimSpace(name), strings.TrimSpace(value), found
}

func readLine(profiles []Profile, line string) []Profile {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		return append(profiles, Profile{Key: line[1 : len(line)-1]})
	}
	name, value, found := setting(line)
	if !found || len(profiles) == 0 {
		return profiles
	}
	profile := &profiles[len(profiles)-1]
	if name == "screen" {
		profile.Screens = append(profile.Screens, Screen{Output: value, On: true, Rotation: "0"})
	} else if len(profile.Screens) > 0 {
		profile.Screens[len(profile.Screens)-1].set(name, value)
	}
	return profiles
}

func Parse(text string) []Profile {
	var profiles []Profile
	for line := range strings.Lines(strings.TrimPrefix(text, "\ufeff")) {
		profiles = readLine(profiles, line)
	}
	return profiles
}

func Find(profiles []Profile, key string) int {
	return slices.IndexFunc(profiles, func(profile Profile) bool { return profile.Key == key })
}

func Put(profiles []Profile, profile Profile) []Profile {
	if at := Find(profiles, profile.Key); at >= 0 {
		return slices.Replace(slices.Clone(profiles), at, at+1, profile)
	}
	return append(slices.Clone(profiles), profile)
}

func Forget(profiles []Profile, key string) ([]Profile, bool) {
	at := Find(profiles, key)
	if at < 0 {
		return profiles, false
	}
	return slices.Delete(slices.Clone(profiles), at, at+1), true
}

func (profile Profile) Screen(output string) int {
	return slices.IndexFunc(profile.Screens, func(screen Screen) bool { return screen.Output == output })
}

func (profile Profile) same(kept Screen) int {
	found := -1
	for at, screen := range profile.Screens {
		if screen.Description != kept.Description {
			continue
		}
		if screen.Output == kept.Output {
			return at
		}
		if found < 0 {
			found = at
		}
	}
	return found
}

func WithSaved(live Profile, saved []Profile) Profile {
	at := Find(saved, live.Key)
	if at < 0 {
		return live
	}
	for _, kept := range saved[at].Screens {
		if here := live.same(kept); here >= 0 && kept.Main {
			live.Screens[here].Main = true
		}
	}
	return live
}
