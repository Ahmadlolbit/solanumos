package move

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/bootimage"
	"github.com/Ahmadlolbit/solanumos/internal/config"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/oldname"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	stampFile  = "moved"
	markFile   = "moving"
	laterFile  = "boot-later"
	described  = "before the move to Solanum"
	folderMode = 0o755
	stampMode  = 0o644
)

var bootCommands = []string{"update-initramfs", "dracut", "update-grub"}

type mover struct {
	out      io.Writer
	notes    []string
	snapshot string
	unlisted bool
	kept     []string
}

func at(path string) string {
	return filepath.Join(env.Or("SOLANUM_MOVE_ROOT", "/"), path)
}

func stateDir() string {
	return at("/var/lib/solanum")
}

func markers() []string {
	return []string{filepath.Join(at("/var/lib/"+oldname.Name), markFile), filepath.Join(stateDir(), markFile)}
}

func oldFolders() bool {
	for _, each := range folders {
		if fsutil.IsDir(at(each.old)) && !fsutil.IsLink(at(each.old)) {
			return true
		}
	}
	return false
}

func Pending() bool {
	if fsutil.Exists(filepath.Join(stateDir(), stampFile)) {
		return false
	}
	return slices.ContainsFunc(markers(), fsutil.Exists) || oldFolders()
}

func Unfinished() bool {
	return Pending() && !oldFolders()
}

func Snapshot() string {
	for _, line := range strings.Split(fsutil.Optional(filepath.Join(stateDir(), stampFile)), "\n") {
		if number, ok := strings.CutPrefix(line, "snapshot "); ok {
			return strings.TrimSpace(number)
		}
	}
	return ""
}

func (work *mover) note(format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	work.notes = append(work.notes, text)
	fmt.Fprintln(work.out, text)
}

func (work *mover) marked() string {
	for _, path := range markers() {
		if text, err := fsutil.Text(path); err == nil {
			return strings.TrimSpace(text)
		}
	}
	return ""
}

func snapshotsWanted() bool {
	settings := config.Read(config.Path())
	return settings.Get("SNAPSHOTS") == "yes" && fsutil.IsFile(env.Or("SOLANUM_SNAPPER_CONFIG", "/etc/snapper/configs/root"))
}

func newestSnapshot() string {
	output, _ := run.CaptureQuiet(run.Minute, "snapper", "--no-dbus", "-c", "root", "--csvout", "--no-headers", "list", "--columns", "number")
	newest := 0
	for _, line := range run.Lines(output) {
		if number, err := strconv.Atoi(strings.TrimSpace(line)); err == nil && number > newest {
			newest = number
		}
	}
	if newest == 0 {
		return ""
	}
	return strconv.Itoa(newest)
}

func (work *mover) takeSnapshot() {
	if work.snapshot = work.marked(); work.snapshot != "" {
		return
	}
	if !snapshotsWanted() {
		work.snapshot = newestSnapshot()
		return
	}
	number, err := run.CaptureQuiet(run.Minute, "snapper", "--no-dbus", "-c", "root", "create", "--description", described, "--cleanup-algorithm", "number", "--print-number")
	if err != nil {
		work.note("The snapshot before the move failed, so only an older snapshot can undo it.")
		work.snapshot = newestSnapshot()
		return
	}
	work.snapshot = strings.TrimSpace(number)
	fmt.Fprintf(work.out, "took snapshot %s, %s\n", work.snapshot, described)
}

func (work *mover) mark() {
	dir := at("/var/lib/" + oldname.Name)
	if !fsutil.IsDir(dir) || fsutil.IsLink(dir) {
		dir = stateDir()
	}
	err := os.MkdirAll(dir, folderMode)
	if err == nil {
		err = atomic.Install(filepath.Join(dir, markFile), []byte(work.snapshot+"\n"), stampMode)
	}
	if err == nil {
		err = atomic.Install(filepath.Join(dir, laterFile), nil, stampMode)
	}
	if err != nil {
		fmt.Fprintln(work.out, "cannot note that the move started:", err)
	}
}

func (work *mover) stamp() error {
	text := fmt.Sprintf("snapshot %s\nwhen %s\n", work.snapshot, time.Now().Format(time.DateTime))
	if err := atomic.Install(filepath.Join(stateDir(), stampFile), []byte(text), stampMode); err != nil {
		return err
	}
	os.Remove(filepath.Join(stateDir(), markFile))
	return nil
}

func Run(out io.Writer) []string {
	if !Pending() {
		return nil
	}
	work := &mover{out: out}
	fmt.Fprintln(out, "moving this computer from Hyprtrixie to Solanum")
	work.takeSnapshot()
	work.mark()
	run.Hold(bootCommands...)
	work.stopUnits()
	work.moveFolders()
	work.moveRecords()
	work.moveCommands()
	work.moveHomes()
	work.moveRest()
	if work.unlisted {
		work.note("The move to Solanum isn't finished: run sudo solanum install again to finish it.")
		return work.notes
	}
	if err := work.stamp(); err != nil {
		work.note("Couldn't note that the move is done (%v): run sudo solanum install again to finish it.", err)
		return work.notes
	}
	work.note("This computer has moved from Hyprtrixie to Solanum. Log out and back in once to finish.")
	return work.notes
}

func Due() bool {
	return fsutil.Exists(filepath.Join(stateDir(), laterFile))
}

func Hold() {
	if Due() {
		run.Hold(bootCommands...)
	}
}

func Settle(out io.Writer) error {
	run.Release()
	if !Due() {
		return nil
	}
	rebuild := bootimage.Rebuild(true)
	if err := run.Logged(out, rebuild[0], rebuild[1:]...); err != nil {
		return err
	}
	if apt.Commands("update-grub") == nil {
		if err := run.Logged(out, "update-grub"); err != nil {
			return err
		}
	}
	return os.Remove(filepath.Join(stateDir(), laterFile))
}
