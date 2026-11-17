package task

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/pkexec"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	usage = `usage: pkexec solanum-task-helper JOB [ARGUMENT]

Runs one of Solanum's tasks as root for the person who asked, with its events
on stdout and its answers on stdin. The jobs:
  update [system|apps|solanum]
  doctor
  cleanup
  restore NUMBER
  undo 09|10|11|16
  app-install APP
  app-remove APP
  pkg-install NAME...
  pkg-remove NAME...
`
	usageCode   = 2
	mostNames   = 100
	groupWrite  = 0o020
	othersWrite = 0o002
	rootAccount = "0"
	safePath    = "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
)

type job struct {
	command  []string
	argument *regexp.Regexp
	needed   bool
	many     bool
}

var (
	appID  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)
	pkgID  = regexp.MustCompile(`^[a-z0-9][a-z0-9+.-]{0,99}(:[a-z0-9-]{1,20})?$`)
	locale = regexp.MustCompile(`^(LANG|LANGUAGE|LC_[A-Z]+)=[A-Za-z0-9_.@:-]{0,64}$`)
	jobs   = map[string]job{
		"update":      {[]string{"update", "--events"}, regexp.MustCompile(`^(system|apps|solanum)$`), false, false},
		"doctor":      {[]string{"doctor", "--events"}, nil, false, false},
		"cleanup":     {[]string{"cleanup", "--events"}, nil, false, false},
		"restore":     {[]string{"restore", "--events"}, regexp.MustCompile(`^[1-9][0-9]{0,8}$`), true, false},
		"undo":        {[]string{"undo", "--events"}, regexp.MustCompile(`^(09|10|11|16)$`), true, false},
		"app-install": {[]string{"app", "install", "--events"}, appID, true, false},
		"app-remove":  {[]string{"app", "remove", "--events"}, appID, true, false},
		"pkg-install": {[]string{"pkg", "--events", "install"}, pkgID, true, true},
		"pkg-remove":  {[]string{"pkg", "--events", "remove"}, pkgID, true, true},
	}
	start = func(program string, args, environment []string) int {
		os.Clearenv()
		for _, entry := range environment {
			name, value, _ := strings.Cut(entry, "=")
			os.Setenv(name, value)
		}
		return run.Report(run.Exec(program, args...))
	}
)

func Main() {
	os.Exit(handle(os.Args[1:]))
}

func command(args []string) ([]string, bool) {
	if len(args) == 0 {
		return nil, false
	}
	wanted, known := jobs[args[0]]
	given := args[1:]
	switch {
	case !known || len(given) > 1 && (!wanted.many || len(given) > mostNames):
		return nil, false
	case len(given) == 0:
		return wanted.command, !wanted.needed
	case wanted.argument == nil || slices.ContainsFunc(given, func(each string) bool { return !wanted.argument.MatchString(each) }):
		return nil, false
	}
	return append(slices.Clone(wanted.command), given...), true
}

func caller() (account.Account, error) {
	uid := os.Getenv("PKEXEC_UID")
	for _, person := range account.Humans() {
		if person.UID == uid {
			return person, nil
		}
	}
	return account.Account{}, errors.New("solanum-task-helper runs only through pkexec, for a person's own account")
}

func safeFor(stat syscall.Stat_t, person account.Account) bool {
	owner, group := strconv.FormatUint(uint64(stat.Uid), 10), strconv.FormatUint(uint64(stat.Gid), 10)
	if (owner != rootAccount && owner != person.UID) || stat.Mode&othersWrite != 0 {
		return false
	}
	return stat.Mode&groupWrite == 0 || group == rootAccount || group == person.GID
}

func trusted(path string, person account.Account) error {
	for ; ; path = filepath.Dir(path) {
		var stat syscall.Stat_t
		if err := syscall.Lstat(path, &stat); err != nil {
			return err
		}
		if !safeFor(stat, person) {
			return fmt.Errorf("%s can be changed by someone other than %s and root, so Solanum doesn't run from it as root", path, person.Name)
		}
		if path == filepath.Dir(path) {
			return nil
		}
	}
}

func launcher(person account.Account) (string, error) {
	program, err := pkexec.Recorded()
	if err != nil {
		return "", err
	}
	return program, trusted(program, person)
}

func environment(person account.Account) []string {
	kept := []string{safePath, "HOME=/root", "USER=root", "LOGNAME=root", "PKEXEC_UID=" + person.UID, "SOLANUM_USER=" + person.Name}
	for _, entry := range os.Environ() {
		if locale.MatchString(entry) {
			kept = append(kept, entry)
		}
	}
	return kept
}

func handle(args []string) int {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Print(usage)
		return 0
	}
	words, ok := command(args)
	if !ok {
		fmt.Fprint(os.Stderr, usage)
		return usageCode
	}
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "solanum-task-helper runs as root: a window asks for your password through pkexec")
		return 1
	}
	person, err := caller()
	if err != nil {
		return refused(err)
	}
	program, err := launcher(person)
	if err != nil {
		return refused(err)
	}
	return start(program, words, environment(person))
}

func refused(err error) int {
	fmt.Fprintln(os.Stderr, err)
	return 1
}
