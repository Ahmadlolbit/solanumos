package pkexec

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/env"
)

const (
	groupWrite  = 0o020
	othersWrite = 0o002
	rootAccount = "0"
)

func Caller() (account.Account, bool) {
	return human(os.Getenv("PKEXEC_UID"))
}

func SudoCaller() (account.Account, bool) {
	return human(os.Getenv("SUDO_UID"))
}

func human(uid string) (account.Account, bool) {
	for _, person := range account.Humans() {
		if person.UID == uid {
			return person, true
		}
	}
	return account.Account{}, false
}

func safeFor(stat syscall.Stat_t, person account.Account) bool {
	owner, group := strconv.FormatUint(uint64(stat.Uid), 10), strconv.FormatUint(uint64(stat.Gid), 10)
	if (owner != rootAccount && owner != person.UID) || stat.Mode&othersWrite != 0 {
		return false
	}
	return stat.Mode&groupWrite == 0 || group == rootAccount || group == person.GID
}

func allowed(person account.Account) string {
	if person.Name == "" {
		return "root"
	}
	return person.Name + " and root"
}

func trusted(path string, person account.Account) error {
	for ; ; path = filepath.Dir(path) {
		var stat syscall.Stat_t
		if err := syscall.Lstat(path, &stat); err != nil {
			return err
		}
		if !safeFor(stat, person) {
			return fmt.Errorf("%s can be changed by someone other than %s, so Solanum doesn't run from it as root", path, allowed(person))
		}
		if path == filepath.Dir(path) {
			return nil
		}
	}
}

func Clone(root string, person account.Account) error {
	return trusted(filepath.Join(root, "solanum"), person)
}

func Record() string {
	return filepath.Join(env.StateDir(), "launcher")
}

func resolved(link string) (string, error) {
	program, err := filepath.EvalSymlinks(link)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(program) {
		return "", errors.New(program + " isn't a full path")
	}
	return program, nil
}

func Recorded() (string, error) {
	if _, err := os.Lstat(Record()); errors.Is(err, fs.ErrNotExist) {
		return resolved(filepath.Join(env.BinDir(), "solanum"))
	}
	return resolved(Record())
}

func Launcher(person account.Account) (string, error) {
	program, err := Recorded()
	if err != nil {
		return "", err
	}
	if err := trusted(program, person); err != nil {
		return "", err
	}
	return program, nil
}

func Entry() (string, error) {
	program, err := resolved(Record())
	if err != nil {
		return "", fmt.Errorf("solanum doesn't know which folder to run from (%w). In your Solanum folder, run: sudo ./solanum install", err)
	}
	if os.Geteuid() != 0 || os.Getenv("SUDO_UID") == "" {
		return program, nil
	}
	person, _ := SudoCaller()
	if err := trusted(program, person); err != nil {
		return "", err
	}
	return program, nil
}
