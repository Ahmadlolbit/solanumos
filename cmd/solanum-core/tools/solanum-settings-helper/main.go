package helper

import (
	"fmt"
	"os"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/config"
	"github.com/Ahmadlolbit/solanumos/internal/pkexec"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/runlog"
	"github.com/Ahmadlolbit/solanumos/internal/settings"
)

const usage = "usage: pkexec solanum-settings-helper NAME VALUE"

var install = func(modules []string) int {
	person, _ := pkexec.Caller()
	program, err := pkexec.Launcher(person)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return run.Status(program, "install", "--only", strings.Join(modules, ","), "--yes")
}

func Main() {
	os.Exit(change(os.Args[1:]))
}

func checked(args []string) (string, string, string, int) {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, usage)
		return "", "", "", 2
	}
	option, found := settings.Option(args[0])
	switch {
	case !found:
		fmt.Fprintf(os.Stderr, "Solanum has no setting for the whole computer named %s\n", args[0])
		return "", "", "", 2
	case !option.Valid(args[1]):
		fmt.Fprintln(os.Stderr, option.Refusal(args[1]))
		return "", "", "", 2
	case os.Geteuid() != 0:
		fmt.Fprintln(os.Stderr, "This changes the computer for everyone, so it runs as root: Settings asks for your password through pkexec")
		return "", "", "", 1
	}
	return args[0], args[1], option.Label, 0
}

func change(args []string) int {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println(usage)
		return 0
	}
	name, value, label, code := checked(args)
	if code != 0 {
		return code
	}
	log, err := runlog.Start("Settings changed " + name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	current := config.Read(config.Path())
	before := current.Get(name)
	current.Set(name, value)
	if err := current.Save(config.Path()); err != nil {
		log.Line("error", err.Error())
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	log.Line("info", fmt.Sprintf("%s: %q -> %q", name, before, value))
	fmt.Printf("%s is %s now.\n", label, value)
	modules := settings.Modules(name)
	if len(modules) == 0 {
		fmt.Println("It takes effect the next time the setup runs.")
		return 0
	}
	fmt.Printf("Running module %s to apply it.\n", strings.Join(modules, " and "))
	code = install(modules)
	log.Line("info", fmt.Sprintf("module %s ended with %d", strings.Join(modules, ","), code))
	return code
}
