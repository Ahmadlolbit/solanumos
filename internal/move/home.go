package move

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/oldname"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const textLimit = 1 << 20

var (
	transient   = []string{"changes.txt", "changes-unseen", "wizard.conf"}
	ownFolders  = []string{".config", ".local/state"}
	personFiles = []string{".config/hypr/hyprland.lua", ".zshrc", ".config/mako/config", ".config/kitty/kitty.conf", ".config/hypr/hyprlock.conf", ".config/hypr/hyprpaper.conf", ".config/solanum/menu.json", ".config/solanum/look.conf"}
	ownRenames  = [][2]string{{".config/kdeglobals" + oldname.KDEBackup, ".config/kdeglobals.before-solanum"}, {".local/zed.app/" + oldname.VersionStamp, ".local/zed.app/.solanum-version"}}
)

func (work *mover) people() []account.Account {
	people, err := account.People()
	if err != nil {
		fmt.Fprintln(work.out, err)
	}
	return people
}

func (work *mover) ownerOf(path string) (account.Account, bool) {
	for _, person := range work.people() {
		if home := filepath.Clean(person.Home); home != "/" && strings.HasPrefix(path, home+"/") {
			return person, true
		}
	}
	return account.Account{}, false
}

func (work *mover) asPerson(person account.Account, args ...string) bool {
	if err := person.Run(work.out, args...); err != nil {
		fmt.Fprintf(work.out, "as %s, %s: %v\n", person.Name, strings.Join(args, " "), err)
		return false
	}
	return true
}

func (work *mover) moveHomes() {
	for _, person := range work.people() {
		if fsutil.IsDir(at(person.Home)) {
			work.moveHome(person)
		}
	}
}

func (work *mover) moveHome(person account.Account) {
	for _, rel := range ownFolders {
		work.moveOwnFolder(person, at(filepath.Join(person.Home, rel, oldname.Name)), at(filepath.Join(person.Home, rel, "solanum")))
	}
	if cache := at(filepath.Join(person.Home, ".cache", oldname.Name)); fsutil.Exists(cache) {
		work.asPerson(person, "rm", "-rf", cache)
	}
	work.rewriteFiles(person)
	work.relink(person)
	work.renameOwn(person)
}

func (work *mover) moveOwnFolder(person account.Account, old, fresh string) {
	if fsutil.IsLink(old) || !fsutil.IsDir(old) {
		return
	}
	if target, err := os.Readlink(fresh); err == nil && filepath.Base(target) == filepath.Base(old) {
		work.asPerson(person, "rm", "-f", fresh)
	}
	if !fsutil.Exists(fresh) && !work.asPerson(person, "mv", "-T", "--update=none-fail", old, fresh) {
		work.note("Left %s where it is: it couldn't be moved to %s", old, fresh)
		return
	}
	if fsutil.IsDir(old) && !work.mergeOwn(person, old, fresh) {
		return
	}
	work.asPerson(person, "ln", "-s", filepath.Base(fresh), old)
}

func (work *mover) mergeOwn(person account.Account, old, fresh string) bool {
	var left []string
	for _, entry := range fsutil.Entries(old) {
		from, to := filepath.Join(old, entry.Name()), filepath.Join(fresh, entry.Name())
		switch {
		case !fsutil.Exists(to) && !fsutil.IsLink(to):
			if !work.asPerson(person, "mv", "-T", "--update=none-fail", from, to) {
				left = append(left, entry.Name())
			}
		case slices.Contains(transient, entry.Name()):
			work.asPerson(person, "rm", "-f", from)
		default:
			left = append(left, entry.Name())
		}
	}
	if len(left) > 0 {
		work.note("Left %s where it is: %s already has %s", old, fresh, strings.Join(left, ", "))
		return false
	}
	return work.asPerson(person, "rmdir", old)
}

func (work *mover) relink(person account.Account) {
	roots := []string{at(filepath.Join(person.Home, ".config")), at(filepath.Join(person.Home, ".local/share")), at(filepath.Join(person.Home, ".local/state/solanum"))}
	output, _ := run.CaptureQuiet(run.Minute, "find", append(roots, "-maxdepth", "5", "-type", "l", "(", "-lname", "*/.local/state/"+oldname.Name+"/*", "-o", "-lname", "*/.config/"+oldname.Name+"/*", ")")...)
	for _, link := range run.Lines(output) {
		target, err := os.Readlink(link)
		if err != nil {
			continue
		}
		fresh := strings.NewReplacer("/.local/state/"+oldname.Name+"/", "/.local/state/solanum/", "/.config/"+oldname.Name+"/", "/.config/solanum/").Replace(target)
		work.asPerson(person, "ln", "-sfn", fresh, link)
	}
}

func (work *mover) renameOwn(person account.Account) {
	for _, pair := range ownRenames {
		work.moveOwn(person, at(filepath.Join(person.Home, pair[0])), at(filepath.Join(person.Home, pair[1])))
	}
	units := at(filepath.Join(person.Home, ".config/systemd/user"))
	for _, entry := range fsutil.Entries(units) {
		if strings.HasPrefix(entry.Name(), oldname.Prefix) {
			work.moveOwn(person, filepath.Join(units, entry.Name()), filepath.Join(units, oldname.Swap(entry.Name())))
		}
	}
	for _, rel := range personFiles {
		for _, pair := range suffixes {
			path := at(filepath.Join(person.Home, rel))
			work.moveOwn(person, path+pair[0], path+pair[1])
		}
	}
}

func (work *mover) moveOwn(person account.Account, old, fresh string) {
	switch {
	case !fsutil.Exists(old) && !fsutil.IsLink(old):
	case fsutil.Exists(fresh) || fsutil.IsLink(fresh):
		work.note("Left %s where it is: %s is there too", old, fresh)
	default:
		work.asPerson(person, "mv", "-T", "--update=none-fail", old, fresh)
	}
}
