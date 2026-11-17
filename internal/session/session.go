package session

import (
	"cmp"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	sessionLog   = ".local/share/sddm/wayland-session.log"
	crashReports = ".cache/hyprland/hyprlandCrashReport*.txt"
	slack        = time.Minute
	logBytes     = 64 << 10
	reportBytes  = 1 << 20
	shownLines   = 3
	longestLine  = 160
	headBytes    = 256
)

var errorLine = regexp.MustCompile(`(?i)\[err\]|error|critical|fatal|fail|abort|assert|segmentation|terminate called|cannot|couldn't|can't|unable`)

type Line struct {
	From string
	Text string
}

type Ending struct {
	Crashed  bool
	Frame    string
	Graphics bool
	Lines    []Line
}

type dump struct {
	Time int64 `json:"time"`
	PID  int64 `json:"pid"`
}

func printable(text string) string {
	return strings.Map(func(letter rune) rune {
		if unicode.IsControl(letter) {
			return -1
		}
		return letter
	}, text)
}

func clipped(line string) string {
	line = strings.TrimSpace(printable(line))
	if runes := []rune(line); len(runes) > longestLine {
		return string(runes[:longestLine]) + "…"
	}
	return line
}

func errorLines(text string, most int) []string {
	var found []string
	for _, line := range strings.Split(text, "\n") {
		if errorLine.MatchString(line) {
			found = append(found, clipped(line))
		}
	}
	return found[max(len(found)-most, 0):]
}

func (ending *Ending) note(from string, texts ...string) {
	for _, text := range texts {
		ending.Lines = append(ending.Lines, Line{printable(from), text})
	}
}

func (ending *Ending) crash(frames []frame) {
	ending.Crashed = true
	ending.Frame = cmp.Or(ending.Frame, clipped(story(frames)))
	ending.Graphics = ending.Graphics || startingGraphics(frames)
}

func dumps(output string) []dump {
	var all []dump
	if json.Unmarshal([]byte(output), &all) == nil {
		return all
	}
	decoder := json.NewDecoder(strings.NewReader(output))
	for {
		var one dump
		if decoder.Decode(&one) != nil {
			return all
		}
		all = append(all, one)
	}
}

func (ending *Ending) fromDump(uid string, since time.Time, limit time.Duration) {
	matches := []string{"COREDUMP_UID=" + uid, "COREDUMP_COMM=" + ownModule}
	output, err := run.CaptureQuiet(limit, "coredumpctl", append([]string{"--no-pager", "--json=short", "list"}, matches...)...)
	all := dumps(output)
	if err != nil || len(all) == 0 {
		return
	}
	newest := slices.MaxFunc(all, func(first, second dump) int { return cmp.Compare(first.Time, second.Time) })
	when := time.UnixMicro(newest.Time)
	if when.Before(since) {
		return
	}
	info, _ := run.CaptureQuiet(limit, "coredumpctl", append([]string{"--no-pager", "-1", "info"}, matches...)...)
	ending.crash(dumpFrames(info))
	pid := strconv.FormatInt(newest.PID, 10)
	ending.note("coredumpctl info "+pid, "Hyprland "+pid+" dumped core on "+when.Format("Mon 2 Jan at 15:04"))
}

func newestReport(home string) (string, time.Time) {
	found, _ := filepath.Glob(filepath.Join(home, crashReports))
	newest, when := "", time.Time{}
	for _, path := range found {
		info, err := os.Lstat(path)
		if err == nil && info.Mode().IsRegular() && info.ModTime().After(when) {
			newest, when = path, info.ModTime()
		}
	}
	return newest, when
}

func reportLines(text string) []string {
	head, tail, _ := strings.Cut(text, "Log tail:")
	var lines []string
	for _, line := range strings.Split(head, "\n") {
		if strings.Contains(line, "received signal") {
			lines = append(lines, clipped(line))
			break
		}
	}
	return append(lines, errorLines(tail, shownLines-1)...)
}

func (ending *Ending) fromReport(home string, uid int, since time.Time) {
	path, when := newestReport(home)
	if path == "" || when.Before(since) {
		return
	}
	data, err := fsutil.ReadPlain(path, reportBytes, uid)
	if err != nil {
		return
	}
	ending.crash(reportFrames(string(data)))
	ending.note("~/"+filepath.Join(filepath.Dir(crashReports), filepath.Base(path)), reportLines(string(data))...)
}

func (ending *Ending) fromLog(home string, uid int) {
	if data, err := fsutil.ReadTail(filepath.Join(home, sessionLog), logBytes, uid); err == nil {
		ending.note("~/"+sessionLog, errorLines(string(data), shownLines)...)
	}
}

func (ending *Ending) fromOwnLog(home string, uid int) {
	if data, err := fsutil.ReadTail(filepath.Join(home, OwnLog), logBytes, uid); err == nil {
		ending.note("~/"+OwnLog, errorLines(string(data), shownLines)...)
	}
}

func loggedIn(home string) time.Time {
	info, err := os.Lstat(filepath.Join(home, sessionLog))
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func ownStart(home string, uid int) time.Time {
	head, err := fsutil.ReadHead(filepath.Join(home, OwnLog), headBytes, uid)
	stamp, found := strings.CutPrefix(string(head), logStamp)
	if err != nil || !found || len(stamp) < len(time.DateTime) {
		return time.Time{}
	}
	when, err := time.ParseInLocation(time.DateTime, stamp[:len(time.DateTime)], time.Local)
	if err != nil {
		return time.Time{}
	}
	return when
}

func Last(person account.Account, limit time.Duration) Ending {
	var ending Ending
	uid, err := strconv.Atoi(person.UID)
	if err != nil || !filepath.IsAbs(person.Home) {
		return ending
	}
	own := ownStart(person.Home, uid)
	since := loggedIn(person.Home)
	if own.After(since) {
		since = own
	}
	if !since.IsZero() {
		since = since.Add(-slack)
	}
	ending.fromDump(person.UID, since, limit)
	ending.fromReport(person.Home, uid, since)
	ending.fromLog(person.Home, uid)
	if !own.IsZero() && !own.Before(since) {
		ending.fromOwnLog(person.Home, uid)
	}
	return ending
}
