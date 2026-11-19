package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/cleanup"
	"github.com/Ahmadlolbit/solanumos/internal/config"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/events"
	"github.com/Ahmadlolbit/solanumos/internal/lock"
	"github.com/Ahmadlolbit/solanumos/internal/move"
	"github.com/Ahmadlolbit/solanumos/internal/pkexec"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
)

const usage = `solanum %s - Hyprland on Debian, set up properly

USAGE
  solanum                          Open the setup
  solanum install [options]        Run the setup with the saved configuration
                                     --config FILE      use this configuration file
                                     --only 04,05       run only these modules
                                     --skip 10,11       leave these modules out
                                     --yes              ask nothing, use the configuration as is
                                     --on-failure MODE  ask | halt | continue
                                     --changes FILE     show these update notes in the summary
  solanum verify [--only ...]      Run every check again
  solanum modules                  List the modules and how they last ended
  solanum undo MODULE [--yes]      Take back what module 09, 10, 11 or 16 installed, wrote and turned on
                                     (run it with sudo); the setup then leaves it out until you run it again
  solanum config [show|path|edit|reset|export]
                                   export prints your settings as a preset to share, without your Debian mirror
  solanum config import FILE|URL
                                   Take a preset's settings from a file or an https address (run it with sudo);
                                     your Debian mirror stays, and it lists what changed
  solanum collect [--verbose]      Pack the logs, the report and system details for a bug report;
                                     --verbose also writes everything to ~/solanum-log-<time>.txt
  solanum verbose [on|off]         Log the desktop in detail from the next login (run it without sudo)
  solanum guide                    Read the guide in the terminal
  solanum keyboard [LAYOUT]        Show the keyboard layout, list them (list), or set one (run it with sudo)
  solanum kernel [keep [VERSION]|newest|once VERSION]
                                   Show the kernels; keep the running one (or VERSION) as the one the computer
                                     starts, start the newest again, or start VERSION at the next restart only
                                     (run it with sudo to change it)
  solanum battery-limit [N|off]    Show or set where charging stops, 50 to 100 (run it with sudo to set it)
  solanum trust [on|off]           Show or set whether you trust the network you're on (run it without sudo);
                                     LocalSend and KDE Connect can reach this computer only on networks you trust
  solanum update [PART]            Update everything (run it without sudo): a snapshot, Debian, the apps,
                                     then Solanum and its installer; PART is system, apps or solanum
  solanum update --list --json     What's waiting to update, as JSON; changes nothing
  solanum changes [--json]         Show what the last solanum update changed
  solanum restore [NUMBER]         List the restore points, or show what changed since one and go back to it
                                     (run it without sudo)
  solanum doctor                   Check the system and offer the safe fixes (run it with sudo)
  solanum cleanup [--yes]          Delete the oldest automatic snapshots until / is under 85%% (run it with sudo)
  solanum software [PAGE]          Open Software to add and remove apps, Debian packages and web apps
  solanum app list                 List the apps you can install, by category (* = installed)
  solanum app install|remove APP
                                   Install or remove one app and tick or untick it in the configuration
  solanum db [list]                List the databases you can start, and which are running
  solanum db start|stop|remove NAME
                                   Postgres, MySQL or Redis in a container on localhost (run it without sudo);
                                     stop keeps the data, remove deletes it
  solanum dotfiles [REPOSITORY]    Apply your dotfiles with chezmoi, which commits what you change there;
                                     chezmoi git push sends it (run it without sudo; chezmoi runs that
                                     repository's scripts). Without one, show what chezmoi would change
  solanum fingerprint add|list|remove
                                   Add, list or remove your fingerprints (run it without sudo), once
                                     Fingerprint is on in the setup
  solanum tpm on|off|tidy|status
                                   Unlock the disk at boot with the TPM and a short PIN, the disk password kept
                                     (run it with sudo, in a terminal); moves the boot image to dracut, and tidy
                                     takes away the boot entry kept from before
  solanum recovery-key [status|new|remove]
                                   Say whether the disk has a recovery key, make a new one that replaces the
                                     old, or take it away (run it with sudo, in a terminal); new asks for the
                                     disk password and shows the key once
  solanum webcam [relay]           List the IPU6 webcam's cameras, or relay one to /dev/video42 for apps
                                     that take only a normal webcam, until Ctrl+C (run it without sudo)
  solanum care [PAGE]              Open Care, at updates, check or changes: what's waiting to update,
                                     the system check and what the last update changed
  solanum settings [PAGE]          Open Settings, at a page such as themes or display
  solanum settings list            Every setting as JSON: its page, kind, value, choices, and whether it's
                                     yours or the whole computer's
  solanum settings get|set NAME [VALUE]
                                   Show or change one; the computer's settings ask for your password
  solanum credits                  Who made Solanum, and the projects it builds on with their licences
  solanum version
  solanum help

Configuration: %s
Logs:          %s/ and ~/runbook-logs/
`

