package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/look"
	"github.com/Ahmadlolbit/solanumos/internal/notify"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

const keyboardSize = 16

var (
	keyboards     = []string{"wvkbd-deskintl", "wvkbd-mobintl", "wvkbd"}
	keyboardParts = [][2]string{{"--bg", "background"}, {"--fg", "lighter_background"}, {"--fg-sp", "selection"}, {"--press", "accent"}, {"--press-sp", "accent"}, {"--text", "foreground"}, {"--text-sp", "foreground"}}
)

func keyboardProgram() string {
	for _, name := range keyboards {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}
	return ""
}

func keyboardUp(name string) bool {
	return name != "" && run.Quiet(run.Quick, "pgrep", "-x", name) == 0
}

func keyboardLook() []string {
	colors := themeColorsNow()
	var args []string
	for _, part := range keyboardParts {
		if theme.IsHex(colors[part[1]]) {
			args = append(args, part[0], strings.TrimPrefix(colors[part[1]], "#"))
		}
	}
	size := math.Round(keyboardSize * theme.TextScale(settings.Get("text")))
	return append(args, "--fn", settings.Get("font")+" "+strconv.FormatFloat(size, 'f', -1, 64))
}

func keyboardOn(name string) int {
	if name == "" {
		fmt.Fprintln(os.Stderr, "The on-screen keyboard isn't installed: run the setup again, or sudo apt install wvkbd.")
		notify.Post([]string{"--app-name=On-screen keyboard"}, "Not installed: run the setup again", "")
		return 1
	}
	if !keyboardUp(name) {
		if code := run.Quiet(run.Quick, "setsid", append([]string{"-f", name}, keyboardLook()...)...); code != 0 {
			return code
		}
	}
	look.CallShell("keyboard", "shown")
	return 0
}

func keyboardOff(name string) int {
	if keyboardUp(name) {
		run.Quiet(run.Quick, "pkill", "-x", name)
	}
	look.CallShell("keyboard", "hidden")
	return 0
}

func keyboard(args []string) int {
	name := keyboardProgram()
	switch strings.ToLower(arg(args, 0)) {
	case "":
		if keyboardUp(name) {
			fmt.Println("shown")
		} else {
			fmt.Println("hidden")
		}
		return 0
	case "show":
		return keyboardOn(name)
	case "hide":
		return keyboardOff(name)
	case "toggle":
		if keyboardUp(name) {
			return keyboardOff(name)
		}
		return keyboardOn(name)
	}
	fmt.Fprintln(os.Stderr, "keyboard takes show, hide or toggle")
	return 2
}
