package main

import (
	"fmt"
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

func takeKept(name string) (string, bool) {
	path := filepath.Join(state(), name)
	value, err := fsutil.Text(path)
	if err != nil {
		return "", false
	}
	os.Remove(path)
	return value, true
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
	before, kept := takeKept(gtkMotionKept)
	switch {
	case !kept || before == now:
	case before == "":
		desktop("reset", "enable-animations")
	default:
		desktop("set", "enable-animations", before)
	}
}

func kdeMotion(text string) string {
	if animationsOff() {
		keepOnce(kdeMotionKept, theme.KDESpeedOf(text))
		return theme.KDEAnimations(text, true)
	}
	if before, kept := takeKept(kdeMotionKept); kept {
		return theme.KDEGivenSpeed(text, before)
	}
	return text
}

func kdeAnimations() {
	globals := filepath.Join(env.ConfigHome(), "kdeglobals")
	current := fsutil.Optional(globals)
	merged := kdeMotion(current)
	if theme.KDESpeedOf(merged) == theme.KDESpeedOf(current) {
		return
	}
	if err := atomic.Write(globals, []byte(merged), 0o666); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	kdeChanged, kdeSpeedChanged = true, true
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
