package main

import (
	"path/filepath"
	"slices"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/snapshots"
)

const careUsage = "usage: solanum care [updates|check|changes|restore [NUMBER]|space|undo|report]"

var carePages = []string{"updates", "check", "changes", "restore", "space", "undo", "report"}

func careDir() string {
	return filepath.Join(env.ShareDir(), "care")
}

func openCare(page string) int {
	if run.Quiet(run.Quick, "qs", "-p", careDir(), "ipc", "call", "care", "open", page) == 0 {
		return 0
	}
	return run.Quiet(run.Quick, "setsid", "-f", "env", "SOLANUM_CARE_PAGE="+page, "qs", "-p", careDir(), "-n")
}

func careCommand() int {
	page, point := env.Arg(2), env.Arg(3)
	switch {
	case page == "-h" || page == "--help":
		return say(careUsage + "\n")
	case env.Arg(4) != "" || point != "" && (page != "restore" || !snapshots.IsNumber(point)):
		return fail(careUsage)
	case page == "":
		return openCare("updates")
	case !slices.Contains(carePages, page):
		return fail("no page named %s; the pages are updates, check, changes, restore, space, undo and report\n%s", page, careUsage)
	case point != "":
		return openCare(page + ":" + point)
	}
	return openCare(page)
}
