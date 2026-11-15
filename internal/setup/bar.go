package setup

import (
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/runlog"
)

const (
	barSettings    = "/etc/solanum/shell.json"
	barSensors     = "solanum-sensors"
	settingsHelper = "solanum-settings-helper"
	settingsPolicy = "/usr/share/polkit-1/actions/org.solanumos.settings.policy"
	taskHelper     = "solanum-task-helper"
	taskPolicy     = "/usr/share/polkit-1/actions/org.solanumos.task.policy"
)

func barDir() string {
	return filepath.Join(env.ShareDir(), "shell")
}

func settingsDir() string {
	return filepath.Join(env.ShareDir(), "settings")
}

func barReady(*runlog.Log) error {
	if err := fileThere(filepath.Join(barDir(), "shell.qml")); err != nil {
		return err
	}
	if err := fileThere(barSettings); err != nil {
		return err
	}
	sensors := filepath.Join(env.BinDir(), barSensors)
	if err := expect(fsutil.Runnable(sensors), sensors+" won't run"); err != nil {
		return err
	}
	return apt.Commands("qs")
}

func settingsReady(*runlog.Log) error {
	if err := fileThere(filepath.Join(settingsDir(), "common", "Theme.qml")); err != nil {
		return err
	}
	for _, policy := range []string{env.Or("SOLANUM_SETTINGS_POLICY", settingsPolicy), env.Or("SOLANUM_TASK_POLICY", taskPolicy)} {
		if err := fileThere(policy); err != nil {
			return err
		}
	}
	for _, name := range []string{settingsHelper, taskHelper} {
		helper := filepath.Join(env.BinDir(), name)
		if err := expect(fsutil.Runnable(helper), helper+" won't run"); err != nil {
			return err
		}
	}
	return nil
}

func (context *Context) settingsInstall(*runlog.Log) error {
	if err := context.Deploy.Tree("settings", settingsDir()); err != nil {
		return err
	}
	if err := context.Deploy.Link("../shell/common", filepath.Join(settingsDir(), "common")); err != nil {
		return err
	}
	for _, name := range []string{settingsHelper, taskHelper} {
		if err := context.Deploy.Tool(name); err != nil {
			return err
		}
	}
	if err := context.Deploy.File("system/polkit/org.solanumos.settings.policy", env.Or("SOLANUM_SETTINGS_POLICY", settingsPolicy), "HELPER="+filepath.Join(env.BinDir(), settingsHelper)); err != nil {
		return err
	}
	return context.Deploy.File("system/polkit/org.solanumos.task.policy", env.Or("SOLANUM_TASK_POLICY", taskPolicy), "HELPER="+filepath.Join(env.BinDir(), taskHelper))
}

func (context *Context) barInstall(*runlog.Log) error {
	if err := context.Deploy.Tree("shell", barDir()); err != nil {
		return err
	}
	return context.Deploy.Tool(barSensors)
}

func (context *Context) barSetup(*runlog.Log) error {
	settings := context.Config
	return context.Deploy.File("system/shell.json", barSettings, "POSITION="+settings.Get("BAR_POSITION"), "CLOCK="+settings.Get("CLOCK_FORMAT"),
		"PRESET="+settings.Get("BAR_PRESET"))
}

func barSteps(context *Context) {
	context.Run.Step("Install the bar", context.barInstall)
	context.Run.Step("Apply the bar settings", context.barSetup)
	context.Run.Step("Install Care", context.careWindowInstall)
	context.Run.Step("Install Settings", context.settingsInstall)
	context.Run.Step("Install Software", context.softwareInstall)
}

func barChecks(context *Context) {
	context.Run.Check("The bar is installed", barReady)
	context.Run.Check("Care is installed", careWindowReady)
	context.Run.Check("Settings is installed", settingsReady)
	context.Run.Check("Software is installed", softwareReady)
}
