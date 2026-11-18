package screens

import (
	"fmt"
	"strconv"
	"strings"
)

type View struct {
	Saved   bool         `json:"saved"`
	Trying  bool         `json:"trying"`
	Screens []ScreenView `json:"screens"`
}

type ScreenView struct {
	Output      string   `json:"output"`
	Description string   `json:"description"`
	Mode        string   `json:"mode"`
	Modes       []string `json:"modes"`
	X           int      `json:"x"`
	Y           int      `json:"y"`
	Width       int      `json:"width"`
	Height      int      `json:"height"`
	Scale       float64  `json:"scale"`
	Rotation    int      `json:"rotation"`
	On          bool     `json:"on"`
	Main        bool     `json:"main"`
	Mirror      string   `json:"mirror"`
}

func size(mode string) (int, int) {
	parts := modePattern.FindStringSubmatch(mode)
	if parts == nil {
		return 0, 0
	}
	wide, high, _ := strings.Cut(parts[1], "x")
	width, _ := strconv.Atoi(wide)
	height, _ := strconv.Atoi(high)
	return width, height
}

func shownScreen(screen Screen, monitor Monitor) ScreenView {
	width, height := size(screen.Mode)
	scale, _ := strconv.ParseFloat(screen.Scale, 64)
	rotation, _ := strconv.Atoi(screen.Rotation)
	return ScreenView{
		Output: screen.Output, Description: screen.Description, Mode: screen.Mode, Modes: Modes(monitor),
		X: monitor.X, Y: monitor.Y, Width: width, Height: height, Scale: scale, Rotation: rotation,
		On: screen.On, Main: screen.Main, Mirror: screen.Mirror,
	}
}

func Shown(live Profile, monitors []Monitor, saved, trying bool) View {
	view := View{Saved: saved, Trying: trying, Screens: []ScreenView{}}
	for at, screen := range live.Screens {
		view.Screens = append(view.Screens, shownScreen(screen, monitors[at]))
	}
	return view
}

func Summary(screen Screen) string {
	parts := []string{fmt.Sprintf("%s: %s at %s, scale %s", screen.Output, screen.Mode, screen.Position, screen.Scale)}
	if screen.Rotation != "" && screen.Rotation != "0" {
		parts = append(parts, "turned "+screen.Rotation)
	}
	if screen.Main {
		parts = append(parts, "main")
	}
	if screen.Mirror != "" {
		parts = append(parts, "mirrors "+screen.Mirror)
	}
	if !screen.On {
		parts = append(parts, "off")
	}
	return strings.Join(parts, ", ")
}
