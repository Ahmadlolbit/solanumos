package main

import (
	"cmp"
	"errors"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/care"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/lock"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	updateUsage = "usage: solanum update [system | apps | solanum | --list --json]"
	nothingToDo = 2
)

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
	Unread   []string  `json:"unread"`
}

func debianWaiting() ([]pending, bool) {
	list := []pending{}
	output, err := run.CaptureInEnglish(run.Minute, "apt", "list", "--upgradable")
	if err != nil {
		return list, false
	}
	for _, line := range run.Lines(output) {
		name, rest, found := strings.Cut(line, "/")
		fields := strings.Fields(rest)
		_, from, upgradable := strings.Cut(line, "[upgradable from: ")
		if found && upgradable && len(fields) > 1 {
			list = append(list, pending{name, strings.TrimSuffix(from, "]"), fields[1]})
		}
	}
	return list, true
}

func appsWaiting() ([]pending, []string, bool) {
	apps, tools := []pending{}, []string{}
	for _, tool := range appTools {
		if installed(tool.name) {
			tools = append(tools, tool.what)
		}
	}
	if !installed("flatpak") {
		return apps, tools, true
	}
	output, err := run.CaptureQuiet(run.Minute, "flatpak", "remote-ls", "--updates", "--cached", "--columns=name,application,version")
	if err != nil {
		output, err = run.CaptureQuiet(run.Minute, "flatpak", "remote-ls", "--updates", "--columns=name,application,version")
	}
	if err != nil {
		return apps, tools, false
	}
	for _, line := range run.Lines(output) {
		fields := strings.Split(line+"\t\t", "\t")
		if name := cmp.Or(strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])); name != "" {
			apps = append(apps, pending{Name: name, To: strings.TrimSpace(fields[2])})
		}
	}
	return apps, tools, true
}

func solanumWaiting() ([]pending, bool) {
	if root() == "" {
		return []pending{}, true
	}
	if gitQuiet(run.Quick, "rev-parse", "--git-dir") != 0 {
		return []pending{}, false
	}
	signers, done, err := trustedSigners()
	defer done()
	if err != nil {
		return []pending{}, errors.Is(err, errNoReleaseKey)
	}
	if tag, _ := newestSigned(signers); tag != "" && !related(tag, "HEAD") {
		return []pending{{"SolanumOS", env.Version, strings.TrimPrefix(tag, "v")}}, true
	}
	return []pending{}, true
}

func firmwareWaiting() ([]pending, bool) {
	list := []pending{}
	if !installed("fwupdmgr") {
		return list, true
	}
	output, err := run.CaptureQuiet(run.Minute, "fwupdmgr", "get-updates", "--json")
	if err != nil && run.Code(err) != nothingToDo {
		return list, false
	}
	for _, update := range care.Updates([]byte(output)) {
		list = append(list, pending{Name: update})
	}
	return list, true
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
	apps, tools, appsRead := appsWaiting()
	debian, debianRead := debianWaiting()
	solanum, solanumRead := solanumWaiting()
	firmware, firmwareRead := firmwareWaiting()
	return printed(waiting{
		Debian:   debian,
		Apps:     apps,
		Tools:    tools,
		Solanum:  solanum,
		Firmware: firmware,
		Last:     lastUpgrade(),
		Restart:  restartPending(),
		Busy:     lock.Busy(),
		Unread:   unreadParts(map[string]bool{"debian": debianRead, "apps": appsRead, "solanum": solanumRead, "firmware": firmwareRead}),
	})
}

func unreadParts(read map[string]bool) []string {
	unread := []string{}
	for _, part := range []string{"debian", "apps", "solanum", "firmware"} {
		if !read[part] {
			unread = append(unread, part)
		}
	}
	return unread
}

func updateCommand() int {
	if env.Arg(2) == "--list" {
		return listUpdates()
	}
	if env.Arg(2) == "--hold" && env.Arg(3) == "" {
		return holdUntilDone()
	}
	return update()
}
