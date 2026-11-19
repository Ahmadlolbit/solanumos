package main

import (
	"errors"
	"os"
	"os/signal"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/cleanup"
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
	dpkgLeft    = apt.Interrupted
	dpkgState   = apt.DpkgStamp
	takeLock    = lock.Take
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

func stoppedEarly() int {
	ui.Done()
	ui.Skip("Stopped, nothing changed")
	return stoppedCode
}

func lockedWhenFree(stop <-chan os.Signal) (*lock.Held, int) {
	for waited := time.Duration(0); ; waited += othersPause {
		if !apt.Busy() {
			if held, err := takeLock(); err == nil {
				return held, 0
			}
		}
		if waited == 0 {
			ui.Busy("Waiting for another install or update to finish, perhaps the daily update")
		}
		if waited >= othersLimit {
			return nil, fail("another install or update went on for 10 minutes, so nothing changed: try again once it ends")
		}
		select {
		case <-stop:
			return nil, stoppedEarly()
		case <-time.After(othersPause):
		}
	}
}

func simulated(action string, names []string) ([]string, error) {
	output, said, err := run.CaptureFed(run.Minute, "", "env", append([]string{"LC_ALL=C.UTF-8", "apt-get", "-s", action}, names...)...)
	var others []string
	for _, line := range run.Lines(output) {
		fields := strings.Fields(line)
		if len(fields) > 1 && fields[0] == "Remv" && !slices.Contains(names, fields[1]) {
			others = append(others, fields[1])
		}
	}
	switch {
	case run.TimedOut(err):
		return others, errors.New("apt-get took longer than a minute")
	case err != nil:
		lines := run.Lines(strings.TrimSpace(said))
		if len(lines) == 0 {
			return others, err
		}
		return others, errors.New(lines[len(lines)-1])
	}
	return others, nil
}

func alsoRemoved(action string, names []string) []string {
	others, _ := simulated(action, names)
	return others
}

func agreed(action string, names []string) bool {
	others := alsoRemoved(action, names)
	return len(others) == 0 || ui.Confirm("This also removes "+strings.Join(others, ", ")+". Go ahead?")
}

func agreedUnlessStopped(stop <-chan os.Signal, action string, names []string) (bool, bool) {
	if len(stop) > 0 {
		return false, true
	}
	answer := make(chan bool, 1)
	go func() { answer <- agreed(action, names) }()
	select {
	case yes := <-answer:
		return yes, len(stop) > 0
	case <-stop:
		return false, true
	}
}

func unasked(args ...string) []string {
	return append([]string{"DEBIAN_FRONTEND=noninteractive", "NEEDRESTART_MODE=a", "APT_LISTCHANGES_FRONTEND=none"}, args...)
}

func finishDpkg() int {
	if !dpkgLeft() {
		return 0
	}
	ui.Busy("dpkg was interrupted, finishing what it left")
	code, _ := run.Shielded(nil, "env", unasked("dpkg", "--force-confdef", "--force-confold", "--configure", "-a")...)
	ui.Done()
	if code != 0 {
		ui.Fail("Finishing what dpkg left failed, so nothing else changed; the log says why")
	}
	return code
}

func sayStopped(before string) int {
	if dpkgState() == before {
		ui.Skip("Stopped, nothing changed")
	} else {
		ui.Skip("Stopped after some packages changed; the log says which")
	}
	return stoppedCode
}

func applied(stop <-chan os.Signal, before, action string, names []string) int {
	listed := strings.Join(names, ", ")
	ui.Busy(pkgVerbs[action][0] + listed)
	code, halted := run.Shielded(stop, "env", unasked(slices.Concat([]string{"apt-get"}, aptOptions, []string{action}, names)...)...)
	ui.Done()
	if halted {
		return sayStopped(before)
	}
	if code != 0 {
		ui.Fail(pkgVerbs[action][0] + listed + " failed; the log says why")
		return code
	}
	ui.OK(listed + pkgVerbs[action][1])
	return 0
}

func changePackages(action string, names []string) int {
	stop := make(chan os.Signal, 1)
	cleanup.NotifyInterrupt(stop)
	defer signal.Stop(stop)
	held, code := lockedWhenFree(stop)
	if held == nil {
		return code
	}
	defer held.Close()
	if len(stop) > 0 {
		return stoppedEarly()
	}
	before := dpkgState()
	if code := finishDpkg(); code != 0 {
		return code
	}
	yes, halted := agreedUnlessStopped(stop, action, names)
	if halted {
		return sayStopped(before)
	}
	if !yes {
		ui.Skip("Nothing changed")
		return 0
	}
	return applied(stop, before, action, names)
}

func pkgCommand() int {
	action, names, ok := packagesWanted()
	if !ok {
		return fail(pkgUsage)
	}
	return asRoot("pkg "+action+" "+strings.Join(names, " "), func() int { return changePackages(action, names) })
}
