package grub

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/bootimage"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	ownFile     = "solanum-kernel.cfg"
	byHand      = "90-keep-*.cfg"
	submenuID   = "gnulinux-advanced-"
	defaultLine = "GRUB_DEFAULT="
)

var (
	menuLine   = regexp.MustCompile(`^\s*(menuentry|submenu) '((?:[^']|'\\'')*)'.*\$menuentry_id_option '([^']*)'`)
	defaultSet = regexp.MustCompile(`^\s*set default="([^"$]*)"\s*$`)
	advanced   = regexp.MustCompile(`^gnulinux-(.+)-advanced-[^>]+$`)
	unquoting  = strings.NewReplacer(`'\''`, `'`)
)

type Entry struct {
	ID      string
	Submenu string
	Title   string
}

func Defaults() string {
	return env.Or("SOLANUM_GRUB_DEFAULTS", "/etc/default/grub.d")
}

func Menu() string {
	return filepath.Join(bootimage.Folder(), "grub", "grub.cfg")
}

func Find(menu, release string) (Entry, bool) {
	var submenu Entry
	for _, line := range strings.Split(menu, "\n") {
		found := menuLine.FindStringSubmatch(line)
		switch {
		case found == nil:
		case found[1] == "submenu" && strings.HasPrefix(found[3], submenuID):
			submenu = Entry{ID: found[3], Title: unquoting.Replace(found[2])}
		case found[1] == "menuentry" && Release(found[3]) == release:
			entry := Entry{ID: found[3], Title: unquoting.Replace(found[2])}
			if submenu.ID != "" {
				entry.ID, entry.Submenu = submenu.ID+">"+entry.ID, submenu.Title
			}
			return entry, true
		}
	}
	return Entry{}, false
}

func Release(id string) string {
	last := id[strings.LastIndexByte(id, '>')+1:]
	if found := advanced.FindStringSubmatch(last); found != nil {
		return found[1]
	}
	return ""
}

func named(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if value, found := strings.CutPrefix(strings.TrimSpace(line), defaultLine); found {
			return Release(strings.Trim(value, `'"`))
		}
	}
	return ""
}

func own() string {
	return filepath.Join(Defaults(), ownFile)
}

func Kept() string {
	return named(fsutil.Optional(own()))
}

func handMade() []string {
	files, _ := filepath.Glob(filepath.Join(Defaults(), byHand))
	var kept []string
	for _, file := range files {
		if named(fsutil.Optional(file)) != "" {
			kept = append(kept, file)
		}
	}
	return kept
}

func HandMade() string {
	if files := handMade(); len(files) > 0 {
		return named(fsutil.Optional(files[0]))
	}
	return ""
}

func menuDefault() string {
	for _, line := range strings.Split(fsutil.Optional(Menu()), "\n") {
		if found := defaultSet.FindStringSubmatch(line); found != nil {
			return found[1]
		}
	}
	return ""
}

func entryFor(log io.Writer, release string) (Entry, error) {
	if entry, ok := Find(fsutil.Optional(Menu()), release); ok {
		return entry, nil
	}
	if err := run.Logged(log, "update-grub"); err != nil {
		return Entry{}, err
	}
	if entry, ok := Find(fsutil.Optional(Menu()), release); ok {
		return entry, nil
	}
	return Entry{}, fmt.Errorf("the boot menu has no entry for %s", release)
}

func kernelPackage(release string) string {
	output, _ := run.CaptureQuiet(run.Quick, "dpkg-query", "-S", filepath.Join(bootimage.Folder(), "vmlinuz-"+release))
	for _, line := range run.Lines(output) {
		if name, _, found := strings.Cut(line, ": "); found && !strings.Contains(name, " ") {
			return name
		}
	}
	return "linux-image-" + release
}

func mark(log io.Writer, how, release string) error {
	return run.LoggedWithin(run.Minute, log, "apt-mark", how, kernelPackage(release))
}

func removeAll(log io.Writer, files []string) (bool, error) {
	removed := false
	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return removed, err
		}
		fmt.Fprintf(log, "removed %s\n", file)
		removed = true
	}
	return removed, nil
}

func Keep(log io.Writer, release string) error {
	entry, err := entryFor(log, release)
	if err != nil {
		return err
	}
	line := defaultLine + "'" + entry.ID + "'"
	if fsutil.Optional(own()) == line && len(handMade()) == 0 && menuDefault() == entry.ID {
		return nil
	}
	if err := mark(log, "manual", release); err != nil {
		return err
	}
	if err := os.MkdirAll(Defaults(), 0o755); err != nil {
		return err
	}
	if err := atomic.Write(own(), []byte(line+"\n"), 0o644); err != nil {
		return err
	}
	if _, err := removeAll(log, handMade()); err != nil {
		return err
	}
	return run.Logged(log, "update-grub")
}

func Newest(log io.Writer) error {
	kept := Kept()
	files := handMade()
	if kept != "" {
		files = append(files, own())
	}
	removed, err := removeAll(log, files)
	if err != nil || !removed && Release(menuDefault()) == "" {
		return err
	}
	if kept != "" {
		if err := mark(log, "auto", kept); err != nil {
			return err
		}
	}
	return run.Logged(log, "update-grub")
}
