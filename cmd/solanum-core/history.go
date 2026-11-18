package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/config"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/restic"
	"github.com/Ahmadlolbit/solanumos/internal/snapshots"
)

const (
	historyUsage = "usage: solanum history [--json | files NUMBER FOLDER --json | restore NUMBER PATH [--copy]]"
	snapperDate  = "2006-01-02 15:04:05"
)

type homeHistory struct {
	State  string            `json:"state"`
	Hourly string            `json:"hourly"`
	Daily  string            `json:"daily"`
	Points []snapshots.Point `json:"points"`
	Unread bool              `json:"unread"`
}

func homeRoot() string {
	return env.Or("SOLANUM_HOME_ROOT", "/home")
}

func homeSnapshotsDir() string {
	return env.Or("SOLANUM_HOME_SNAPSHOTS", filepath.Join(homeRoot(), ".snapshots"))
}

func historyJSON() int {
	settings := config.Read(config.Path())
	found := homeHistory{State: "off", Hourly: settings.Get("HOME_SNAPSHOT_HOURLY"), Daily: settings.Get("HOME_SNAPSHOT_DAILY"), Points: []snapshots.Point{}}
	if settings.Get("SNAPSHOTS") == "yes" && settings.Get("HOME_SNAPSHOTS") == "yes" {
		found.State = "on"
	}
	if !fsutil.IsDir(homeSnapshotsDir()) {
		found.State = "none"
		return printed(found)
	}
	points, err := snapshots.Of("home")
	found.Unread = err != nil
	slices.Reverse(points)
	found.Points = snapshots.Points(points)
	return printed(found)
}

func inSnapshot(number, path string) (string, bool) {
	clean, root := filepath.Clean(path), filepath.Clean(homeRoot())
	if !snapshots.IsNumber(number) || !restic.Inside(path, env.Home()) && clean != filepath.Clean(env.Home()) || !restic.Inside(env.Home(), root) {
		return "", false
	}
	return filepath.Join(homeSnapshotsDir(), number, "snapshot", strings.TrimPrefix(clean, root+"/")), true
}

func historyFiles(number, folder string) int {
	source, ok := inSnapshot(number, folder)
	if !ok {
		return fail("%s isn't a folder in your home that snapshot %s can hold", folder, number)
	}
	entries, err := restic.Listed(source)
	if err != nil {
		return fail("%s can't be read in snapshot %s", folder, number)
	}
	return printed(entries)
}

func snapshotWhen(number string) string {
	points, _ := snapshots.Of("home")
	for _, point := range points {
		if at, err := time.ParseInLocation(snapperDate, point.Date, time.Local); err == nil && point.Number == number {
			return at.Format(restic.WhenLayout)
		}
	}
	return "snapshot " + number
}

func historyRestore(number, path string, asCopy bool) int {
	source, ok := inSnapshot(number, path)
	if !ok || !restic.Inside(path, env.Home()) {
		return fail("%s isn't a file or folder in your home that can be restored", path)
	}
	info, err := os.Lstat(source)
	if err != nil {
		return fail("%s isn't in snapshot %s", path, number)
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Clean(path)), 0o777); err != nil {
		return fail("%v", err)
	}
	placed, aside, err := restic.PutBack(source, filepath.Clean(path), snapshotWhen(number), info.IsDir(), asCopy)
	if err != nil {
		return fail("%s couldn't be restored: %v", path, err)
	}
	text := "Restored " + placed + "\n"
	if aside != "" {
		text += "What was there is now " + aside + "\n"
	}
	return say(text)
}

func historyCommand() int {
	words := os.Args[min(2, len(os.Args)):]
	switch {
	case len(words) == 1 && words[0] == "--json":
		return historyJSON()
	case len(words) == 4 && words[0] == "files" && words[3] == "--json":
		return historyFiles(words[1], words[2])
	case len(words) == 3 && words[0] == "restore":
		return historyRestore(words[1], words[2], false)
	case len(words) == 4 && words[0] == "restore" && words[3] == "--copy":
		return historyRestore(words[1], words[2], true)
	}
	return fail(historyUsage)
}
