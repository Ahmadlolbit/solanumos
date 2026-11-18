package password

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
)

const (
	usage        = "usage: pkexec --user YOU solanum-backup-password\n\nShows your own backup password once you've typed your password. It runs only as the person who asked.\n"
	usageCode    = 2
	passwordSize = 4096
)

func Main() {
	os.Exit(show(os.Args[1:]))
}

func caller() (account.Account, bool) {
	uid := os.Getenv("PKEXEC_UID")
	if uid == "" || uid != strconv.Itoa(os.Geteuid()) {
		return account.Account{}, false
	}
	for _, person := range account.Humans() {
		if person.UID == uid {
			return person, true
		}
	}
	return account.Account{}, false
}

func passwordFile(person account.Account) string {
	return env.Or("SOLANUM_BACKUP_PASSWORD", filepath.Join(person.Home, ".config", "restic-backup", "password"))
}

func show(args []string) int {
	if len(args) != 0 {
		fmt.Fprint(os.Stderr, usage)
		return usageCode
	}
	person, ok := caller()
	if !ok {
		fmt.Fprintln(os.Stderr, "solanum-backup-password shows only your own backup password, through pkexec --user with your own account")
		return 1
	}
	uid, _ := strconv.Atoi(person.UID)
	data, err := fsutil.ReadPlain(passwordFile(person), passwordSize, uid)
	if err != nil {
		fmt.Fprintln(os.Stderr, "There's no backup password yet: restic-backup init makes one.")
		return 1
	}
	fmt.Println(strings.TrimSpace(string(data)))
	return 0
}
