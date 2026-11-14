package move

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/oldname"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

func personTextBack(text string) string {
	return replacer(swaps(), true).Replace(text)
}

func menuTextBack(text string) string {
	return replacer(commandSwaps(), true).Replace(text)
}

func Before(snapshot string) bool {
	number, err := strconv.Atoi(snapshot)
	moved, movedErr := strconv.Atoi(Snapshot())
	return err == nil && movedErr == nil && number <= moved
}

func HomeApart() bool {
	var home, root syscall.Stat_t
	if syscall.Stat(at("/home"), &home) != nil || syscall.Stat(at("/"), &root) != nil {
		return false
	}
	return home.Dev != root.Dev
}

func Back(out io.Writer) []string {
	work := &mover{out: out}
	for _, person := range work.people() {
		if fsutil.IsDir(at(person.Home)) {
			work.backHome(person)
		}
	}
	return work.notes
}

func (work *mover) backHome(person account.Account) {
	for _, rel := range ownFolders {
		work.moveOwnFolder(person, at(filepath.Join(person.Home, rel, "solanum")), at(filepath.Join(person.Home, rel, oldname.Name)))
	}
	for _, rel := range personFiles {
		change := personTextBack
		if filepath.Base(rel) == "menu.json" {
			change = menuTextBack
		}
		work.rewrite(person, at(filepath.Join(person.Home, oldname.Back(rel))), change)
	}
	apps, _ := filepath.Glob(at(filepath.Join(person.Home, ".local/share/applications/webapp-*.desktop")))
	for _, app := range apps {
		work.rewrite(person, app, personTextBack)
	}
	plugins := at(filepath.Join(person.Home, ".config/nvim/lua/plugins"))
	if fresh := filepath.Join(plugins, "solanum.lua"); fsutil.IsFile(fresh) {
		work.rewrite(person, fresh, personTextBack)
		work.moveOwn(person, fresh, filepath.Join(plugins, oldname.Name+".lua"))
	}
	work.relinkBack(person)
	work.renameOwnBack(person)
}

func (work *mover) relinkBack(person account.Account) {
	roots := []string{at(filepath.Join(person.Home, ".config")), at(filepath.Join(person.Home, ".local/share")), at(filepath.Join(person.Home, ".local/state", oldname.Name))}
	output, _ := run.CaptureQuiet(run.Minute, "find", append(roots, "-maxdepth", "5", "-type", "l", "(", "-lname", "*/.local/state/solanum/*", "-o", "-lname", "*/.config/solanum/*", ")")...)
	for _, link := range run.Lines(output) {
		target, err := os.Readlink(link)
		if err != nil {
			continue
		}
		old := strings.NewReplacer("/.local/state/solanum/", "/.local/state/"+oldname.Name+"/", "/.config/solanum/", "/.config/"+oldname.Name+"/").Replace(target)
		work.asPerson(person, "ln", "-sfn", old, link)
	}
}

func (work *mover) renameOwnBack(person account.Account) {
	for _, pair := range ownRenames {
		work.moveOwn(person, at(filepath.Join(person.Home, pair[1])), at(filepath.Join(person.Home, pair[0])))
	}
	units := at(filepath.Join(person.Home, ".config/systemd/user"))
	for _, entry := range fsutil.Entries(units) {
		if strings.HasPrefix(entry.Name(), "solanum-") {
			work.moveOwn(person, filepath.Join(units, entry.Name()), filepath.Join(units, oldname.Back(entry.Name())))
		}
	}
	for _, rel := range personFiles {
		path := at(filepath.Join(person.Home, oldname.Back(rel)))
		for _, pair := range suffixes {
			work.moveOwn(person, path+pair[1], path+pair[0])
		}
	}
}
