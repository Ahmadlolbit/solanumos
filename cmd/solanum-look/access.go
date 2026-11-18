package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	zoomStep    = 1.25
	zoomMost    = 8.0
	zoomToggled = 2.0
)

var filters = []string{"none", "greyscale", "deuteranopia", "protanopia", "tritanopia", "invert"}

func zoomFile() string {
	return filepath.Join(env.Or("XDG_RUNTIME_DIR", stateHome("solanum", "look")), "solanum-zoom")
}

func zoomNow() float64 {
	factor, err := strconv.ParseFloat(strings.TrimSpace(stored(zoomFile())), 64)
	if err != nil || factor < 1 || factor > zoomMost {
		return 1
	}
	return factor
}

func hyprConfig(lua string) {
	run.Quiet(run.Quick, "hyprctl", "eval", "hl.config({ "+lua+" })")
}

func hyprReload() int {
	run.Quiet(run.Quick, "hyprctl", "reload")
	return 0
}

func zoomTo(factor float64) int {
	factor = math.Round(min(max(factor, 1), zoomMost)*100) / 100
	text := strconv.FormatFloat(factor, 'f', -1, 64)
	if factor == 1 {
		remove(zoomFile())
	} else {
		save(zoomFile(), text)
	}
	hyprConfig("cursor = { zoom_factor = " + text + " }")
	return 0
}

func zoom(args []string) int {
	switch strings.ToLower(arg(args, 0)) {
	case "":
		fmt.Println(strconv.FormatFloat(zoomNow(), 'f', -1, 64))
		return 0
	case "in":
		return zoomTo(zoomNow() * zoomStep)
	case "out":
		return zoomTo(zoomNow() / zoomStep)
	case "reset":
		return zoomTo(1)
	case "toggle":
		if zoomNow() > 1 {
			return zoomTo(1)
		}
		return zoomTo(zoomToggled)
	}
	fmt.Fprintln(os.Stderr, "zoom takes in, out, reset or toggle")
	return 2
}

func zoomFollow(args []string) int {
	switch value := strings.ToLower(arg(args, 0)); value {
	case "":
		return show("zoom_follow")
	case "edges", "centre":
		return putThen("zoom_follow", value, func() int {
			hyprConfig("cursor = { zoom_rigid = " + strconv.FormatBool(value == "centre") + " }")
			return 0
		})()
	}
	fmt.Fprintln(os.Stderr, "zoom-follow takes edges or centre")
	return 2
}

func filter(args []string) int {
	value := strings.ToLower(arg(args, 0))
	switch {
	case value == "":
		return show("filter")
	case slices.Contains(filters, value):
		return putThen("filter", value, hyprReload)()
	}
	fmt.Fprintln(os.Stderr, "filter takes none, greyscale, deuteranopia, protanopia, tritanopia or invert")
	return 2
}

func animationsAgain() int {
	reloadAll()
	return run.Within(run.Slow, run.Sibling("solanum-theme-set"), "--animations")
}

func animations(args []string) int {
	return onOff{"animations", putThen("animations", "yes", animationsAgain), putThen("animations", "no", animationsAgain)}.run(args)
}
