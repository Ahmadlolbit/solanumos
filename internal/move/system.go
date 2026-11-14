package move

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/files"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/oldname"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

type folder struct {
	old  string
	link bool
}

var folders = []folder{
	{"/etc/" + oldname.Name, true},
	{"/var/lib/" + oldname.Name, true},
	{"/var/cache/" + oldname.Name, true},
	{"/var/log/" + oldname.Name, true},
	{"/var/log/" + oldname.Name + "-live", false},
	{"/usr/share/" + oldname.Name, true},
	{"/usr/share/plymouth/themes/" + oldname.Name, false},
	{"/usr/share/sddm/themes/" + oldname.Name, true},
	{"/boot/grub/themes/" + oldname.Name, true},
}

var renamedInside = [][2]string{
	{"/etc/solanum/" + oldname.Name + ".conf", "/etc/solanum/solanum.conf"},
	{"/var/lib/solanum/sources/" + oldname.ThemesSource, "/var/lib/solanum/sources/solanumos-themes"},
	{"/var/cache/solanum/sources/" + oldname.ThemesSource, "/var/cache/solanum/sources/solanumos-themes"},
}

var leftovers = []string{"/usr/share/plymouth/themes/solanum/" + oldname.Name + ".script"}

func (work *mover) quietly(name string, args ...string) {
	if err := run.LoggedWithin(run.Minute, work.out, name, args...); err != nil {
		fmt.Fprintf(work.out, "%s %s: %v\n", name, strings.Join(args, " "), err)
	}
}

func oldUnits(dir string) []string {
	var found []string
	for _, entry := range fsutil.Entries(at(dir)) {
		if strings.HasPrefix(entry.Name(), oldname.Prefix) && !entry.IsDir() {
			found = append(found, entry.Name())
		}
	}
	return found
}

func (work *mover) stopUnits() {
	first := oldname.Prefix + "first-boot.service"
	for _, unit := range oldUnits("/etc/systemd/system") {
		if unit == first {
			work.quietly("systemctl", "disable", unit)
			continue
		}
		work.quietly("systemctl", "disable", "--now", unit)
	}
	for _, unit := range oldUnits("/etc/systemd/user") {
		work.quietly("systemctl", "--global", "disable", unit)
	}
	work.moveFirstBoot(first)
}

func firstBootUnit(text string) string {
	replacer := strings.NewReplacer(strings.ToUpper(oldname.Name)+"_USER", "SOLANUM_USER", "/etc/"+oldname.Name+"/", "/etc/solanum/", "Description=Finish setting up "+oldname.Back("Solanum"), "Description=Finish setting up Solanum")
	return replacer.Replace(text)
}

func (work *mover) moveFirstBoot(unit string) {
	old := at("/etc/systemd/system/" + unit)
	if !fsutil.IsFile(old) {
		return
	}
	if fsutil.Exists(at("/etc/" + oldname.Name + "/first-boot")) {
		fresh := oldname.Swap(unit)
		text, err := os.ReadFile(old)
		if err == nil {
			err = atomic.Install(at("/etc/systemd/system/"+fresh), []byte(firstBootUnit(string(text))), stampMode)
		}
		if err != nil {
			work.note("The first start's unit stays under its old name: %v", err)
			return
		}
		work.quietly("systemctl", "enable", fresh)
	}
	work.remove(old)
}

func (work *mover) remove(path string) {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(work.out, "cannot remove %s: %v\n", path, err)
	}
}

func (work *mover) moveFolders() {
	for _, each := range folders {
		old := at(each.old)
		work.moveFolder(old, oldname.Swap(old), each.link)
	}
	for _, pair := range renamedInside {
		work.moveFile(at(pair[0]), at(pair[1]))
	}
	for _, path := range leftovers {
		work.remove(at(path))
	}
}

func (work *mover) moveFolder(old, fresh string, link bool) {
	var err error
	switch {
	case fsutil.IsLink(old) || !fsutil.IsDir(old):
		return
	case !fsutil.Exists(fresh) && !fsutil.IsLink(fresh):
		if err = os.MkdirAll(filepath.Dir(fresh), folderMode); err == nil {
			err = atomic.Move(old, fresh)
		}
	default:
		if !work.merge(old, fresh) {
			return
		}
	}
	if err != nil {
		work.note("Left %s where it is: %v", old, err)
		return
	}
	if link {
		work.link(old, fresh)
	}
}

func (work *mover) link(old, fresh string) {
	if err := atomic.Symlink(filepath.Base(fresh), old); err != nil {
		work.note("%s couldn't point to %s: %v", old, fresh, err)
	}
}

func spare(from, to string) bool {
	if fsutil.IsLink(from) {
		return true
	}
	one, err := files.Hash(from)
	other, otherErr := files.Hash(to)
	return err == nil && otherErr == nil && one == other
}

func (work *mover) merge(old, fresh string) bool {
	var left []string
	for _, entry := range fsutil.Entries(old) {
		from, to := filepath.Join(old, entry.Name()), filepath.Join(fresh, entry.Name())
		switch {
		case !fsutil.Exists(to) && !fsutil.IsLink(to):
			if atomic.Move(from, to) != nil {
				left = append(left, entry.Name())
			}
		case spare(from, to):
			work.remove(from)
		default:
			left = append(left, entry.Name())
		}
	}
	if len(left) > 0 {
		slices.Sort(left)
		work.note("Left %s where it is: %s already has %s", old, fresh, strings.Join(left, ", "))
		return false
	}
	return os.Remove(old) == nil
}

func (work *mover) moveFile(old, fresh string) {
	switch {
	case !fsutil.Exists(old) && !fsutil.IsLink(old):
	case fsutil.Exists(fresh) || fsutil.IsLink(fresh):
		work.note("Left %s where it is: %s is there too", old, fresh)
	default:
		if err := atomic.Move(old, fresh); err != nil {
			work.note("Left %s where it is: %v", old, err)
		}
	}
}
