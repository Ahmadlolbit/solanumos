package main

import (
	"cmp"
	"fmt"
	"os"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/apps"
	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/config"
	"github.com/Ahmadlolbit/solanumos/internal/engine"
	"github.com/Ahmadlolbit/solanumos/internal/env"
)

const appUsage = "usage: solanum app [list [installed|available|--json] | install APP | remove APP]"

type listedApp struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Group     string `json:"group"`
	Source    string `json:"source"`
	From      string `json:"from"`
	Installed bool   `json:"installed"`
}

func appFrom(source, teamsSource string) string {
	kind, _, _ := strings.Cut(source, ":")
	switch {
	case kind == "debian":
		return "Debian"
	case kind == "release":
		return "Maker's release"
	case kind == "webapp":
		return "Web app"
	case kind == "teams" && teamsSource == "snap":
		return "Snap"
	}
	return "Maker's repository"
}

func appsListed(catalogue []apps.App, present func(apps.App) bool, teamsSource string) []listedApp {
	listed := []listedApp{}
	for _, category := range apps.Categories {
		for _, app := range apps.InCategory(catalogue, category[0]) {
			listed = append(listed, listedApp{app.ID, app.Name, category[0], category[1], app.Source, appFrom(app.Source, teamsSource), present(app)})
		}
	}
	return listed
}

func filteredApps(catalogue []apps.App, present func(apps.App) bool, filter string) string {
	var list strings.Builder
	for _, app := range catalogue {
		if (filter == "installed" && present(app)) || (filter == "available" && !present(app)) {
			fmt.Fprintf(&list, "%s\t%s\n", app.ID, app.Name)
		}
	}
	return list.String()
}

func appsByCategory(catalogue []apps.App, present func(apps.App) bool) string {
	var list strings.Builder
	for _, category := range apps.Categories {
		group := apps.InCategory(catalogue, category[0])
		if len(group) == 0 {
			continue
		}
		list.WriteString(category[1] + "\n")
		for _, app := range group {
			mark := " "
			if present(app) {
				mark = "*"
			}
			fmt.Fprintf(&list, "  %s %-14s %s\n", mark, app.ID, app.Name)
		}
	}
	return list.String() + "\n* installed\n"
}

func appChange(action string, change func(apps.App, engine.Options) int) int {
	options := engine.Options{Events: env.Arg(3) == "--events"}
	id := env.Arg(3)
	if options.Events {
		id = env.Arg(4)
	}
	if id == "" {
		return fail("usage: sudo solanum app %s APP (see: solanum app list)", action)
	}
	app, found := apps.Find(id)
	if !found {
		return fail("no app named %s (see: solanum app list)", id)
	}
	if os.Getenv("SOLANUM_ENGINE") == "" {
		return throughScript()
	}
	return asRoot("app "+action+" "+id, func() int { return change(app, options) })
}

func appCommand() int {
	switch action := cmp.Or(env.Arg(2), "list"); action {
	case "install":
		return appChange(action, engine.AddApp)
	case "remove":
		return appChange(action, engine.RemoveApp)
	case "list":
		return appList()
	}
	return fail(appUsage)
}

func appList() int {
	packages, err := apt.Installed()
	if err != nil {
		return fail("the installed packages can't be read: %v", err)
	}
	teamsSource := config.Read(config.Path()).Get("TEAMS_SOURCE")
	present := func(app apps.App) bool { return app.Present(packages, teamsSource) }
	catalogue := apps.All()
	if env.Arg(3) == "--json" {
		return printed(appsListed(catalogue, present, teamsSource))
	}
	if filter := env.Arg(3); filter != "" {
		return say(filteredApps(catalogue, present, filter))
	}
	return say(appsByCategory(catalogue, present))
}
