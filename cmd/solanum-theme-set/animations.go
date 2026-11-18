package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

const (
	gtkMotionKept = "animations.gtk"
	kdeMotionKept = "animations.kde"
	kdeStill      = "0"
)

var kdeSpeedChanged = false

func animationsOff() bool {
	return settings.Get("animations") == "no" && settings.Get("apps") != appsOff
}

func keepOnce(name, value string) {
	path := filepath.Join(state(), name)
	if fsutil.Exists(path) {
		return
	}
	if err := atomic.Write(path, []byte(value+"\n"), 0o666); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func keptValue(name string) (string, bool) {
	value, err := fsutil.Text(filepath.Join(state(), name))
	return value, err == nil
}

func forget(name string) {
	if err := os.Remove(filepath.Join(state(), name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(os.Stderr, err)
	}
}

func gtkGivenBack(before, now string) bool {
	switch {
	case before == now:
		return true
	case before == "":
		return desktop("reset", "enable-animations")
	}
	return desktop("set", "enable-animations", before)
}

func gtkAnimations(current map[string]string) {
	now := current["enable-animations"]
	if animationsOff() {
		keepOnce(gtkMotionKept, now)
		if now != "false" {
			desktop("set", "enable-animations", "false")
		}
		return
	}
	if before, kept := keptValue(gtkMotionKept); kept && gtkGivenBack(before, now) {
		forget(gtkMotionKept)
	}
}

func kdeMotion(text string) (string, bool) {
	if animationsOff() {
		keepOnce(kdeMotionKept, theme.KDESpeedOf(text))
		return theme.KDEAnimations(text, true), false
	}
	before, kept := keptValue(kdeMotionKept)
	if !kept || theme.KDESpeedOf(text) != kdeStill {
		return text, kept
	}
	return theme.KDEGivenSpeed(text, before), true
}

func kdeAnimations() {
	globals := filepath.Join(env.ConfigHome(), "kdeglobals")
	current := fsutil.Optional(globals)
	merged, givenBack := kdeMotion(current)
	if theme.KDESpeedOf(merged) != theme.KDESpeedOf(current) {
		if err := atomic.Write(globals, []byte(merged), 0o666); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		kdeChanged, kdeSpeedChanged = true, true
	}
	if givenBack {
		forget(kdeMotionKept)
	}
}

func kdeSpeedSignal() string {
	var bytes []string
	for _, each := range []byte(strings.TrimSuffix(theme.KDESpeed, "=")) {
		bytes = append(bytes, fmt.Sprintf("0x%02x", each))
	}
	return "{'KDE': [[byte " + strings.Join(bytes, ", ") + "]]}"
}

func tellKDESpeed() {
	if !kdeSpeedChanged {
		return
	}
	run.Quiet(run.Quick, "gdbus", "emit", "--session", "--object-path", "/kdeglobals", "--signal", "org.kde.kconfig.notify.ConfigChanged", kdeSpeedSignal())
	kdeSpeedChanged = false
}

func animationsOnly() int {
	if _, err := exec.LookPath("gsettings"); err == nil {
		gtkAnimations(desktopValues())
	}
	kdeAnimations()
	tellKDEApps()
	return 0
}
