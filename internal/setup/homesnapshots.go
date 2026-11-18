package setup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/Ahmadlolbit/solanumos/internal/btrfs"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/files"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/runlog"
)

const (
	homeSnapper   = "home"
	homeFolder    = "/home"
	subvolumeRoot = 256
	steamFolder   = ".var/app/com.valvesoftware.Steam"
	privateFolder = "700"
)

func homeSnapperConfig() string {
	return env.Or("SOLANUM_HOME_SNAPPER_CONFIG", "/etc/snapper/configs/home")
}

func homeApart() bool {
	if !btrfs.Mounted(homeFolder) {
		return false
	}
	kind, err := run.CaptureQuiet(run.Minute, "findmnt", "-no", "FSTYPE", homeFolder)
	return err == nil && strings.TrimSpace(kind) == "btrfs"
}

func isSubvolume(path string) bool {
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Ino == subvolumeRoot
}

func (context *Context) keptOutOfHome() []string {
	folders := []string{".cache"}
	if slices.Contains(strings.Fields(context.Config.Get("EXTRAS")), "gaming") {
		folders = append(folders, steamFolder)
	}
	return folders
}

func keptIn(log *runlog.Log, rel string) error {
	fmt.Fprintf(log, "~/%s already holds files, so it stays in the snapshots of /home: making it a subvolume of its own would move them, and that isn't done under a running desktop\n", rel)
	return nil
}

func (context *Context) keepOut(log *runlog.Log, rel string) error {
	person := context.Person
	path := person.Path(rel)
	switch {
	case isSubvolume(path):
		return nil
	case fsutil.IsLink(path) || !empty(path):
		return keptIn(log, rel)
	case fsutil.Exists(path):
		if person.Run(log, "rmdir", path) != nil {
			return keptIn(log, rel)
		}
	}
	if err := person.Run(log, "mkdir", "-p", filepath.Dir(path)); err != nil {
		return err
	}
	if err := person.Run(log, "btrfs", "subvolume", "create", path); err != nil {
		return err
	}
	return person.Run(log, "chmod", privateFolder, path)
}

func (context *Context) homeKeptOut(log *runlog.Log) error {
	for _, rel := range context.keptOutOfHome() {
		if err := context.keepOut(log, rel); err != nil {
			return err
		}
	}
	return nil
}

func (context *Context) homeSnapperSettings() [][2]string {
	return [][2]string{
		{"TIMELINE_CREATE", "yes"}, {"TIMELINE_CLEANUP", "yes"}, {"TIMELINE_LIMIT_HOURLY", context.Config.Get("HOME_SNAPSHOT_HOURLY")},
		{"TIMELINE_LIMIT_DAILY", context.Config.Get("HOME_SNAPSHOT_DAILY")}, {"TIMELINE_LIMIT_WEEKLY", "0"}, {"TIMELINE_LIMIT_MONTHLY", "0"},
		{"TIMELINE_LIMIT_YEARLY", "0"}, {"NUMBER_CLEANUP", "yes"}, {"NUMBER_LIMIT", "10"}, {"ALLOW_USERS", context.Person.Name},
		{"ALLOW_GROUPS", adminGroup}, {"SYNC_ACL", "yes"},
	}
}

func (context *Context) homeSnapperConfigure(log *runlog.Log) error {
	if !fsutil.IsFile(homeSnapperConfig()) {
		if err := run.Logged(log, "snapper", "--no-dbus", "-c", homeSnapper, "create-config", "--fstype", "btrfs", homeFolder); err != nil {
			return err
		}
	}
	if !fsutil.IsFile(homeSnapperConfig()) {
		return errors.New("snapper made no configuration for /home")
	}
	if err := snapperRegisterAs(homeSnapper); err != nil {
		return err
	}
	for _, setting := range context.homeSnapperSettings() {
		if err := files.ConfSet(homeSnapperConfig(), setting[0], `"`+setting[1]+`"`); err != nil {
			return err
		}
	}
	return nil
}

func homeSnapshotsStop(log *runlog.Log) error {
	if err := files.ConfSet(homeSnapperConfig(), "TIMELINE_CREATE", `"no"`); err != nil {
		return err
	}
	log.Line("info", "no new snapshots of /home are taken from now on; the ones taken so far stay until sudo snapper -c home delete removes them")
	return nil
}

func homeSnapperReady(*runlog.Log) error {
	configs, _ := run.CaptureQuiet(run.Minute, "snapper", "--no-dbus", "list-configs")
	return expect(lineHasPrefix(configs, homeSnapper+" "), "snapper has no config for /home")
}

func (context *Context) homeSnapshotSteps() {
	switch {
	case context.Config.Get("HOME_SNAPSHOTS") != "yes":
		if fsutil.IsFile(homeSnapperConfig()) {
			context.Run.Step("Stop taking snapshots of /home", homeSnapshotsStop)
		}
		return
	case !homeApart():
		context.Run.Note("/home isn't a btrfs subvolume of its own, so there are no hourly snapshots of it.")
		return
	}
	context.Run.Step("Keep ~/.cache and Steam's games out of the snapshots of /home", context.homeKeptOut)
	context.Run.Step("Take a snapshot of /home every hour", context.homeSnapperConfigure)
}

func (context *Context) homeSnapshotChecks() {
	if context.Config.Get("HOME_SNAPSHOTS") == "yes" && homeApart() {
		context.Run.Check("Snapshots of /home are set up", homeSnapperReady)
	}
}
