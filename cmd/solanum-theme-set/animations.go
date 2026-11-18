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

var kdeSpeedChanged = false

func animationsOff() bool {
	return settings.Get("animations") == "no"
}

func gtkAnimations(current map[string]string) {
	switch now := current["enable-animations"]; {
	case animationsOff() && now != "false":
		desktop("set", "enable-animations", "false")
	case !animationsOff() && now == "false":
		desktop("reset", "enable-animations")
	}
}

func kdeAnimations() {
	globals := filepath.Join(env.ConfigHome(), "kdeglobals")
	current := fsutil.Optional(globals)
	merged := theme.KDEAnimations(current, animationsOff())
	if theme.KDEStill(merged) == theme.KDEStill(current) {
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
