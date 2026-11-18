package osinstall

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/disks"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/luks"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/words"
)

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
	if _, messages, err := run.CaptureFed(run.Minute, "", "mount", stick.Path, folder); err != nil {
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
	err = atomic.Stage(filepath.Join(folder, name), []byte(KeyFile(key, host, made)), 0o644)
	if mounted {
		if _, messages, unmounted := run.CaptureFed(run.Minute, "", "umount", folder); unmounted != nil && err == nil {
			err = failed(stick.Path+" didn't unmount, so take it out only after the restart", messages)
		}
	}
	return name, err
}