var commands = map[string]func() int{
	"version":       nothingAfter("version", showVersion),
	"--version":     nothingAfter("--version", showVersion),
	"help":          nothingAfter("help", showHelp),
	"-h":            nothingAfter("-h", showHelp),
	"--help":        nothingAfter("--help", showHelp),
	"verbose":       verbose,
	"changes":       changes,
	"update":        updateCommand,
	"config":        configCommand,
	"modules":       nothingAfter("modules", modulesCommand),
	"collect":       collect,
	"doctor":        nothingAfter("doctor", doctor),
	"app":           appCommand,
	"pkg":           pkgCommand,
	"db":            dbCommand,
	"dotfiles":      dotfilesCommand,
	"restore":       restore,
	"cleanup":       cleanupCommand,
	"install":       installCommand,
	"verify":        verifyCommand,
	"account":       accountCommand,
	"guide":         nothingAfter("guide", guide),
	"keyboard":      keyboardCommand,
	"kernel":        kernelCommand,
	"battery-limit": batteryLimit,
	"trust":         trustCommand,
	"module":        module,
	"undo":          undoCommand,
	"credits":       nothingAfter("credits", creditsCommand),
	"settings":      settingsCommand,
	"care":          careCommand,
	"software":      softwareCommand,
	"fingerprint":   fingerprintCommand,
	"tpm":           tpmCommand,
	"recovery-key":  recoveryKeyCommand,
	"install-os":    installOSCommand,
	"first-start":   firstStartCommand,
	"webcam":        webcamCommand,
}

func fail(format string, args ...any) int {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	return 1
}

func asRoot(mode string, command func() int) int {
	if os.Geteuid() != 0 {
		return fail("Solanum needs root for this. Run it again with sudo: sudo %s %s", os.Args[0], mode)
	}
	defer cleanup.IgnoreQuit()()
	return command()
}

func onBehalf() (account.Account, bool) {
	caller := os.Getenv("PKEXEC_UID")
	if os.Geteuid() != 0 || caller == "" {
		return account.Account{}, false
	}
	person, found := account.Target()
	human := slices.ContainsFunc(account.Humans(), func(each account.Account) bool { return each.UID == caller })
	return person, found && human && person.UID == caller
}

func forAPerson() bool {
	_, helped := onBehalf()
	return os.Geteuid() != 0 || helped
}

func nothingAfter(name string, command func() int) func() int {
	return func() int {
		if env.Arg(2) != "" {
			return fail("usage: solanum %s", name)
		}
		return command()
	}
}

func alone(command func() int) func() int {
	return func() int {
		held, err := lock.Take()
		if err != nil {
			return fail("%v", err)
		}
		defer held.Close()
		return command()
	}
}

func say(text string) int {
	ui.Print(text)
	return ui.ExitCode()
}

func remove(path string) int {
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func showVersion() int {
	return say("solanum " + env.Version + "\n")
}

func showHelp() int {
	return say(fmt.Sprintf(usage, env.Version, config.Path(), env.LogDir()))
}

func root() string {
	return os.Getenv("SOLANUM_ROOT")
}

func script() string {
	return filepath.Join(root(), "solanum")
}

func safeCopy() error {
	if os.Geteuid() != 0 {
		return nil
	}
	return pkexec.Clone(root(), pkexec.Person(script()))
}

func enter() int {
	program, err := pkexec.Entry()
	if err != nil {
		return fail("%v", err)
	}
	os.Setenv("SOLANUM_COMMAND", os.Args[0])
	return run.Report(run.Exec(program, os.Args[1:]...))
}

func throughScript() int {
	return run.Report(run.Exec("bash", append([]string{script()}, os.Args[1:]...)...))
}

var (
	eventCommands = []string{"update", "doctor", "cleanup", "restore", "undo", "pkg"}
	streamed      bool
)

func takeEvents() {
	if !slices.Contains(eventCommands, env.Arg(1)) || env.Arg(2) != "--events" {
		return
	}
	os.Args = slices.Delete(os.Args, 2, 3)
	streamed = true
	cleanup.OwnGroup()
	ui.Attach(events.Open(os.Stdin, os.Stdout, cleanup.StopGroup))
	run.Mirror(ui.Mirror)
}

var beforeTheMove = []string{"install", "update", "app", "changes", "collect", "restore", "guide", "credits", "help", "-h", "--help", "version", "--version", "first-start", "install-os"}

func command(name string) (func() int, bool) {
	found, ok := commands[name]
	if ok && !slices.Contains(beforeTheMove, name) && move.Pending() {
		return func() int {
			if move.Unfinished() {
				return fail("The move to Solanum isn't finished. Finish it first: sudo solanum install")
			}
			return fail("This computer still has Hyprtrixie's folders. Move it over to Solanum first: sudo solanum install")
		}, true
	}
	return found, ok
}

func main() {
	runTool()
	if env.Arg(1) == "trust" && env.Arg(2) == "apply" {
		os.Exit(asRoot("trust apply", trustApply))
	}
	if root() == "" {
		os.Exit(run.Report(run.Exec("solanum", os.Args[1:]...)))
	}
	if err := safeCopy(); err != nil {
		os.Exit(fail("%v", err))
	}
	if command, ok := command(env.Arg(1)); ok {
		takeEvents()
		os.Exit(command())
	}
	os.Exit(fail("unknown command: %s (see: solanum help)", env.Arg(1)))
}
