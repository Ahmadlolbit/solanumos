package main

import (
	"cmp"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/care"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/lock"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const updateUsage = "usage: solanum update [system | apps | solanum | --list --json]"

type pending struct {
	Name string `json:"name"`
	From string `json:"from"`
	To   string `json:"to"`
}

type waiting struct {
	Debian   []pending `json:"debian"`
	Apps     []pending `json:"apps"`
	Tools    []string  `json:"tools"`
	Solanum  []pending `json:"solanum"`
	Firmware []pending `json:"firmware"`
	Last     string    `json:"last"`
	Restart  bool      `json:"restart"`
	Busy     bool      `json:"busy"`
}

func debianWaiting() []pending {
	list := []pending{}
	output, err := run.CaptureInEnglish(run.Minute, "apt", "list", "--upgradable")
	if err != nil {
		return list
	}
	for _, line := range run.Lines(output) {
		name, rest, found := strings.Cut(line, "/")
		fields := strings.Fields(rest)
		_, from, upgradable := strings.Cut(line, "[upgradable from: ")
		if found && upgradable && len(fields) > 1 {
			list = append(list, pending{name, strings.TrimSuffix(from, "]"), fields[1]})
		}
	}
	return list
}

func appsWaiting() ([]pending, []string) {
	apps, tools := []pending{}, []string{}
	for _, tool := range appTools {
		if installed(tool.name) {
			tools = append(tools, tool.what)
		}
	}
	if !installed("flatpak") {
		return apps, tools
	}
	output, _ := run.CaptureQuiet(run.Minute, "flatpak", "remote-ls", "--updates", "--cached", "--columns=name,application,version")
	for _, line := range run.Lines(output) {
		fields := strings.Split(line+"\t\t", "\t")
		if name := cmp.Or(strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])); name != "" {
			apps = append(apps, pending{Name: name, To: strings.TrimSpace(fields[2])})
		}
	}
	return apps, tools
}

func solanumWaiting() []pending {
	if root() == "" || gitQuiet(run.Quick, "rev-parse", "--git-dir") != 0 {
		return []pending{}
	}
	signers, done, err := trustedSigners()
	defer done()
	if err != nil {
		return []pending{}
	}
	if tag, _ := newestSigned(signers); tag != "" && !related(tag, "HEAD") {
		return []pending{{"SolanumOS", env.Version, tag}}
	}
	return []pending{}
}

func firmwareWaiting() []pending {
	list := []pending{}
	if !installed("fwupdmgr") {
		return list
	}
	output, _ := run.CaptureQuiet(run.Minute, "fwupdmgr", "get-updates", "--json")
	for _, update := range care.Updates([]byte(output)) {
		list = append(list, pending{Name: update})
	}
	return list
}

func lastUpgrade() string {
	started, last := "", ""
	for _, line := range strings.Split(fsutil.Optional(env.Or("SOLANUM_APT_HISTORY", "/var/log/apt/history.log")), "\n") {
		if date, found := strings.CutPrefix(line, "Start-Date: "); found {
			started = strings.Join(strings.Fields(date), "T")
		} else if strings.HasPrefix(line, "Commandline: ") && strings.Contains(line, "upgrade") {
			last = started
		}
	}
	return last
}

func restartPending() bool {
	if fsutil.IsFile(env.Or("SOLANUM_REBOOT_REQUIRED", "/run/reboot-required")) {
		return true
	}
	boot := env.Or("SOLANUM_BOOT_DIR", "/boot")
	release, err := run.CaptureQuiet(run.Quick, "uname", "-r")
	running := strings.TrimSpace(release)
	kernels, unread := kernelsIn(boot)
	if err != nil || unread != nil || !slices.Contains(kernels, running) {
		return false
	}
	newest, unsorted := newestKernel(boot, kernels)
	return unsorted == nil && newest != running
}

func listUpdates() int {
	if env.Arg(3) != "--json" || env.Arg(4) != "" {
		return fail(updateUsage)
	}
	apps, tools := appsWaiting()
	return printed(waiting{
		Debian:   debianWaiting(),
		Apps:     apps,
		Tools:    tools,
		Solanum:  solanumWaiting(),
		Firmware: firmwareWaiting(),
		Last:     lastUpgrade(),
		Restart:  restartPending(),
		Busy:     lock.Busy(),
	})
}

func updateCommand() int {
	if env.Arg(2) == "--list" {
		return listUpdates()
	}
	return update()
}
