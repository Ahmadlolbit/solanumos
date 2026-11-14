package main

import (
	"cmp"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/config"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/lock"
	"github.com/Ahmadlolbit/solanumos/internal/oldname"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
)

func installed(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func git(args ...string) string {
	output, _ := run.Capture(run.Slow, "git", append([]string{"-C", root()}, args...)...)
	return output
}

func gitLog(args ...string) string {
	return git(append([]string{"log", "--no-show-signature"}, args...)...)
}

var snapshotTaken bool

func snapshotsOff() bool {
	if config.Read(config.Path()).Get("SNAPSHOTS") == "yes" {
		return false
	}
	ui.Print("Snapshots are off, so there's no restore point before this\n")
	return true
}

func snapshotFirst() int {
	if snapshotTaken {
		return 0
	}
	if snapshotsOff() {
		snapshotTaken = true
		return 0
	}
	if !installed("snapper") || !fsutil.IsFile(env.Or("SOLANUM_SNAPPER_CONFIG", "/etc/snapper/configs/root")) {
		ui.Print("Snapshots aren't set up, so there's no restore point.\n")
		snapshotTaken = true
		return 0
	}
	ui.Print("Taking a snapshot first\n")
	if run.Status("sudo", "snapper", "--no-dbus", "-c", "root", "create", "--description", "before solanum update", "--cleanup-algorithm", "number") != 0 {
		return fail("the snapshot failed, so nothing was updated")
	}
	snapshotTaken = true
	return 0
}

func updateSystem() int {
	if status := snapshotFirst(); status != 0 {
		return status
	}
	if apt.Interrupted() {
		ui.Print("dpkg was interrupted, finishing what it left\n")
		if run.Status("sudo", "dpkg", "--force-confdef", "--force-confold", "--configure", "-a") != 0 {
			return fail("finishing what dpkg left failed, so nothing was updated")
		}
	}
	ui.Print("Updating Debian\n")
	if run.Status("sudo", "apt", "update") != 0 {
		ui.Print("apt couldn't refresh every list (see above), so Debian updates from the lists it has. sudo solanum install turns off a vendor repository that apt keeps rejecting.\n")
	}
	if upgrade := fullUpgrade(); run.Status(upgrade[0], upgrade[1:]...) != 0 {
		return fail("updating Debian failed")
	}
	return 0
}

func fullUpgrade() []string {
	if !streamed {
		return []string{"sudo", "apt", "full-upgrade"}
	}
	return []string{"sudo", "env", "DEBIAN_FRONTEND=noninteractive", "NEEDRESTART_MODE=a", "APT_LISTCHANGES_FRONTEND=none",
		"apt", "-y", "-o", "Dpkg::Options::=--force-confdef", "-o", "Dpkg::Options::=--force-confold", "full-upgrade"}
}

func updateApps() int {
	for _, tool := range []struct {
		name, what, action, unasked string
	}{
		{"flatpak", "Flatpak apps", "update", "--noninteractive"},
		{"mise", "mise's tools", "upgrade", "--yes"},
		{"rustup", "Rust", "update", ""},
	} {
		if !installed(tool.name) {
			continue
		}
		ui.Print("Updating " + tool.what + "\n")
		args := []string{tool.action}
		if streamed && tool.unasked != "" {
			args = append(args, tool.unasked)
		}
		if run.Status(tool.name, args...) != 0 {
			return fail("updating %s failed", tool.what)
		}
	}
	return 0
}

func firstLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if line != "" {
			return line
		}
	}
	return ""
}

func mergeHeading(subject, body string) string {
	heading := cmp.Or(firstLine(body), subject)
	rest, merged := strings.CutPrefix(subject, "Merge pull request #")
	if number, _, spaced := strings.Cut(rest, " "); merged && spaced {
		return "#" + number + " " + heading
	}
	return heading
}

func changesBetween(from, to string) string {
	var text strings.Builder
	for _, merge := range strings.Split(gitLog("--first-parent", "--merges", "-z", "--format=%P%n%s%n%b", from+".."+to), "\x00") {
		parents, message, _ := strings.Cut(merge, "\n")
		subject, body, _ := strings.Cut(message, "\n")
		fields := strings.Fields(parents)
		if len(fields) < 2 {
			continue
		}
		text.WriteString(mergeHeading(subject, body) + "\n")
		if merged := gitLog("--no-merges", "--format=  - %s", fields[0]+".."+fields[1]); merged != "" {
			text.WriteString(merged + "\n")
		}
	}
	if others := gitLog("--first-parent", "--no-merges", "--format=  - %s", from+".."+to); others != "" {
		text.WriteString("Other changes\n" + others + "\n")
	}
	return text.String()
}

func writeChanges(dir, text string) error {
	if err := os.MkdirAll(dir, 0o777); err != nil {
		return err
	}
	if err := atomic.Write(filepath.Join(dir, "changes.txt"), []byte(text), 0o666); err != nil {
		return err
	}
	return atomic.Write(filepath.Join(dir, "changes-unseen"), nil, 0o666)
}

func updatePart() string {
	part := cmp.Or(env.Arg(2), "all")
	if part == oldname.UpdatePart {
		return "solanum"
	}
	return part
}

func update() int {
	part := updatePart()
	if !slices.Contains([]string{"all", "system", "apps", "solanum"}, part) || env.Arg(3) != "" {
		return fail("usage: solanum update [system | apps | solanum]")
	}
	if os.Geteuid() == 0 {
		return fail("run it as yourself, without sudo: solanum update")
	}
	if lock.Busy() {
		return fail("a Solanum run is still going, so nothing was updated: run solanum update once it ends")
	}
	wanted := func(name string) bool { return part == "all" || part == name }
	if wanted("solanum") && run.Quiet(run.Quick, "git", "-C", root(), "rev-parse", "--git-dir") != 0 {
		return fail("%s is not a git clone, so it can't update itself", root())
	}
	for _, step := range []struct {
		name string
		work func() int
	}{{"system", updateSystem}, {"apps", updateApps}, {"solanum", updateSolanum}} {
		if wanted(step.name) {
			if status := step.work(); status != 0 {
				return status
			}
		}
	}
	return 0
}
