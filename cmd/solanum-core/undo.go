package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/files"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/inhibit"
	"github.com/Ahmadlolbit/solanumos/internal/ledger"
	"github.com/Ahmadlolbit/solanumos/internal/modules"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/runlog"
	"github.com/Ahmadlolbit/solanumos/internal/service"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
)

const (
	shownNames    = 8
	backupsModule = "16"
	backupHelper  = "restic-backup"
)

var undoable = []string{"09", "10", "11", "16"}

type undoPlan struct {
	id       string
	packages []string
	services []string
	files    []string
	made     []string
}

func moduleTitle(id string) string {
	for _, module := range modules.All {
		if module.ID == id {
			return module.Title
		}
	}
	return id
}

func takenBack(record ledger.Record, installed map[string]bool, id string) []string {
	elsewhere := ledger.WantedElsewhere(id)
	wanted := record.Values(ledger.Wants)
	var names []string
	for _, name := range record.Values(ledger.Package) {
		if installed[name] && slices.Contains(wanted, name) && !elsewhere[name] {
			names = append(names, name)
		}
	}
	return names
}

func planUndo(id string) (undoPlan, error) {
	installed, err := apt.Installed()
	if err != nil {
		return undoPlan{}, fmt.Errorf("the installed packages can't be read: %w", err)
	}
	record := ledger.Load(id)
	if len(record.Values(ledger.Made)) > 0 {
		if _, err := account.People(); err != nil {
			return undoPlan{}, err
		}
	}
	record[ledger.Package] = append(record[ledger.Package], ledger.Unfinished(id, installed)...)
	shared := ledger.Elsewhere(id, ledger.File)
	plan := undoPlan{id: id, services: record.Values(ledger.Service), packages: takenBack(record, installed, id)}
	for _, path := range record.Values(ledger.File) {
		if !shared[path] {
			plan.files = append(plan.files, path)
		}
	}
	for _, path := range record.Values(ledger.Made) {
		if removable(path) && (fsutil.Exists(path) || fsutil.IsLink(path)) {
			plan.made = append(plan.made, path)
		}
	}
	return plan, nil
}

func removable(path string) bool {
	clean := filepath.Clean(path)
	if !filepath.IsAbs(clean) || strings.Count(clean, "/") < 2 {
		return false
	}
	people, err := account.People()
	if err != nil {
		return false
	}
	homes := []string{env.Home()}
	for _, person := range people {
		homes = append(homes, person.Home)
	}
	return !slices.ContainsFunc(homes, func(home string) bool {
		return strings.HasPrefix(filepath.Clean(home)+"/", clean+"/")
	})
}

func named(noun string, names []string) string {
	if len(names) == 0 {
		return ""
	}
	shown := names[:min(len(names), shownNames)]
	text := fmt.Sprintf("  %d %s: %s", len(names), noun, strings.Join(shown, " "))
	if len(names) > len(shown) {
		text += fmt.Sprintf(" and %d more", len(names)-len(shown))
	}
	return text + "\n"
}

func (plan undoPlan) describe() string {
	text := fmt.Sprintf("Undoing module %s, %s, takes back:\n", plan.id, moduleTitle(plan.id))
	text += named("packages", plan.packages) + named("services", plan.services) + named("files", plan.files) + named("unpacked", plan.made)
	return text + "Packages and files another module also uses stay, and so do packages you asked apt for yourself and files\n" +
		"you've changed since. apt also removes other packages nothing needs any more, as apt autoremove does.\n"
}

func snapshotBefore(description string) (bool, error) {
	if snapshotsOff() {
		return false, nil
	}
	if !installed("snapper") || !fsutil.IsFile(env.Or("SOLANUM_SNAPPER_CONFIG", "/etc/snapper/configs/root")) {
		ui.Print("Snapshots aren't set up, so there's no restore point.\n")
		return false, nil
	}
	ui.Print("Taking a restore point first\n")
	if run.Status("snapper", "--no-dbus", "-c", "root", "create", "--description", description, "--cleanup-algorithm", "number") != 0 {
		return false, errors.New("the restore point failed, so nothing was changed")
	}
	return true, nil
}

func (plan undoPlan) takeFiles(log *runlog.Log) error {
	deployer, err := files.New(log, account.Account{})
	if err != nil {
		return err
	}
	for _, path := range plan.files {
		outcome, err := deployer.Undeploy(path)
		if err != nil {
			return err
		}
		if outcome == "kept" {
			fmt.Fprintln(log, "kept "+path+": you changed it since")
		}
		if err := ledger.Forget(plan.id, ledger.File, path); err != nil {
			return err
		}
	}
	return nil
}

func removeMade(log *runlog.Log, path string) error {
	person, ok, err := account.Owner(path)
	if err != nil {
		return err
	}
	if ok {
		return person.Run(log, "rm", "-rf", path)
	}
	return os.RemoveAll(path)
}

func (plan undoPlan) takeMade(log *runlog.Log) error {
	for _, path := range plan.made {
		fmt.Fprintln(log, "removing "+path)
		if err := removeMade(log, path); err != nil {
			return err
		}
	}
	return nil
}

func backupRunsOff(log *runlog.Log, helper string) error {
	person, ok, err := account.Owner(helper)
	if err != nil || !ok {
		return err
	}
	if err := person.Run(log, helper, "usb", "off"); err != nil {
		return err
	}
	if !fsutil.Exists(person.Path(".config/systemd/user/restic-backup.timer")) {
		return nil
	}
	return person.Run(log, helper, "schedule", "off")
}

