package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/bootimage"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/grub"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	historyDate = "2006-01-02  15:04:05"
	shownDate   = "2006-01-02 at 15:04"
	tailLines   = 8
	screenWidth = 76
	goodBytes   = 256
)

type Newer struct {
	Kernel   string
	Previous string
	Came     string
}

type Facts struct {
	Ending   Ending
	Tail     []string
	Changed  string
	Newer    Newer
	Restore  string
	Writable bool
	CameUp   bool
}

type change struct {
	when time.Time
	text string
}

func historyText() string {
	return fsutil.Optional(env.Or("SOLANUM_APT_HISTORY", "/var/log/apt/history.log"))
}

func startOf(entry string) (time.Time, bool) {
	for _, line := range strings.Split(entry, "\n") {
		if stamp, found := strings.CutPrefix(line, "Start-Date: "); found {
			when, err := time.ParseInLocation(historyDate, stamp, time.Local)
			return when, err == nil
		}
	}
	return time.Time{}, false
}

func changes(text string) []change {
	var all []change
	for _, entry := range strings.Split(text, "\n\n") {
		if when, ok := startOf(entry); ok {
			all = append(all, change{when, entry})
		}
	}
	return all
}

func arrived(all []change, kernel string) string {
	name := "linux-image-" + kernel
	for _, each := range slices.Backward(all) {
		if strings.Contains(each.text, name+":") || strings.Contains(each.text, name+" (") {
			return each.when.Format(shownDate)
		}
	}
	return ""
}

func goodKernel(home string) string {
	data, err := fsutil.ReadPlain(filepath.Join(home, GoodLogin), goodBytes)
	if fields := strings.Fields(string(data)); err == nil && len(fields) > 0 {
		return fields[0]
	}
	return ""
}

func NewerKernel(home string, limit time.Duration) Newer {
	good := goodKernel(home)
	running, _ := run.CaptureQuiet(limit, "uname", "-r")
	kernels := bootimage.Kernels()
	at, was := slices.Index(kernels, running), slices.Index(kernels, good)
	if good == "" || was < 0 || at <= was {
		return Newer{}
	}
	return Newer{Kernel: running, Previous: good, Came: arrived(changes(historyText()), running)}
}

func changedSince(home string) string {
	info, err := os.Lstat(filepath.Join(home, GoodLogin))
	all := changes(historyText())
	if err != nil || len(all) == 0 || !all[len(all)-1].when.After(info.ModTime()) {
		return ""
	}
	return all[len(all)-1].when.Format(shownDate)
}

func plain(line string) string {
	line = strings.Map(func(letter rune) rune {
		if letter > '~' {
			return '?'
		}
		return letter
	}, strings.TrimSpace(printable(line)))
	if len(line) > screenWidth {
		return line[:screenWidth-3] + "..."
	}
	return line
}

func tailOf(path string, uid, most int) []string {
	data, _ := fsutil.ReadTail(path, logBytes, uid)
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if line = plain(line); line != "" && !strings.HasPrefix(line, logStamp) {
			lines = append(lines, line)
		}
	}
	return lines[max(len(lines)-most, 0):]
}

func tail(home string, uid int) []string {
	return tailOf(filepath.Join(home, OwnLog), uid, tailLines)
}

func Gather(person account.Account, limit time.Duration) Facts {
	uid, _ := strconv.Atoi(person.UID)
	facts := Facts{Ending: Last(person, limit), Tail: tail(person.Home, uid), Newer: NewerKernel(person.Home, limit), CameUp: fsutil.Exists(MarkerPath())}
	facts.Changed = changedSince(person.Home)
	if _, err := exec.LookPath("snapper"); err == nil {
		facts.Restore = FindWayOut(limit).Restore.Number
	}
	if facts.Newer.Previous != "" {
		facts.Writable = grub.EnvWritable()
	}
	return facts
}
