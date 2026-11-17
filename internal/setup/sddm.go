package setup

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/files"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/ledger"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/runlog"
	"github.com/Ahmadlolbit/solanumos/internal/service"
)

const (
	sddmTheme    = "solanum"
	oldSddmTheme = "catppuccin-mocha-blue"
	sddmUnit     = "sddm.service"
	sddmProgram  = "/usr/bin/sddm"
	sessionFile  = "/usr/share/wayland-sessions/hyprland.desktop"
	debianHypr   = "hyprland.desktop"
	ourEntry     = "solanum.desktop"
	sddmChoice   = "sddm shared/default-x-display-manager select sddm\n"
)

func displayFile() string {
	return env.Or("SOLANUM_DISPLAY_FILE", "/etc/X11/default-display-manager")
}

func displayAlias() string {
	return env.Or("SOLANUM_DISPLAY_ALIAS", "/etc/systemd/system/display-manager.service")
}

func sddmConf() string {
	return env.Or("SOLANUM_SDDM_CONF", "/etc/sddm.conf.d/10-solanum.conf")
}

func sddmState() string {
	return env.Or("SOLANUM_SDDM_STATE", "/var/lib/sddm/state.conf")
}

func sessionsDir() string {
	return env.Or("SOLANUM_SESSIONS", "/usr/share/wayland-sessions")
}

func sddmThemes() string {
	return env.Or("SOLANUM_SDDM_THEMES", "/usr/share/sddm/themes")
}

func (context *Context) sddmPackages(log *runlog.Log) error {
	if err := apt.InstallLists(log, "sddm"); err != nil {
		return err
	}
	if context.Config.Get("GREETER") == "wayland" {
		return apt.InstallLists(log, "sddm-wayland")
	}
	return nil
}

func (context *Context) sddmFirstState() error {
	template, err := os.ReadFile(filepath.Join(env.Root(), "system/sddm/state.conf"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(sddmState()), programMode); err != nil {
		return err
	}
	uid, gid := sddmOwner()
	text := files.Render(string(template), "USER="+context.Person.Name, "SESSIONS="+sessionsDir())
	return atomic.InstallAs(sddmState(), []byte(text), plainMode, uid, gid)
}

func sddmOwner() (int, int) {
	sddm, found := account.Lookup("sddm")
	if !found {
		return 0, 0
	}
	uid, gid, err := sddm.IDs()
	if err != nil {
		return 0, 0
	}
	return uid, gid
}

func (context *Context) sddmConfigure(*runlog.Log) error {
	source := "system/sddm/" + context.Config.Get("GREETER") + ".conf"
	if err := context.Deploy.File(source, sddmConf(), "THEME="+sddmTheme); err != nil {
		return err
	}
	if context.Person.Name != "" && !fileLineHasPrefix(sddmState(), "Session=") {
		return context.sddmFirstState()
	}
	return nil
}

func stateOnOurs(text string) (string, bool) {
	debian, ours := "Session="+filepath.Join(sessionsDir(), debianHypr), "Session="+filepath.Join(sessionsDir(), ourEntry)
	lines := strings.Split(text, "\n")
	at := slices.Index(lines, debian)
	if at < 0 {
		return text, false
	}
	lines[at] = ours
	return strings.Join(lines, "\n"), true
}

func (context *Context) sessionOffer(log *runlog.Log) error {
	if err := context.Deploy.File("system/sddm/"+ourEntry, filepath.Join(sessionsDir(), ourEntry), "BIN="+env.BinDir()); err != nil {
		return err
	}
	text, err := fsutil.Text(sddmState())
	switched, changed := stateOnOurs(text)
	if err != nil || !changed {
		return nil
	}
	log.Line("info", "the login screen starts the Solanum session from now on")
	return atomic.Write(sddmState(), []byte(switched+"\n"), plainMode)
}

func displayManager() string {
	target, err := os.Readlink(displayAlias())
	if err != nil {
		return ""
	}
	return filepath.Base(target)
}

func takeDisplayManager(log *runlog.Log) error {
	if service.Enabled(sddmUnit) {
		return nil
	}
	if old := displayManager(); old != "" && old != sddmUnit {
		ledger.Note(ledger.Replaced, old)
		log.Line("info", "the login screen was "+old+", kept in module 07's record")
	}
	if err := run.Supply(run.Minute, log, sddmChoice, "debconf-set-selections"); err != nil {
		return err
	}
	return service.Replace(log, sddmUnit)
}

func sddmEnable(log *runlog.Log) error {
	if err := takeDisplayManager(log); err != nil {
		return err
	}
	if text, _ := fsutil.Text(displayFile()); text != sddmProgram {
		if err := atomic.Write(displayFile(), []byte(sddmProgram+"\n"), plainMode); err != nil {
			return err
		}
	}
	current, _ := run.CaptureTo(run.Minute, log, "systemctl", "get-default")
	if current == "graphical.target" {
		return nil
	}
	return run.Logged(log, "systemctl", "set-default", "graphical.target")
}

func (context *Context) sddmThemeInstall(log *runlog.Log) error {
	if err := apt.InstallLists(log, "sddm-theme"); err != nil {
		return err
	}
	if err := context.Deploy.Tree("system/sddm/solanum", filepath.Join(sddmThemes(), sddmTheme)); err != nil {
		return err
	}
	mark := filepath.Join(env.StateDir(), "releases", "catppuccin-sddm")
	if !fsutil.Exists(mark) {
		return nil
	}
	if err := context.Deploy.RemoveTree(filepath.Join(sddmThemes(), oldSddmTheme)); err != nil {
		return err
	}
	return os.RemoveAll(mark)
}

func sddmThemeReady(*runlog.Log) error {
	if err := fileThere(filepath.Join(sddmThemes(), sddmTheme, "Main.qml")); err != nil {
		return err
	}
	return expect(fileLineHasPrefix(sddmConf(), "Current="+sddmTheme), sddmConf()+" names another theme")
}

func sddmSteps(context *Context) {
	context.Run.Step("Install the login screen", context.sddmPackages)
	context.Run.Step("Install the Solanum login screen", context.sddmThemeInstall)
	context.Run.Step("Configure the login screen", context.sddmConfigure)
	context.Run.Step("Offer the Solanum session", context.sessionOffer)
	context.Run.Step("Start into the login screen at boot", sddmEnable)
}

func sddmStarts(*runlog.Log) error {
	if !service.Enabled(sddmUnit) {
		return errors.New(sddmUnit + " isn't enabled")
	}
	text, _ := fsutil.Text(displayFile())
	return expect(text == sddmProgram, displayFile()+" names another login screen")
}

func sddmChecks(context *Context) {
	context.Run.Check("The login screen starts at boot", sddmStarts)
	context.Run.Check("Hyprland is offered as a session", filesCheck(sessionFile))
	context.Run.Check("The Solanum session is offered", filesCheck(filepath.Join(sessionsDir(), ourEntry)))
	context.Run.Check("The login screen uses the Solanum theme", sddmThemeReady)
}
