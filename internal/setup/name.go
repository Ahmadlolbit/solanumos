package setup

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/detect"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/runlog"
)

const (
	nameOption    = "SYSTEM_NAME"
	nameOurs      = "solanumos"
	brandName     = "SolanumOS"
	brandRelease  = "SolanumOS Chert"
	releaseFile   = "/etc/os-release"
	debianRelease = "/usr/lib/os-release"
	releaseHook   = "/etc/apt/apt.conf.d/80solanum-os-release"
	grubNameFile  = "/etc/default/grub.d/solanum-name.cfg"
	fastfetchFile = "/etc/xdg/fastfetch/config.jsonc"
	issueFile     = "/etc/issue"
	issueNetFile  = "/etc/issue.net"
	debianLink    = "../usr/lib/os-release"
)

var grubName = "case ${0##*/} in\ngrub-mkconfig) GRUB_DISTRIBUTOR=" + brandName + " ;;\n*) GRUB_DISTRIBUTOR=Debian ;;\nesac\n"

func namedPath(path string) string {
	return filepath.Join(env.Or("SOLANUM_NAME_ROOT", "/"), path)
}

func ourRelease(debian string) string {
	var lines []string
	for _, line := range strings.Split(strings.TrimSuffix(debian, "\n"), "\n") {
		switch {
		case strings.HasPrefix(line, "PRETTY_NAME="):
			line = `PRETTY_NAME="` + brandRelease + `"`
		case strings.HasPrefix(line, "NAME="):
			line = `NAME="` + brandName + `"`
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n") + "\n"
}

func diverted(path string) bool {
	output, _ := run.CaptureQuiet(run.Quick, "dpkg-divert", "--listpackage", path)
	return strings.TrimSpace(output) == "LOCAL"
}

func releaseRename(log *runlog.Log) error {
	release := namedPath(releaseFile)
	if !diverted(release) {
		if err := run.Logged(log, "dpkg-divert", "--local", "--no-rename", "--divert", release+".debian", "--add", release); err != nil {
			return err
		}
	}
	if !fsutil.Exists(release+".debian") && !fsutil.IsLink(release+".debian") {
		if err := atomic.Symlink(debianLink, release+".debian"); err != nil {
			return err
		}
	}
	debian, err := os.ReadFile(namedPath(debianRelease))
	if err != nil {
		return err
	}
	text := ourRelease(string(debian))
	if fsutil.Optional(release) == strings.TrimSuffix(text, "\n") {
		return nil
	}
	return atomic.Install(release, []byte(text), plainMode)
}

func debianBack(release string) error {
	if target, err := os.Readlink(release + ".debian"); err == nil {
		return atomic.Symlink(target, release)
	}
	if data, err := os.ReadFile(release + ".debian"); err == nil {
		return atomic.Install(release, data, plainMode)
	}
	return atomic.Symlink(debianLink, release)
}

func releaseBack(log *runlog.Log) error {
	release := namedPath(releaseFile)
	if !diverted(release) {
		return nil
	}
	if err := debianBack(release); err != nil {
		return err
	}
	if err := run.Logged(log, "dpkg-divert", "--local", "--no-rename", "--remove", release); err != nil {
		return err
	}
	return removeFiles(release + ".debian")
}

func (context *Context) grubNamed(log *runlog.Log, menu *followUp) error {
	if detect.Chroot() || apt.Commands("update-grub") != nil {
		return nil
	}
	return menu.do(func() error { return run.Logged(log, "update-grub") })
}

func (context *Context) nameInstall(log *runlog.Log) error {
	if err := releaseRename(log); err != nil {
		return err
	}
	for _, pair := range [][2]string{{issueFile, brandRelease + " \\n \\l\n\n"}, {issueNetFile, brandRelease + "\n"}} {
		if _, err := context.Deploy.Content([]byte(pair[1]), namedPath(pair[0])); err != nil {
			return err
		}
	}
	for _, pair := range [][2]string{{"system/apt/80solanum-os-release", releaseHook}, {"system/fastfetch.jsonc", fastfetchFile}} {
		if err := context.Deploy.File(pair[0], namedPath(pair[1])); err != nil {
			return err
		}
	}
	menu := &followUp{name: "grub-name"}
	changed, err := context.Deploy.Content([]byte(grubName), namedPath(grubNameFile))
	menu.need(changed)
	if err != nil {
		return err
	}
	return context.grubNamed(log, menu)
}

func (context *Context) nameRemove(log *runlog.Log) error {
	if err := releaseBack(log); err != nil {
		return err
	}
	for _, path := range []string{issueFile, issueNetFile, releaseHook, fastfetchFile, grubNameFile} {
		if _, err := context.Deploy.Undeploy(namedPath(path)); err != nil {
			return err
		}
	}
	menu := &followUp{name: "grub-name"}
	menu.need(true)
	return context.grubNamed(log, menu)
}

func (context *Context) nameSteps() {
	if context.Config.Get(nameOption) == nameOurs {
		context.Run.Step("Call the system SolanumOS", context.nameInstall)
	} else if fsutil.Exists(namedPath(grubNameFile)) || diverted(namedPath(releaseFile)) {
		context.Run.Step("Call the system Debian again", context.nameRemove)
	}
}
