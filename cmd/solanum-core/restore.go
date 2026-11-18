package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/Ahmadlolbit/solanumos/internal/cleanup"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/lock"
	"github.com/Ahmadlolbit/solanumos/internal/move"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/snapshots"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
)

const (
	kernelFile  = "vmlinuz-"
	initrdFile  = "initrd.img-"
	mountField  = 4
	confirmWord = "CONFIRM\n"
)

func restoreRows(points []snapshots.Snapshot, format string) string {
	var rows strings.Builder
	for _, snapshot := range slices.Backward(points) {
		fmt.Fprintf(&rows, format, snapshot.Number, snapshot.Label())
	}
	return rows.String()
}

func separate(path, root string) bool {
	var one, other syscall.Stat_t
	if syscall.Stat(path, &one) != nil || syscall.Stat(root, &other) != nil {
		return false
	}
	return one.Dev != other.Dev
}

func mountedAt(table, path string) bool {
	for _, line := range strings.Split(table, "\n") {
		if fields := strings.Fields(line); len(fields) > mountField && fields[mountField] == path {
			return true
		}
	}
	return false
}

func mountPoint(path string) string {
	var one, up syscall.Stat_t
	if syscall.Stat(path, &one) != nil {
		return path
	}
	for path != "/" {
		if syscall.Stat(filepath.Dir(path), &up) != nil || up.Dev != one.Dev {
			return path
		}
		path = filepath.Dir(path)
	}
	return path
}

func linkedToMount(table, path string) bool {
	if !fsutil.IsLink(path) {
		return false
	}
	real, err := filepath.EvalSymlinks(path)
	return err == nil && mountedAt(table, mountPoint(real))
}

func keptWord(root, table, rel, inside, apart, nested string) string {
	path := filepath.Join(root, rel)
	switch {
	case !separate(path, root):
		return inside
	case mountedAt(table, "/"+rel) || linkedToMount(table, path):
		return apart
	}
	return nested
}

func keptWords(root, table string) string {
	return keptWord(root, table, "home", "/home is part of this snapshot, so your files go back too.",
		"Going back keeps your files in /home as they are.",
		"/home is a subvolume inside the root, so it comes up empty after going back, and your files stay in the old root.") +
		keptWord(root, table, "var/log", "", " The logs in /var/log stay as they are.",
			" /var/log comes up empty, and the old logs stay in the old root.") + "\n"
}

func movedBackWords(before, apart bool) string {
	if !before || !apart {
		return ""
	}
	return "That snapshot is from before the move to Solanum, so the folders and lines in each home go back to Hyprtrixie's names too.\n"
}

func kernelsIn(boot string) ([]string, error) {
	entries, err := os.ReadDir(boot)
	if err != nil {
		return nil, err
	}
	var kernels []string
	for _, entry := range entries {
		version, found := strings.CutPrefix(entry.Name(), kernelFile)
		if found && fsutil.IsFile(filepath.Join(boot, initrdFile+version)) {
			kernels = append(kernels, version)
		}
	}
	return kernels, nil
}

func newestKernel(boot string, kernels []string) (string, error) {
	args := []string{"-v", "--quoting-style=literal"}
	for _, kernel := range kernels {
		args = append(args, filepath.Join(boot, kernelFile+kernel))
	}
	sorted, err := run.CaptureQuiet(run.Quick, "ls", args...)
	if err == nil && sorted == "" {
		err = os.ErrNotExist
	}
	last := sorted[strings.LastIndexByte(sorted, '\n')+1:]
	return strings.TrimPrefix(filepath.Base(last), kernelFile), err
}

func kernelAdvice(snapshotModules, boot, root string) string {
	couldNot := "Couldn't read the kernels in " + boot + ", so this can't tell if the snapshot has one that will start.\n"
	if !fsutil.IsDir(boot) {
		return couldNot
	}
	if !separate(boot, root) {
		return ""
	}
	kernels, err := kernelsIn(boot)
	if err != nil || len(kernels) == 0 {
		return couldNot
	}
	if _, err := os.ReadDir(snapshotModules); err != nil {
		return "Couldn't read the kernels in this snapshot, so this can't tell if it has one that will start.\n"
	}
	newest, err := newestKernel(boot, kernels)
	if err != nil {
		return couldNot
	}
	inSnapshot := func(kernel string) bool { return fsutil.IsDir(filepath.Join(snapshotModules, kernel)) }
	if inSnapshot(newest) {
		return ""
	}
	left := slices.DeleteFunc(kernels, func(kernel string) bool { return !inSnapshot(kernel) })
	if len(left) == 0 {
		return "This snapshot has no kernel that will start: every kernel in " + boot + " came after it.\n"
	}
	pick, err := newestKernel(boot, left)
	if err != nil {
		return couldNot
	}
	return "This snapshot doesn't have the newest kernel, " + newest + ". When the machine restarts, press Esc right after the laptop's logo, " +
		"open Advanced options and pick " + pick + ". Do that at every restart until the next update brings " + newest + " back.\n"
}

