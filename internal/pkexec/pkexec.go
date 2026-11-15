package pkexec

import (
	"errors"
	"fmt"
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
	uid := os.Getenv("PKEXEC_UID")
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

func Launcher(person account.Account) (string, error) {
	program, err := filepath.EvalSymlinks(filepath.Join(env.BinDir(), "solanum"))
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(program) {
		return "", errors.New(program + " isn't a full path")
	}
	if err := trusted(program, person); err != nil {
		return "", err
	}
	return program, nil
}
