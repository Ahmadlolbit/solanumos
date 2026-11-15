package main

import (
	"cmp"
	"path/filepath"
	"slices"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const softwareUsage = "usage: solanum software [apps [APP]|packages|webapps]"

var softwarePages = []string{"apps", "packages", "webapps"}

func softwareDir() string {
	return filepath.Join(env.ShareDir(), "software")
}

func openSoftware(page, item string) int {
	if run.Quiet(run.Quick, "qs", "-p", softwareDir(), "ipc", "call", "software", "open", page, item) == 0 {
		return 0
	}
	return run.Quiet(run.Quick, "setsid", "-f", "env", "SOLANUM_SOFTWARE_PAGE="+page, "SOLANUM_SOFTWARE_ITEM="+item, "qs", "-p", softwareDir(), "-n")
}

func softwareCommand() int {
	page, item := cmp.Or(env.Arg(2), "apps"), env.Arg(3)
	switch {
	case page == "-h" || page == "--help":
		return say(softwareUsage + "\n")
	case !slices.Contains(softwarePages, page) || env.Arg(4) != "" || item != "" && page != "apps":
		return fail(softwareUsage)
	}
	return openSoftware(page, item)
}