func (plan undoPlan) backupsOff(log *runlog.Log) error {
	if plan.id != backupsModule {
		return nil
	}
	for _, path := range plan.files {
		if filepath.Base(path) != backupHelper || !fsutil.Runnable(path) {
			continue
		}
		if err := backupRunsOff(log, path); err != nil {
			return err
		}
	}
	return nil
}

func (plan undoPlan) run(log *runlog.Log) error {
	steps := []func(*runlog.Log) error{
		plan.backupsOff,
		func(log *runlog.Log) error { return service.Disable(log, plan.services...) },
		plan.takeMade,
		plan.takeFiles,
		func(log *runlog.Log) error { return apt.Unneeded(log, plan.packages...) },
	}
	for _, step := range steps {
		if err := step(log); err != nil {
			return err
		}
	}
	if err := ledger.Drop(plan.id); err != nil {
		return err
	}
	return modules.SaveStatus(map[string]string{plan.id: modules.Undone})
}

func undoFailed(log *runlog.Log, format string, args ...any) int {
	log.Line("error", fmt.Sprintf(format, args...))
	return fail(format, args...)
}

func takeBack(id string, yes bool, log *runlog.Log) int {
	plan, err := planUndo(id)
	if err != nil {
		return undoFailed(log, "Nothing was changed: %v.", err)
	}
	description := plan.describe()
	fmt.Fprint(log, description)
	ui.Print(description)
	if !yes && !ui.Confirm("Undo module "+id+"?") {
		return undoFailed(log, "Nothing was changed.")
	}
	defer inhibit.While("Undoing module " + id)()
	taken, err := snapshotBefore("before undoing module " + id)
	if err != nil {
		return undoFailed(log, "%v", err)
	}
	log.Mirror(ui.Mirror)
	err = plan.run(log)
	log.Mirror(nil)
	if err != nil && taken {
		return undoFailed(log, "Undoing module %s stopped: %v. The restore point taken first can bring the system back: solanum restore.", id, err)
	} else if err != nil {
		return undoFailed(log, "Undoing module %s stopped: %v. Run it again once that's fixed: it picks up where it stopped.", id, err)
	}
	log.Line("info", "module "+id+" undone")
	ui.Print(fmt.Sprintf("Module %s is undone. The setup leaves it out until you run it again: sudo solanum install --only %s\n", id, id))
	return 0
}

func undo(id string, yes bool) int {
	if status, _ := modules.StatusOf(fsutil.Optional(modules.StatusFile()), id); !ledger.Known(id) && status == modules.Undone {
		return say(fmt.Sprintf("Module %s is undone already. The setup leaves it out until you run it again: sudo solanum install --only %s\n", id, id))
	}
	if !ledger.Known(id) {
		return fail("There's no record of module %s yet. Run the setup once (sudo solanum install), then undo it.", id)
	}
	log, err := runlog.Start("undo " + id)
	if err != nil {
		return fail("Nothing was changed: cannot write logs to %s.", env.LogDir())
	}
	code := takeBack(id, yes, log)
	log.Line("info", "finished with exit code "+strconv.Itoa(code))
	return code
}

type undoEntry struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	State    string   `json:"state"`
	Packages []string `json:"packages"`
	Services []string `json:"services"`
	Files    []string `json:"files"`
	Unpacked []string `json:"unpacked"`
	Problem  string   `json:"problem"`
}

func listed(names []string) []string {
	if names == nil {
		return []string{}
	}
	return names
}

func undoEntryFor(id string) undoEntry {
	entry := undoEntry{ID: id, Title: moduleTitle(id), State: "unknown", Packages: []string{}, Services: []string{}, Files: []string{}, Unpacked: []string{}}
	status, _ := modules.StatusOf(fsutil.Optional(modules.StatusFile()), id)
	switch {
	case !ledger.Known(id) && status == modules.Undone:
		entry.State = "undone"
		return entry
	case !ledger.Known(id):
		return entry
	}
	plan, err := planUndo(id)
	if err != nil {
		entry.State, entry.Problem = "unread", err.Error()
		return entry
	}
	entry.State = "ready"
	entry.Packages, entry.Services, entry.Files, entry.Unpacked = listed(plan.packages), listed(plan.services), listed(plan.files), listed(plan.made)
	return entry
}

func listUndo() int {
	entries := []undoEntry{}
	for _, id := range undoable {
		entries = append(entries, undoEntryFor(id))
	}
	return printed(entries)
}

func undoCommand() int {
	id, yes := env.Arg(2), env.Arg(3) == "--yes"
	if id == "--list" && env.Arg(3) == "--json" && env.Arg(4) == "" {
		return listUndo()
	}
	if id == "" || id == "--list" || (env.Arg(3) != "" && !yes) || env.Arg(4) != "" {
		return fail("usage: solanum undo MODULE [--yes], where MODULE is 09, 10, 11 or 16")
	}
	if !slices.Contains(undoable, id) {
		return fail("Module %s can't be undone: only 09, 10, 11 and 16 can, because the others hold up the desktop or the boot.", id)
	}
	return asRoot("undo "+id, alone(func() int { return undo(id, yes) }))
}
