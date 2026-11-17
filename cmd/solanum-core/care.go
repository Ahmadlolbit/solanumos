package main

import (
	"path/filepath"
	"slices"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const careUsage = "usage: solanum care [updates|check|changes]"

var carePages = []string{"updates", "check", "changes"}

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
	page := env.Arg(2)
	switch {
	case page == "-h" || page == "--help":
		return say(careUsage + "\n")
	case env.Arg(3) != "":
		return fail(careUsage)
	case page == "":
		return openCare("updates")
	case !slices.Contains(carePages, page):
		return fail("no page named %s; the pages are updates, check and changes\n%s", page, careUsage)
	}
	return openCare(page)
}
