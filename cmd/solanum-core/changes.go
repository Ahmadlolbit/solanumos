package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/notify"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

func changesDir() string {
	return env.OwnState()
}

func notifyFirstStart(dir string) {
	note := filepath.Join(dir, "first-start.txt")
	if !fsutil.IsFile(note) {
		return
	}
	remove(note)
	action := notify.Ask([]string{"--app-name=solanum", "--action=default=Read"}, "The setup isn't finished",
		"The first start couldn't do all of it. Connect to the internet, then run: sudo solanum install --yes")
	if action == "default" {
		run.Status("solanum-terminal", "--wait", "less", filepath.Join(os.Getenv("HOME"), "runbook-report.md"))
	}
}

func notifyChanges(dir string) int {
	notifyFirstStart(dir)
	unseen := filepath.Join(dir, "changes-unseen")
	if !fsutil.IsFile(unseen) {
		return 0
	}
	text := fsutil.Optional(filepath.Join(dir, "changes.txt"))
	if text == "" {
		return 0
	}
	remove(unseen)
	count := strings.Count("\n"+text, "\n  - ")
	action := notify.Ask([]string{"--app-name=solanum", "--action=default=Read"}, "Solanum was updated",
		fmt.Sprintf("%d changes since the last update. Click to read them, or open Menu › Update › What changed.", count))
	if action == "default" {
		return run.Report(run.Exec("solanum-terminal", "--wait", script(), "changes"))
	}
	return 0
}

type changeGroup struct {
	Heading string   `json:"heading"`
	Items   []string `json:"items"`
}

type changeNotes struct {
	Date   string        `json:"date"`
	Groups []changeGroup `json:"groups"`
}

func notesGroups(text string) []changeGroup {
	groups := []changeGroup{}
	for _, line := range strings.Split(text, "\n") {
		item, listed := strings.CutPrefix(line, "  - ")
		switch {
		case strings.TrimSpace(line) == "":
		case !listed:
			groups = append(groups, changeGroup{Heading: line, Items: []string{}})
		case len(groups) == 0:
			groups = append(groups, changeGroup{Items: []string{item}})
		default:
			groups[len(groups)-1].Items = append(groups[len(groups)-1].Items, item)
		}
	}
	return groups
}

func changesJSON(dir string) int {
	path := filepath.Join(dir, "changes.txt")
	notes := changeNotes{Groups: notesGroups(fsutil.Optional(path))}
	if info, err := os.Stat(path); err == nil && len(notes.Groups) > 0 {
		notes.Date = info.ModTime().Format(time.RFC3339)
	}
	return printed(notes)
}

func changes() int {
	dir := changesDir()
	if env.Arg(3) == "" {
		switch env.Arg(2) {
		case "--notify":
			return notifyChanges(dir)
		case "--json":
			return changesJSON(dir)
		}
	}
	if env.Arg(2) != "" {
		return fail("usage: solanum changes")
	}
	if text := fsutil.Optional(filepath.Join(dir, "changes.txt")); text != "" {
		say(text + "\n")
	} else {
		say("No update notes yet. solanum update writes them.\n")
	}
	return remove(filepath.Join(dir, "changes-unseen"))
}
