package main

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/lock"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
)

const (
	pkgUsage    = "usage: sudo solanum pkg install|remove NAME..."
	stoppedCode = 130
)

var (
	packageName = regexp.MustCompile(`^[a-z0-9][a-z0-9+.-]{0,99}(:[a-z0-9-]{1,20})?$`)
	othersPause = 2 * time.Second
	othersLimit = 10 * time.Minute
	pkgVerbs    = map[string][2]string{"install": {"Installing ", " is installed"}, "remove": {"Removing ", " is removed"}}
	aptOptions  = []string{"-y", "-o", "DPkg::Lock::Timeout=600", "-o", "Dpkg::Use-Pty=0", "-o", "Dpkg::Options::=--force-confdef", "-o", "Dpkg::Options::=--force-confold"}
)

func aPackage(name string) bool {
	return run.Quiet(run.Quick, "apt-cache", "show", "--no-all-versions", name) == 0
}

func packagesWanted() (string, []string, bool) {
	action, names := env.Arg(2), os.Args[min(3, len(os.Args)):]
	bad := func(name string) bool {
		return !packageName.MatchString(name) || strings.ContainsAny(name[len(name)-1:], "+-") && !aPackage(name)
	}
	_, known := pkgVerbs[action]
	return action, names, known && len(names) > 0 && !slices.ContainsFunc(names, bad)
}

func lockedWhenFree() (*lock.Held, int) {
	for waited := time.Duration(0); ; waited += othersPause {
		if !apt.Busy() {
			if held, err := lock.Take(); err == nil {
				return held, 0
			}
		}
		if waited == 0 {
			ui.Busy("Waiting for another install or update to finish, perhaps the daily update")
		}
		if run.Stopped() {
			ui.Done()
			ui.Skip("Stopped, nothing changed")
			return nil, stoppedCode
		}
		if waited >= othersLimit {
			return nil, fail("another install or update went on for 10 minutes, so nothing changed: try again once it ends")
		}
		time.Sleep(othersPause)
	}
}

func alsoRemoved(action string, names []string) []string {
	output, _ := run.CaptureQuiet(run.Minute, "env", append([]string{"LC_ALL=C.UTF-8", "apt-get", "-s", action}, names...)...)
	var others []string
	for _, line := range run.Lines(output) {
		fields := strings.Fields(line)
		if len(fields) > 1 && fields[0] == "Remv" && !slices.Contains(names, fields[1]) {
			others = append(others, fields[1])
		}
	}
	return others
}

func agreed(action string, names []string) bool {
	others := alsoRemoved(action, names)
	return len(others) == 0 || ui.Confirm("This also removes "+strings.Join(others, ", ")+". Go ahead?")
}

func changePackages(action string, names []string) int {
	held, code := lockedWhenFree()
	if held == nil {
		return code
	}
	defer held.Close()
	listed := strings.Join(names, ", ")
	if !agreed(action, names) {
		ui.Skip("Nothing changed")
		return 0
	}
	ui.Busy(pkgVerbs[action][0] + listed)
	code = run.Status("env", append(append([]string{"DEBIAN_FRONTEND=noninteractive", "NEEDRESTART_MODE=a", "APT_LISTCHANGES_FRONTEND=none", "apt-get"}, aptOptions...), append([]string{action}, names...)...)...)
	ui.Done()
	if code != 0 {
		ui.Fail(pkgVerbs[action][0] + listed + " failed; the log says why")
		return code
	}
	ui.OK(listed + pkgVerbs[action][1])
	return 0
}

func pkgCommand() int {
	action, names, ok := packagesWanted()
	if !ok {
		return fail(pkgUsage)
	}
	return asRoot("pkg "+action+" "+strings.Join(names, " "), func() int { return changePackages(action, names) })
}
