package main

import (
	"os"
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/engine-switch"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/snapper-rollback"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-background"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-care"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-follow-theme"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-installer"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-pick"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-pkg"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-record"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-screensaver"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-sensors"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-session"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-settings-helper"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-start"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-task-helper"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-terminal"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-tip"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-weather"
	"github.com/Ahmadlolbit/solanumos/cmd/solanum-core/tools/solanum-webapp"
	"github.com/Ahmadlolbit/solanumos/internal/oldname"
)

var tools = map[string]func(){
	"engine-switch":           engineswitch.Main,
	"solanum":                 wizard.Main,
	"solanum-background":      background.Main,
	"solanum-care":            care.Main,
	"solanum-follow-theme":    followtheme.Main,
	"solanum-installer":       installer.Main,
	"solanum-pick":            pick.Main,
	"solanum-pkg":             pkg.Main,
	"solanum-record":          record.Main,
	"solanum-screensaver":     screensaver.Main,
	"solanum-sensors":         sensors.Main,
	"solanum-session":         login.Main,
	"solanum-settings-helper": helper.Main,
	"solanum-start":           start.Main,
	"solanum-task-helper":     task.Main,
	"solanum-terminal":        terminal.Main,
	"solanum-tip":             tip.Main,
	"solanum-weather":         weather.Main,
	"solanum-webapp":          webapp.Main,
	"snapper-rollback":        snapperrollback.Main,
}

func runTool() {
	name := oldname.Swap(filepath.Base(os.Args[0]))
	if name == "solanum" && root() == "" {
		os.Exit(enter())
	}
	if name == "solanum" && os.Getenv("SOLANUM_ENGINE") != "" {
		return
	}
	if tool, ok := tools[name]; ok {
		tool()
		os.Exit(0)
	}
}
