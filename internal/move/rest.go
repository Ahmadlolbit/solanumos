package move

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/oldname"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

var remotes = []string{"https://github.com/Ahmadlolbit/", "git@github.com:Ahmadlolbit/"}

func (work *mover) moveSavedInEtc() {
	filepath.WalkDir(at("/etc"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		for _, pair := range suffixes {
			if base, found := strings.CutSuffix(path, pair[0]); found {
				work.moveFile(path, base+pair[1])
			}
		}
		if strings.HasSuffix(path, oldname.PartSuffix) {
			work.remove(path)
		}
		return nil
	})
}

func (work *mover) moveConnections() {
	folder := at("/etc/NetworkManager/system-connections")
	moved := false
	for _, entry := range fsutil.Entries(folder) {
		if strings.HasPrefix(entry.Name(), oldname.Prefix) {
			work.moveFile(filepath.Join(folder, entry.Name()), filepath.Join(folder, oldname.Swap(entry.Name())))
			moved = true
		}
	}
	if moved {
		work.quietly("nmcli", "connection", "reload")
	}
}

func newRemote(url string) string {
	for _, start := range remotes {
		for _, end := range []string{"", ".git"} {
			if url == start+oldname.Name+end {
				return start + "solanumos" + end
			}
		}
	}
	return ""
}

func (work *mover) cloneOwner(clone string) (account.Account, bool) {
	return work.ownerOf(strings.TrimPrefix(clone, strings.TrimSuffix(at("/"), "/")))
}

func (work *mover) inClone(clone string, args ...string) (string, bool) {
	command := append([]string{"-C", clone}, args...)
	if person, found := work.cloneOwner(clone); found {
		output, err := run.CaptureQuiet(run.Minute, "runuser", person.RunuserArgs(append([]string{"git"}, command...)...)...)
		return output, err == nil
	}
	output, err := run.CaptureQuiet(run.Minute, "git", command...)
	return output, err == nil
}

func (work *mover) moveRemote(clone string) {
	url, ok := work.inClone(clone, "remote", "get-url", "origin")
	if fresh := newRemote(strings.TrimSpace(url)); ok && fresh != "" {
		if _, done := work.inClone(clone, "remote", "set-url", "origin", fresh); done {
			work.note("The clone in %s now fetches from %s", clone, fresh)
		}
	}
}

func (work *mover) tidyClone(clone string) {
	built, _ := filepath.Glob(filepath.Join(clone, "build", "bin", oldname.Name+"*"))
	if len(built) == 0 {
		return
	}
	if person, found := work.cloneOwner(clone); found {
		work.asPerson(person, append([]string{"rm", "-f", "--"}, built...)...)
		return
	}
	for _, path := range built {
		work.remove(path)
	}
}

func (work *mover) moveRest() {
	work.moveSavedInEtc()
	work.moveConnections()
	if clone := env.Root(); clone != "" && fsutil.IsDir(filepath.Join(clone, ".git")) {
		work.moveRemote(clone)
		work.tidyClone(clone)
	}
	work.quietly("systemctl", "daemon-reload")
}
