package care

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/care"
	"github.com/Ahmadlolbit/solanumos/internal/config"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/hypr"
	"github.com/Ahmadlolbit/solanumos/internal/notify"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/snapshots"
)

const usage = `Usage: solanum-care [firmware|disk|hyprland]

Looks after the system quietly. firmware tells you about new firmware from
fwupd, once per version. disk warns when / is 90% full, once a day, and
offers to clean up old snapshots. hyprland checks that Hyprland still takes
its configuration, as the next login needs, and says so once per complaint.
Without an argument it does all three, the firmware and Hyprland once a day;
Firmware and disk warnings off leaves out the first two. A user timer runs it
every hour.
`

const (
	diskLimit   = 90
	nothingToDo = 2
)

func save(path, text string) {
	err := os.MkdirAll(filepath.Dir(path), 0o777)
	if err == nil {
		err = atomic.Write(path, []byte(text), 0o666)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func stateFolder() string {
	return filepath.Join(env.StateHome(), "solanum")
}

func stamp(name string) string {
	return filepath.Join(env.Or("XDG_RUNTIME_DIR", stateFolder()), "solanum-"+name)
}

func today() string {
	return time.Now().Format(time.DateOnly)
}

func offer(app, urgency, summary, body string) bool {
	return notify.Post([]string{"--app-name=" + app, "--urgency=" + urgency, "--expire-time=0", "--hint=string:x-dunst-stack-tag:" + app}, summary, body) == 0
}

func firmware() bool {
	output, err := run.Capture(run.Minute, "fwupdmgr", "get-updates", "--json")
	waiting := care.Updates([]byte(output))
	seen := filepath.Join(stateFolder(), "firmware-seen")
	fresh := care.Unseen(waiting, fsutil.Optional(seen))
	if len(fresh) > 0 {
		if !offer("solanum-firmware", "normal", "Firmware update", "Click to update.\n"+strings.Join(fresh, "\n")) {
			return false
		}
		save(seen, strings.Join(waiting, "\n")+"\n")
	}
	return err == nil || run.Code(err) == nothingToDo
}

func dailyFirmware() {
	checked := stamp("firmware-checked")
	if fsutil.Optional(checked) != today() && firmware() {
		save(checked, today())
	}
}

func warned(use int) bool {
	title := fmt.Sprintf("The disk is %d%% full", use)
	all, _ := snapshots.All()
	if len(snapshots.Disposable(all)) == 0 {
		return offer("solanum", "critical", title, "")
	}
	return offer("solanum-disk", "critical", title, fmt.Sprintf("%d snapshots are on it. Click to delete the oldest automatic ones.", len(all)))
}

func disk() {
	full, err := strconv.Atoi(os.Getenv("SOLANUM_DISK_LIMIT"))
	if err != nil {
		full = diskLimit
	}
	use, ok := snapshots.DiskUse()
	warnedToday := stamp("disk-warned")
	if ok && use >= full && fsutil.Optional(warnedToday) != today() && warned(use) {
		save(warnedToday, today())
	}
}

func turnedOff() bool {
	return config.Read(config.Path()).Get("CARE_NOTIFY") != "yes"
}

func hyprland() bool {
	seen := filepath.Join(stateFolder(), "hyprland-refused")
	ok, complaint := hypr.Accepts(io.Discard, account.Account{})
	tell, keep := care.Refusal(ok, complaint, fsutil.Optional(seen))
	if tell && !offer("solanum-hyprland", "critical", "Hyprland may not start at the next login", keep+"\nBefore you log out: sudo solanum doctor, or solanum restore to go back.") {
		return false
	}
	if keep == "" {
		if err := os.Remove(seen); err != nil && !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, err)
		}
		return true
	}
	save(seen, keep)
	return true
}

func dailyHyprland() {
	checked := stamp("hyprland-checked")
	if _, err := exec.LookPath("Hyprland"); err == nil && fsutil.Optional(checked) != today() && hyprland() {
		save(checked, today())
	}
}

func Main() {
	switch env.Arg(1) {
	case "":
		dailyHyprland()
		if turnedOff() {
			return
		}
		disk()
		dailyFirmware()
	case "firmware":
		firmware()
	case "disk":
		disk()
	case "hyprland":
		hyprland()
	case "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
}
