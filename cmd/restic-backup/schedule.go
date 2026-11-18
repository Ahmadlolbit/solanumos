package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/battery"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/notify"
	"github.com/Ahmadlolbit/solanumos/internal/restic"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	timerUnit   = "restic-backup.timer"
	serviceUnit = "restic-backup.service"
	dayLayout   = "2006-01-02"
)

type state struct {
	Location string `json:"location"`
	Schedule string `json:"schedule"`
	USB      bool   `json:"usb"`
	Last     string `json:"last"`
	Stale    bool   `json:"stale"`
	Charger  bool   `json:"charger"`
}

func userUnit(name string) string {
	return filepath.Join(env.ConfigHome(), "systemd", "user", name)
}

func stateDir() string {
	return filepath.Join(env.StateHome(), "restic-backup")
}

func stamp(name string) string {
	return filepath.Join(stateDir(), name)
}

func stamped(name string) string {
	return strings.TrimSpace(fsutil.Optional(stamp(name)))
}

func note(name, text string) {
	if err := os.MkdirAll(stateDir(), 0o777); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if err := atomic.Write(stamp(name), []byte(text+"\n"), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func lastBackup() time.Time {
	last, _ := time.Parse(time.RFC3339, stamped("last"))
	return last
}

func scheduledSince() time.Time {
	info, err := os.Stat(userUnit(timerUnit))
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func onCharger() bool {
	reading, found := battery.Read()
	return !found || reading.Plugged
}

func location() string {
	return strings.TrimSpace(fsutil.Optional(repositoryFile()))
}

func status(args []string) int {
	if len(args) != 1 || args[0] != "--json" {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	now, last := time.Now(), lastBackup()
	found := state{Location: location(), Schedule: "off", USB: fsutil.IsFile(unitFile()), Charger: onCharger()}
	if fsutil.IsFile(userUnit(timerUnit)) {
		found.Schedule = "daily"
	}
	if !last.IsZero() {
		found.Last = last.Format(time.RFC3339)
	}
	found.Stale = found.Location != "" && restic.Stale(last, scheduledSince(), now)
	data, err := json.Marshal(found)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(string(data))
	return 0
}

func listJSON() int {
	if !useRepository() {
		return 1
	}
	output, err := run.Capture(run.Slow, "restic", "snapshots", "--json")
	if err != nil {
		return 1
	}
	backups, err := restic.Backups([]byte(output))
	if err != nil {
		fmt.Fprintln(os.Stderr, "restic's list of backups can't be read:", err)
		return 1
	}
	data, err := json.Marshal(backups)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(string(data))
	return 0
}

func list(args []string) int {
	switch {
	case len(args) == 0:
		return repository("snapshots")
	case len(args) == 1 && args[0] == "--json":
		return listJSON()
	}
	fmt.Fprint(os.Stderr, usage)
	return 2
}

func scheduleDaily() int {
	if !useRepository() {
		return 1
	}
	program, err := self()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := os.MkdirAll(filepath.Dir(userUnit(timerUnit)), 0o777); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for name, text := range map[string]string{serviceUnit: restic.DailyService(program), timerUnit: restic.DailyTimer()} {
		if err := atomic.Write(userUnit(name), []byte(text), 0o666); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	if code := run.Within(serviceLimit, "systemctl", "--user", "daemon-reload"); code != 0 {
		return code
	}
	if code := run.Within(serviceLimit, "systemctl", "--user", "enable", "--now", timerUnit); code != 0 {
		return code
	}
	fmt.Println("Backups now run once a day, while the computer is on the charger and the backup location can be reached.")
	return 0
}

func scheduleOff() int {
	run.Quiet(serviceLimit, "systemctl", "--user", "disable", "--now", timerUnit)
	for _, name := range []string{timerUnit, serviceUnit} {
		if err := os.Remove(userUnit(name)); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Daily backups are still on: %v\n", err)
			return 1
		}
	}
	run.Within(serviceLimit, "systemctl", "--user", "daemon-reload")
	fmt.Println("Daily backups are off.")
	return 0
}

func schedule(args []string) int {
	switch {
	case len(args) == 1 && args[0] == "daily":
		return scheduleDaily()
	case len(args) == 1 && args[0] == "off":
		return scheduleOff()
	}
	fmt.Fprint(os.Stderr, usage)
	return 2
}

func remind(last, now time.Time) {
	if !restic.Stale(last, scheduledSince(), now) || stamped("told") == now.Format(dayLayout) {
		return
	}
	body := "There's no backup yet."
	if !last.IsZero() {
		body = "The last one was on " + last.Format("Monday 2 January") + "."
	}
	notify.Post([]string{"--app-name=solanum-backup", "--urgency=critical"}, "No backup for a week",
		body+" Backups run once a day while the computer is on the charger and the backup location can be reached.")
	note("told", now.Format(dayLayout))
}

func scheduled([]string) int {
	now, last := time.Now(), lastBackup()
	if !restic.Due(last, now) {
		return 0
	}
	code := 0
	if where := location(); where != "" && onCharger() && restic.Reachable(where) {
		if code = backup(); code == 0 || code == unreadable {
			return 0
		}
	}
	remind(last, now)
	return code
}

func files(args []string) int {
	if len(args) != 3 || args[2] != "--json" {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	if !useRepository() {
		return 1
	}
	folder := filepath.Clean(args[1])
	output, err := run.Capture(run.Slow, "restic", "ls", "--json", args[0], folder)
	if err != nil {
		return 1
	}
	data, err := json.Marshal(restic.Entries(output, folder))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(string(data))
	return 0
}

func backupWhen(id string) string {
	output, err := run.CaptureQuiet(run.Slow, "restic", "snapshots", "--json", id)
	backups, unread := restic.Backups([]byte(output))
	if err != nil || unread != nil || len(backups) == 0 {
		return "backup " + id
	}
	at, err := time.Parse(time.RFC3339, backups[0].Time)
	if err != nil {
		return "backup " + id
	}
	return at.Local().Format(restic.WhenLayout)
}

func restorePath(args []string) int {
	if (len(args) != 2 && (len(args) != 3 || args[2] != "--copy")) || !restic.Inside(args[1], env.Home()) {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	if !useRepository() {
		return 1
	}
	path := filepath.Clean(args[1])
	work, err := os.MkdirTemp(env.Home(), ".restic-backup-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(work)
	if code := run.Status("restic", "restore", args[0], "--target", work, "--include", restic.Literal(path)); code != 0 {
		return code
	}
	return placeRestored(filepath.Join(work, path), path, backupWhen(args[0]), len(args) == 3)
}

func placeRestored(source, path, when string, asCopy bool) int {
	info, err := os.Lstat(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s isn't in that backup.\n", path)
		return 1
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	placed, aside, err := restic.PutBack(source, path, when, info.IsDir(), asCopy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s couldn't be restored: %v\n", path, err)
		return 1
	}
	fmt.Println("Restored " + placed)
	if aside != "" {
		fmt.Println("What was there is now " + aside)
	}
	return 0
}
