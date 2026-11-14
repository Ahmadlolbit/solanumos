package lock

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/oldname"
)

const (
	fileMode   = 0o600
	lockFields = 6
)

var heldLocks = "/proc/locks"

type Held struct {
	files []*os.File
}

func Path() string {
	return env.Or("SOLANUM_LOCK", "/run/solanum.lock")
}

func oldPath() string {
	if os.Getenv("SOLANUM_LOCK") != "" {
		return os.Getenv("SOLANUM_OLD_LOCK")
	}
	return oldname.Lock
}

func open(path string, how int) (*os.File, error) {
	file, err := os.OpenFile(path, how, fileMode)
	if err != nil {
		return nil, err
	}
	if err := file.Chmod(fileMode); err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

func (held *Held) Close() error {
	var err error
	for _, file := range held.files {
		err = errors.Join(err, file.Close())
	}
	return err
}

func (held *Held) hold(file *os.File) error {
	held.files = append(held.files, file)
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		held.Close()
		return errors.New("another Solanum run is still going")
	}
	return nil
}

func Take() (*Held, error) {
	path := Path()
	file, err := open(path, os.O_CREATE|os.O_WRONLY)
	if err != nil {
		return nil, errors.New("cannot create " + path)
	}
	held := &Held{}
	if err := held.hold(file); err != nil {
		return nil, err
	}
	if old := oldPath(); old != "" {
		if file, err := open(old, os.O_WRONLY); err == nil {
			if err := held.hold(file); err != nil {
				return nil, err
			}
		}
	}
	return held, nil
}

func identity(info os.FileInfo) string {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return ""
	}
	device := uint64(stat.Dev)
	major := (device&0x00000000000fff00)>>8 | (device&0xfffff00000000000)>>32
	minor := device&0x00000000000000ff | (device&0x00000ffffff00000)>>12
	return fmt.Sprintf("%02x:%02x:%d", major, minor, stat.Ino)
}

func written(locks, id string) bool {
	for _, line := range strings.Split(locks, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= lockFields && fields[1] == "FLOCK" && fields[3] == "WRITE" && fields[5] == id {
			return true
		}
	}
	return false
}

func busy(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	locks, err := os.ReadFile(heldLocks)
	return err == nil && written(string(locks), identity(info))
}

func Busy() bool {
	old := oldPath()
	return busy(Path()) || old != "" && busy(old)
}
