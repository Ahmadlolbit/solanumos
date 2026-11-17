package session

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	Marker    = "solanum-session-up"
	OwnLog    = ".local/state/solanum/session.log"
	GoodLogin = ".local/state/solanum/session-ok"
	TriedNote = ".local/state/solanum/kernel-tried"
	logMode   = 0o600
	ownMode   = 0o700
	logStamp  = "==== "
)

var (
	upAfter      = time.Minute
	desktopNames = []string{"WAYLAND_DISPLAY", "DISPLAY", "HYPRLAND_INSTANCE_SIGNATURE", "DBUS_SESSION_BUS_ADDRESS"}
)

type Outcome int

const (
	NeverStarted Outcome = iota
	Up
	Signalled
)

func opened(log string) *os.File {
	if os.MkdirAll(filepath.Dir(log), ownMode) == nil {
		if file, err := atomic.Rotate(log, logMode); err == nil {
			return file
		}
	}
	return os.Stderr
}

func Start(command []string, marker, log string, atMinute func()) Outcome {
	if err := os.Remove(marker); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(os.Stderr, err)
	}
	output := opened(log)
	fmt.Fprintf(output, "%s%s %s\n", logStamp, time.Now().Format(time.DateTime), strings.Join(command, " "))
	minute := time.AfterFunc(upAfter, atMinute)
	code, signalled := run.Attached(output, command[0], command[1:]...)
	ranAMinute := !minute.Stop()
	fmt.Fprintf(output, "%s%s %s ended (exit %d)\n", logStamp, time.Now().Format(time.DateTime), filepath.Base(command[0]), code)
	if output != os.Stderr {
		output.Close()
	}
	switch {
	case signalled:
		return Signalled
	case ranAMinute || code == 0 && fsutil.Exists(marker):
		return Up
	}
	return NeverStarted
}

func MarkUp(marker string) error {
	var lines strings.Builder
	for _, name := range desktopNames {
		if value, set := os.LookupEnv(name); set {
			lines.WriteString(name + "=" + value + "\n")
		}
	}
	return atomic.Write(marker, []byte(lines.String()), logMode)
}

func Desktop(marker string) []string {
	data, err := fsutil.ReadPlain(marker, logBytes)
	if err != nil {
		return nil
	}
	lines := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		if name, _, found := strings.Cut(line, "="); found && slices.Contains(desktopNames, name) {
			lines = append(lines, line)
		}
	}
	return lines
}

func MarkGood(good string) (string, error) {
	kernel, _ := run.CaptureQuiet(run.Quick, "uname", "-r")
	if err := os.MkdirAll(filepath.Dir(good), ownMode); err != nil {
		return kernel, err
	}
	return kernel, atomic.Write(good, []byte(kernel+"\n"), logMode)
}

func NoteTried(home, kernel string) error {
	path := filepath.Join(home, TriedNote)
	if err := os.MkdirAll(filepath.Dir(path), ownMode); err != nil {
		return err
	}
	return atomic.Write(path, []byte(kernel+"\n"), logMode)
}

func Tried(home string) string {
	return strings.TrimSpace(fsutil.Optional(filepath.Join(home, TriedNote)))
}

func TakeTried(home string) string {
	path := filepath.Join(home, TriedNote)
	kernel := strings.TrimSpace(fsutil.Optional(path))
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(os.Stderr, err)
	}
	return kernel
}
