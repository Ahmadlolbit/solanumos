package main

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/cleanup"
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

func admin(args ...string) []string {
	if _, helped := onBehalf(); helped {
		return args
	}
	return append([]string{"sudo"}, args...)
}

func asThePerson(args ...string) []string {
	if person, helped := onBehalf(); helped {
		return append([]string{"runuser"}, person.RunuserArgs(args...)...)
	}
	return args
}

func rootsOnly(path string) bool {
	if !filepath.IsAbs(path) {
		return false
	}
	for path = filepath.Join(path, ".git"); ; path = filepath.Dir(path) {
		info, err := os.Lstat(path)
		if err != nil || !info.IsDir() || info.Mode().Perm()&0o022 != 0 {
			return false
		}
		if stat, ok := info.Sys().(*syscall.Stat_t); !ok || stat.Uid != 0 {
			return false
		}
		if path == filepath.Dir(path) {
			return true
		}
	}
}

func gitCommand(args ...string) []string {
	command := append(append([]string{"git", "-C", root()}, readable(root())...), args...)
	if rootsOnly(root()) {
		return command
	}
	return asThePerson(command...)
}

func git(args ...string) string {
	command := gitCommand(args...)
	output, _ := run.Capture(run.Slow, command[0], command[1:]...)
	return output
}

func readable(path string) []string {
	info, err := os.Stat(path)
	if err != nil || os.Geteuid() == 0 {
		return nil
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); !ok || stat.Uid != 0 {
		return nil
	}
	return []string{"-c", "safe.directory=" + path}
}

func gitQuiet(limit time.Duration, args ...string) int {
	command := gitCommand(args...)
	return run.Quiet(limit, command[0], command[1:]...)
}

func gitOutput(limit time.Duration, args ...string) (string, error) {
	command := gitCommand(args...)
	return run.CaptureQuiet(limit, command[0], command[1:]...)
}

func gitLog(args ...string) string {
	return git(append([]string{"log", "--no-show-signature"}, args...)...)
}

const (
	holdBusy   = 75
	holdFailed = 74
	holdGrace  = 10 * time.Second
)

var errRunGoing = errors.New("a Solanum run is still going, so nothing was updated: run solanum update once it ends")

var (
	snapshotTaken bool
	debianFailed  bool
	updateHeld    struct {
		held  *lock.Held
		child *run.Child
	}
)

func holdForUpdate() error {
	if os.Geteuid() == 0 {
		held, err := lock.Take()
		updateHeld.held = held
		return heldOrWhy(err)
	}
	command := admin(script(), "update", "--hold")
	child, err := run.SpawnWith(run.Spawning{Shown: true}, command[0], command[1:]...)
	if err != nil {
		return fmt.Errorf("%v, so nothing was updated", err)
	}
	updateHeld.child = child
	if answer, _ := bufio.NewReader(child.Output).ReadString('\n'); answer == "held\n" {
		return nil
	}
	switch releaseUpdateHold() {
	case holdBusy:
		return errRunGoing
	case holdFailed:
		return errors.New("the run lock couldn't be taken, so nothing was updated")
	}
	return errors.New("sudo didn't run solanum, so nothing was updated")
}

func heldOrWhy(err error) error {
	switch {
	case errors.Is(err, lock.ErrBusy):
		return errRunGoing
	case err != nil:
		return fmt.Errorf("%v, so nothing was updated", err)
	}
	return nil
}

func releaseUpdateHold() int {
	code := 0
	if updateHeld.held != nil {
		updateHeld.held.Close()
	}
	if updateHeld.child != nil {
		code = letGo(updateHeld.child)
	}
	updateHeld.held, updateHeld.child = nil, nil
	return code
}

func letGo(child *run.Child) int {
	child.Input.Close()
	ended := make(chan int, 1)
	go func() {
		code, _ := child.Wait()
		ended <- code
	}()
	select {
	case code := <-ended:
		return code
	case <-time.After(holdGrace):
	}
	child.Stop()
	ui.Print("Letting go of the run lock took longer than 10 seconds, so its holder was stopped.\n")
	return <-ended
}

func updateHolding() bool {
	return updateHeld.held != nil || updateHeld.child != nil
}

func holdUntilDone() int {
	held, err := lock.Take()
	switch {
	case errors.Is(err, lock.ErrBusy):
		return holdBusy
	case err != nil:
		fmt.Fprintln(os.Stderr, err)
		return holdFailed
	}
	defer held.Close()
	cleanup.NotifyInterrupt(make(chan os.Signal, 1))
	ui.Print("held\n")
	io.Copy(io.Discard, os.Stdin)
	return 0
}

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
	if snapshot := admin("snapper", "--no-dbus", "-c", "root", "create", "--description", "before solanum update", "--cleanup-algorithm", "number"); run.Status(snapshot[0], snapshot[1:]...) != 0 {
		return fail("the snapshot failed, so nothing was updated")
	}
	snapshotTaken = true
	return 0
}

func ran(command []string, failure string) int {
	if run.Status(command[0], command[1:]...) == 0 {
		return 0
	}
	if run.Stopped() {
		return stoppedCode
	}
	return fail("%s", failure)
}

