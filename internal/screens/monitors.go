package screens

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type Monitor struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Model          string   `json:"model"`
	Serial         string   `json:"serial"`
	Width          int      `json:"width"`
	Height         int      `json:"height"`
	RefreshRate    float64  `json:"refreshRate"`
	X              int      `json:"x"`
	Y              int      `json:"y"`
	Scale          float64  `json:"scale"`
	Transform      int      `json:"transform"`
	Disabled       bool     `json:"disabled"`
	MirrorOf       string   `json:"mirrorOf"`
	AvailableModes []string `json:"availableModes"`
}

func ParseMonitors(data string) ([]Monitor, error) {
	var found []Monitor
	if err := json.Unmarshal([]byte(data), &found); err != nil {
		return nil, fmt.Errorf("hyprctl monitors: %w", err)
	}
	for at := range found {
		monitor := &found[at]
		monitor.Description = strings.TrimSpace(strings.TrimSuffix(monitor.Description, " ("+monitor.Name+")"))
		monitor.Model = strings.TrimSpace(monitor.Model)
		monitor.Serial = strings.TrimSpace(monitor.Serial)
	}
	return found, nil
}

func Key(monitors []Monitor) string {
	descriptions := make([]string, 0, len(monitors))
	for _, monitor := range monitors {
		descriptions = append(descriptions, monitor.Description)
	}
	slices.Sort(descriptions)
	return strings.Join(descriptions, ", ")
}

func Modes(monitor Monitor) []string {
	modes := make([]string, 0, len(monitor.AvailableModes))
	for _, mode := range monitor.AvailableModes {
		if mode = strings.TrimSuffix(strings.TrimSpace(mode), "Hz"); !slices.Contains(modes, mode) {
			modes = append(modes, mode)
		}
	}
	return modes
}

func CurrentMode(monitor Monitor) string {
	if monitor.Width <= 0 || monitor.Height <= 0 {
		if modes := Modes(monitor); len(modes) > 0 {
			return modes[0]
		}
		return Preferred
	}
	return fmt.Sprintf("%dx%d@%.2f", monitor.Width, monitor.Height, monitor.RefreshRate)
}

func mirrorName(monitors []Monitor, of string) string {
	for _, other := range monitors {
		if other.Name == of || strconv.Itoa(other.ID) == of {
			return other.Name
		}
	}
	return ""
}

func scaleText(scale float64) string {
	if scale <= 0 {
		return "1"
	}
	return strconv.FormatFloat(scale, 'f', -1, 64)
}

func Live(monitors []Monitor) Profile {
	profile := Profile{Key: Key(monitors)}
	for _, monitor := range monitors {
		profile.Screens = append(profile.Screens, Screen{
			Output:      monitor.Name,
			Description: monitor.Description,
			Model:       monitor.Model,
			Serial:      monitor.Serial,
			Mode:        CurrentMode(monitor),
			Position:    fmt.Sprintf("%dx%d", monitor.X, monitor.Y),
			Scale:       scaleText(monitor.Scale),
			Rotation:    strconv.Itoa(monitor.Transform % 4 * 90),
			On:          !monitor.Disabled,
			Mirror:      mirrorName(monitors, monitor.MirrorOf),
		})
	}
	return profile
}