func comparing(number string) int {
	text := "Comparing snapshot " + number + " with today, this can take a minute"
	if !streamed {
		return say(text + "...\n")
	}
	ui.Busy(text)
	return 0
}

type comparison struct {
	snapshots.Tally
	Number string `json:"number"`
	Kept   string `json:"kept"`
	Kernel string `json:"kernel"`
}

func compared(number string) (comparison, error) {
	status, err := run.Capture(run.Slow, "snapper", "-c", "root", "status", number+"..0")
	if err != nil {
		return comparison{}, err
	}
	modules := filepath.Join(env.Or("SOLANUM_SNAPSHOTS", "/.snapshots"), number, "snapshot/usr/lib/modules")
	return comparison{
		Tally:  snapshots.Compare(status),
		Number: number,
		Kept:   keptWords("/", fsutil.Optional("/proc/self/mountinfo")) + movedBackWords(move.Before(number), move.HomeApart()),
		Kernel: kernelAdvice(modules, env.Or("SOLANUM_BOOT", "/boot"), "/"),
	}, nil
}

func compareJSON(number string) int {
	found, err := compared(number)
	if err != nil {
		return fail("snapshot %s can't be read: see solanum restore", number)
	}
	found.Kept, found.Kernel = strings.TrimSpace(found.Kept), strings.TrimSpace(found.Kernel)
	return printed(found)
}

func goBack(number string) int {
	defer cleanup.IgnoreQuit()()
	if lock.Busy() {
		return fail("a Solanum run is still going, so nothing went back: run solanum restore %s once it ends", number)
	}
	if code := comparing(number); code != 0 {
		return code
	}
	found, err := compared(number)
	if streamed {
		ui.Done()
	}
	if err != nil {
		return fail("snapshot %s can't be read: see solanum restore", number)
	}
	if code := say(found.Text() + "\n" + found.Kept + found.Kernel + "\n"); code != 0 {
		return code
	}
	code := rollBack(number)
	if code == 0 && found.Kernel != "" {
		return say("\n" + found.Kernel)
	}
	return code
}

func rollBack(number string) int {
	command := admin("snapper-rollback", number)
	if !streamed {
		return run.Status(command[0], command[1:]...)
	}
	if !ui.Confirm("Go back to snapshot " + number + "?") {
		return fail("nothing changed")
	}
	return run.Typed(confirmWord, command[0], command[1:]...)
}

func showPoints(target string) int {
	points, err := snapshots.All()
	switch {
	case err != nil:
		return fail("the snapshots can't be read")
	case target == "list":
		return say(restoreRows(points, "%s\t%s\n"))
	case len(points) == 0:
		return say("There are no snapshots yet.\n")
	}
	return say(restoreRows(points, "%6s  %s\n") + "\nGo back to one: solanum restore NUMBER\n")
}

func pointsJSON() int {
	points, err := snapshots.All()
	if err != nil {
		return fail("the snapshots can't be read")
	}
	slices.Reverse(points)
	return printed(snapshots.Points(points))
}

func restoreAction() (func() int, bool) {
	target, more := env.Arg(2), env.Arg(3)
	switch {
	case target == "list" && more == "--json" && env.Arg(4) == "":
		return pointsJSON, true
	case target == "compare" && snapshots.IsNumber(more) && env.Arg(4) == "--json" && env.Arg(5) == "":
		return func() int { return compareJSON(more) }, true
	case more != "":
		return nil, false
	case target == "" || target == "list":
		return func() int { return showPoints(target) }, true
	case snapshots.IsNumber(target):
		return func() int { return goBack(target) }, true
	}
	return nil, false
}

func restore() int {
	action, ok := restoreAction()
	if !ok {
		return fail("usage: solanum restore [list | NUMBER]")
	}
	if !forAPerson() {
		return fail("run it as yourself, without sudo: solanum restore")
	}
	if !installed("snapper") {
		return fail("snapper isn't installed, so there are no restore points")
	}
	return action()
}