func apart(stoppable bool, command []string, failure string) int {
	if _, helped := onBehalf(); !streamed || !helped {
		return ran(command, failure)
	}
	var stop chan os.Signal
	if stoppable {
		stop = make(chan os.Signal, 1)
		cleanup.NotifyInterrupt(stop)
		defer signal.Stop(stop)
	}
	code, halted := run.Shielded(stop, command[0], command[1:]...)
	switch {
	case halted:
		run.Stop()
		return stoppedCode
	case code == 0:
		return 0
	case run.Stopped():
		return stoppedCode
	}
	return fail("%s", failure)
}

func updateSystem() int {
	if status := snapshotFirst(); status != 0 {
		return status
	}
	if apt.Interrupted() {
		ui.Print("dpkg was interrupted, finishing what it left\n")
		if status := apart(false, admin("dpkg", "--force-confdef", "--force-confold", "--configure", "-a"), "finishing what dpkg left failed, so Debian wasn't updated"); status != 0 {
			debianFailed = true
			return status
		}
		if run.Stopped() {
			return stoppedCode
		}
	}
	ui.Print("Updating Debian\n")
	if refresh := admin("apt", "update"); run.Status(refresh[0], refresh[1:]...) != 0 && !run.Stopped() {
		ui.Print("apt couldn't refresh every list (see above), so Debian updates from the lists it has. sudo solanum install turns off a vendor repository that apt keeps rejecting.\n")
	}
	if run.Stopped() {
		return stoppedCode
	}
	before := sessionVersions()
	status := apart(true, fullUpgrade(), "updating Debian failed")
	if run.Stopped() {
		return stoppedCode
	}
	if status == 0 {
		keptKernelNews()
	}
	if check := sessionCheck(before); check != 0 {
		return check
	}
	debianFailed = status != 0
	return status
}

func fullUpgrade() []string {
	if !streamed {
		return admin("apt", "full-upgrade")
	}
	return admin("env", "DEBIAN_FRONTEND=noninteractive", "NEEDRESTART_MODE=a", "APT_LISTCHANGES_FRONTEND=none",
		"apt", "-y", "-o", "Dpkg::Options::=--force-confdef", "-o", "Dpkg::Options::=--force-confold", "full-upgrade")
}

var appTools = []struct {
	name, what, action, unasked string
}{
	{"flatpak", "Flatpak apps", "update", "--noninteractive"},
	{"mise", "mise's tools", "upgrade", "--yes"},
	{"rustup", "Rust", "update", ""},
}

func updateApps() int {
	for _, tool := range appTools {
		if !installed(tool.name) {
			continue
		}
		ui.Print("Updating " + tool.what + "\n")
		command := asThePerson(tool.name, tool.action)
		if streamed && tool.unasked != "" {
			command = append(command, tool.unasked)
		}
		if status := ran(command, "updating "+tool.what+" failed"); status != 0 {
			return status
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

func notesFor(person account.Account, text string) error {
	dir := oldname.Prefer(person.Path(filepath.Join(".local", "state", "solanum")))
	if err := person.Run(io.Discard, "mkdir", "-p", dir); err != nil {
		return err
	}
	if err := person.Write(io.Discard, filepath.Join(dir, "changes.txt"), text); err != nil {
		return err
	}
	return person.Write(io.Discard, filepath.Join(dir, "changes-unseen"), "")
}

func saveChanges(text string) (string, error) {
	person, helped := onBehalf()
	if !helped {
		dir := changesDir()
		return filepath.Join(dir, "changes.txt"), writeChanges(dir, text)
	}
	if err := notesFor(person, text); err != nil {
		return "", err
	}
	if err := os.MkdirAll(env.StateDir(), 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(env.StateDir(), "update-notes.txt")
	return path, atomic.Write(path, []byte(text), 0o644)
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
	if !forAPerson() {
		return fail("run it as yourself, without sudo: solanum update")
	}
	if lock.Busy() {
		return fail("%v", errRunGoing)
	}
	wanted := func(name string) bool { return part == "all" || part == name }
	if wanted("solanum") && gitQuiet(run.Quick, "rev-parse", "--git-dir") != 0 {
		return fail("%s is not a git clone, so it can't update itself", root())
	}
	if wanted("system") || wanted("solanum") {
		if err := holdForUpdate(); err != nil {
			return fail("%v", err)
		}
		defer releaseUpdateHold()
	}
	stopsTheUpdate()
	return updateParts(wanted)
}

func stopsTheUpdate() {
	caught := make(chan os.Signal, 1)
	cleanup.Notify(caught)
	go func() {
		<-caught
		run.Stop()
	}()
}

func updateParts(wanted func(string) bool) int {
	failed := 0
	for _, step := range []struct {
		name string
		work func() int
	}{{"system", updateSystem}, {"apps", updateApps}, {"solanum", updateSolanum}} {
		if !wanted(step.name) {
			continue
		}
		status := step.work()
		if run.Stopped() {
			ui.Skip("Stopped, so the update ends here")
			return stoppedCode
		}
		if status != 0 && (step.name != "system" || !debianFailed) {
			return status
		}
		failed = max(failed, status)
	}
	return failed
}
