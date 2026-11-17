package session

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

var upAfter = time.Minute

type Outcome int

const (
	NeverStarted Outcome = iota
	Up
)

func opened(log string) *os.File {
	if os.MkdirAll(filepath.Dir(log), ownMode) == nil {
		if file, err := atomic.Rotate(log, logMode); err == nil {
			return file
		}
	}
	return os.Stderr
}

func Start(command []string, marker, log string) Outcome {
	if err := os.Remove(marker); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(os.Stderr, err)
	}
	output := opened(log)
	began := time.Now()
	fmt.Fprintf(output, "%s%s %s\n", logStamp, began.Format(time.DateTime), strings.Join(command, " "))
	code := run.Attached(output, command[0], command[1:]...)
	fmt.Fprintf(output, "%s%s %s ended (exit %d)\n", logStamp, time.Now().Format(time.DateTime), filepath.Base(command[0]), code)
	if output != os.Stderr {
		output.Close()
	}
	if fsutil.Exists(marker) || time.Since(began) >= upAfter {
		return Up
	}
	return NeverStarted
}

func MarkUp(marker, good string) (string, error) {
	kernel, _ := run.CaptureQuiet(run.Quick, "uname", "-r")
	if err := os.MkdirAll(filepath.Dir(good), ownMode); err != nil {
		return kernel, err
	}
	if err := atomic.Write(good, []byte(kernel+"\n"), logMode); err != nil {
		return kernel, err
	}
	return kernel, atomic.Write(marker, []byte(kernel+"\n"), logMode)
}

func NoteTried(home, kernel string) error {
	path := filepath.Join(home, TriedNote)
	if err := os.MkdirAll(filepath.Dir(path), ownMode); err != nil {
		return err
	}
	return atomic.Write(path, []byte(kernel+"\n"), logMode)
}

func TakeTried(home string) string {
	path := filepath.Join(home, TriedNote)
	kernel := strings.TrimSpace(fsutil.Optional(path))
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(os.Stderr, err)
	}
	return kernel
}
