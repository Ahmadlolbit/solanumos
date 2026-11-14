package move

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/files"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/oldname"
)

const databaseMode = 0o600

var (
	cleanedUp = []string{"/etc/xdg/walker/themes/" + oldname.Name, "/usr/local/bin/" + oldname.Prefix + "power-menu", "/usr/local/bin/" + oldname.Prefix + "theme-picker"}
	suffixes  = [][2]string{{oldname.OldSuffix, ".solanum-old"}, {oldname.NewSuffix, ".solanum-new"}}
)

func homeOf(path string) (string, string) {
	rest, found := strings.CutPrefix(path, "/home/")
	if !found {
		return "", path
	}
	name, tail, _ := strings.Cut(rest, "/")
	return "/home/" + name, "/" + tail
}

func inAHome(path string) bool {
	home, _ := homeOf(strings.TrimPrefix(path, strings.TrimSuffix(at("/"), "/")))
	return home != ""
}

func swapPath(path string) string {
	home, tail := homeOf(path)
	if home != "" && (tail == "/"+oldname.Clone || strings.HasPrefix(tail, "/"+oldname.Clone+"/")) {
		return path
	}
	return home + oldname.Swap(tail)
}

func gone(folder string) bool {
	return fsutil.IsLink(at(folder)) || !fsutil.IsDir(at(folder)) && fsutil.IsDir(at(oldname.Swap(folder)))
}

func moved(path string) string {
	for _, each := range folders {
		if rest, found := strings.CutPrefix(path, each.old+"/"); found && gone(each.old) {
			return oldname.Swap(each.old) + "/" + rest
		}
	}
	return path
}

func ours(path string) bool {
	return !slices.ContainsFunc(cleanedUp, func(old string) bool { return path == old || strings.HasPrefix(path, old+"/") })
}

func unchanged(path, hash string) bool {
	found, err := files.Hash(path)
	return err == nil && found == hash
}

func (work *mover) moveSaved(now, fresh string) {
	for _, pair := range suffixes {
		work.moveFile(at(now+pair[0]), at(fresh+pair[1]))
	}
	work.remove(at(now + oldname.PartSuffix))
}

func (work *mover) movePath(path, hash string) string {
	if !ours(path) {
		return path
	}
	if strings.HasPrefix(path, "/usr/local/bin/"+oldname.Name) {
		return ""
	}
	now, fresh := moved(path), swapPath(path)
	work.moveSaved(now, fresh)
	if now == fresh || !fsutil.Exists(at(now)) && !fsutil.IsLink(at(now)) {
		return fresh
	}
	switch {
	case !fsutil.Exists(at(fresh)) && !fsutil.IsLink(at(fresh)):
		if !work.rename(now, fresh) {
			return path
		}
	case sameFile(at(now), at(fresh)):
	case unchanged(at(now), hash):
		work.remove(at(now))
	default:
		work.note("Left %s where it is: you changed it, and %s is there too", now, fresh)
		return path
	}
	return fresh
}

func sameFile(one, other string) bool {
	first, err := os.Lstat(one)
	if err != nil {
		return false
	}
	second, err := os.Lstat(other)
	return err == nil && os.SameFile(first, second)
}

func (work *mover) rename(now, fresh string) bool {
	if person, found := work.ownerOf(now); found {
		work.asPerson(person, "mkdir", "-p", filepath.Dir(at(fresh)))
		return work.asPerson(person, "mv", "-T", "--update=none-fail", at(now), at(fresh))
	}
	if inAHome(now) {
		work.note("Left %s where it is: its owner couldn't be found", now)
		return false
	}
	err := os.MkdirAll(filepath.Dir(at(fresh)), folderMode)
	if err == nil {
		err = atomic.Move(at(now), at(fresh))
	}
	if err != nil {
		work.note("Left %s where it is: %v", now, err)
	}
	return err == nil
}

func (work *mover) moveDatabase() {
	path := filepath.Join(stateDir(), "files.db")
	lines, err := fsutil.Lines(path)
	if err != nil || len(lines) == 0 {
		return
	}
	hashes := map[string]string{}
	for _, line := range lines {
		known, hash, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		if fresh := work.movePath(known, hash); fresh != "" {
			hashes[fresh] = hash
		}
	}
	var text strings.Builder
	for _, known := range slices.Sorted(maps.Keys(hashes)) {
		text.WriteString(known + "\t" + hashes[known] + "\n")
	}
	if err := atomic.Write(path, []byte(text.String()), databaseMode); err != nil {
		work.note("files.db keeps the old names: %v", err)
	}
}

func recordLine(line string) string {
	kind, value, found := strings.Cut(line, " ")
	switch {
	case !found:
		return line
	case kind == "file" || kind == "made" || kind == "replaced":
		return kind + " " + swapPath(value)
	case kind == "service" && strings.HasPrefix(value, oldname.Prefix):
		return kind + " " + oldname.Swap(value)
	}
	return line
}

func (work *mover) moveRecords() {
	work.moveDatabase()
	records, _ := filepath.Glob(filepath.Join(stateDir(), "modules", "*.record"))
	for _, record := range records {
		lines, err := fsutil.Lines(record)
		if err != nil {
			continue
		}
		var text strings.Builder
		for _, line := range lines {
			text.WriteString(recordLine(line) + "\n")
		}
		if err := atomic.Write(record, []byte(text.String()), stampMode); err != nil {
			work.note("%s keeps the old names: %v", record, err)
		}
	}
}
