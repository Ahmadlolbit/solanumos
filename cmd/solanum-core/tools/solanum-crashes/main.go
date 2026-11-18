package notice

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/crashes"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/notify"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/session"
)

const usage = `Usage: solanum-crashes [show PID]

Tells you when a program of yours stops unexpectedly. Without an argument it
follows the journal for systemd-coredump's word of your programs' crashes and
shows a notification for each: "<program> stopped unexpectedly". Click it to
see what happened, or middle-click it to write a report (sudo solanum
collect). It says so at most once an hour for each program, and the next
notice counts the crashes in between. Hyprland's own crash is told at the
next login instead. show prints the journal's last lines from the process
PID, this boot, and what coredumpctl kept of its crash. A user service runs
it.
`

const (
	appName   = "solanum-crash"
	desktop   = "Hyprland"
	terminal  = "solanum-terminal"
	ownName   = "solanum-crashes"
	journaled = "40"
	lineSize  = 1 << 20
	notFound  = 127
)

var options = []string{"--app-name=" + appName, "--expire-time=0", "--action=default=What happened"}

func marker() string {
	return filepath.Join(env.Or("XDG_RUNTIME_DIR", env.OwnState()), session.Marker)
}

func fromDesktop() {
	for _, line := range session.Desktop(marker()) {
		name, value, _ := strings.Cut(line, "=")
		os.Setenv(name, value)
	}
}

func act(action string, what ...string) int {
	if action != "default" {
		return 0
	}
	fromDesktop()
	return run.Status(terminal, append([]string{ownName}, what...)...)
}

func tell(crash crashes.Crash, missed int) {
	action, _ := notify.Shown(options, crashes.Summary(crash.Program), crashes.Body(missed))
	act(action, "show", crash.PID)
}

func watch(input io.Reader, limit *crashes.Limit, notice func(crashes.Crash, int)) {
	lines := bufio.NewScanner(input)
	lines.Buffer(nil, lineSize)
	for lines.Scan() {
		crash, ok := crashes.Parse(lines.Text())
		if !ok || crash.Program == desktop {
			continue
		}
		if shown, missed := limit.Take(crash.Program, time.Now()); shown {
			notice(crash, missed)
		}
	}
}

func follow() int {
	args := append([]string{"--follow", "--lines=0"}, crashes.Journal("COREDUMP_UID="+strconv.Itoa(os.Getuid()))...)
	child, err := run.Spawn(nil, "journalctl", args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	watch(child.Output, crashes.NewLimit(), func(crash crashes.Crash, missed int) { go tell(crash, missed) })
	code, message := child.Wait()
	fmt.Fprintf(os.Stderr, "journalctl ended (exit %d) %s\n", code, message)
	return 1
}

func said(text string, err error, none string) string {
	switch {
	case run.Code(err) == notFound:
		return none + " It isn't installed.\n"
	case strings.TrimSpace(text) == "":
		return none + "\n"
	}
	return strings.TrimRight(text, "\n") + "\n"
}

func whatHappened(pid string) string {
	lines, err := run.CaptureQuiet(run.Minute, "journalctl", "--no-pager", "--quiet", "--boot", "--lines="+journaled, "_PID="+pid)
	info, failed := run.CaptureQuiet(run.Minute, "coredumpctl", "--no-pager", "info", pid)
	return "The journal's last lines from process " + pid + ", this boot:\n\n" + said(lines, err, "None.") +
		"\nWhat coredumpctl kept of its crash (coredumpctl info " + pid + "):\n\n" + said(info, failed, "Nothing.")
}

func show(pid string) int {
	if number, err := strconv.Atoi(pid); err != nil || number <= 0 {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	run.Page(whatHappened(pid))
	return 0
}

func Main() {
	switch env.Arg(1) {
	case "":
		os.Exit(follow())
	case "show":
		os.Exit(show(env.Arg(2)))
	case "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
}
