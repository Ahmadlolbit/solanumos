package osinstall

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/disks"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/luks"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/runlog"
	"github.com/Ahmadlolbit/solanumos/internal/step"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
	"github.com/Ahmadlolbit/solanumos/internal/words"
)

type handed struct {
	key string
	why string
}

func KeyReady(answers Answers) error {
	if _, err := exec.LookPath(luks.Enroller); answers.Key && err != nil {
		return errors.New("systemd-cryptenroll isn't on this live system, so the recovery key can't be made, and nothing was erased: turn Recovery key off on the disk password page, or start from a stick that has systemd-cryptsetup")
	}
	return nil
}

func (work *install) makeKey(log *runlog.Log) handed {
	key, messages, err := luks.Make(work.plan.RootDevice(), work.answers.Password)
	io.WriteString(log, messages)
	if key == "" {
		log.Line("warn", "no recovery key, and the install goes on: "+err.Error())
		return handed{why: err.Error()}
	}
	if err != nil {
		log.Line("warn", "the recovery key opens the disk, though "+luks.Enroller+" ended with an error: "+err.Error())
	}
	log.Line("info", "the recovery key is made: the installer shows it once at the end, and no log keeps it")
	return handed{key: key}
}

func (work *install) keyStep(runner *step.Runner, title string) handed {
	if runner.Ended() {
		return handed{}
	}
	runner.Log.Line("info", "step: "+title)
	ui.Busy(words.T(title))
	made := work.makeKey(runner.Log)
	ui.Done()
	if made.why == "" {
		ui.OK(words.T(title))
		return made
	}
	ui.Warn(words.T(title))
	ui.Note(made.why)
	return made
}

func keyMount() string {
	return env.Or("SOLANUM_KEY_MOUNT", "/run/solanum-key")
}

func KeySticks(target string) ([]disks.Stick, error) {
	output, err := run.CaptureQuiet(run.Quick, "lsblk", listing...)
	if err != nil {
		return nil, err
	}
	all, err := disks.Parse(output)
	return disks.Sticks(all, target), err
}

func KeyFile(key, host string, made time.Time) string {
	return words.T("Solanum recovery key") + " · " + host + " · " + made.Format("2006-01-02 15:04") + "\n\n" + key + "\n\n" +
		words.T("Type it in small letters, with its dashes, wherever the disk password is asked for.") + "\n"
}

func failed(what, messages string) error {
	if reason := luks.Reason(messages); reason != "" {
		return errors.New(what + ": " + reason)
	}
	return errors.New(what)
}

func stickFolder(stick disks.Stick) (string, bool, error) {
	if stick.Mountpoint != "" {
		return stick.Mountpoint, false, nil
	}
	folder := keyMount()
	if err := os.MkdirAll(folder, 0o700); err != nil {
		return "", false, err
	}
	if _, messages, err := run.CaptureFed(run.Minute, "", "mount", "-o", "nosuid,nodev,noexec", stick.Path, folder); err != nil {
		return "", false, failed(stick.Path+" didn't mount", messages)
	}
	return folder, true, nil
}

func SaveKey(stick disks.Stick, key, host string) (string, error) {
	folder, mounted, err := stickFolder(stick)
	if err != nil {
		return "", err
	}
	made := time.Now()
	name := "solanum-recovery-key-" + host + "-" + made.Format("20060102-150405") + ".txt"
	path := filepath.Join(folder, name)
	err = atomic.Stage(path, []byte(KeyFile(key, host, made)), 0o644)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		os.Remove(path)
	}
	if mounted {
		if _, messages, unmounted := run.CaptureFed(run.Minute, "", "umount", folder); unmounted != nil && err == nil {
			err = failed(stick.Path+" didn't unmount, so take it out only after the restart", messages)
		}
	}
	return name, err
}
