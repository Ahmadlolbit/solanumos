package main

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/settings"
)

const settingsUsage = "usage: solanum settings [PAGE|list|get NAME|set NAME VALUE|about]"

func settingsDir() string {
	return filepath.Join(env.ShareDir(), "settings")
}

func openSettings(page string) int {
	if run.Quiet(run.Quick, "qs", "-p", settingsDir(), "ipc", "call", "settings", "open", page) == 0 {
		return 0
	}
	return run.Quiet(run.Quick, "setsid", "-f", "env", "SOLANUM_SETTINGS_PAGE="+page, "qs", "-p", settingsDir(), "-n")
}

func printed(value any) int {
	data, err := json.Marshal(value)
	if err != nil {
		return fail("%v", err)
	}
	return say(string(data) + "\n")
}

func settingValue(name string) int {
	row, found := settings.Find(settings.List().Rows, name)
	if !found {
		return fail("Solanum has no setting named %s; see solanum settings list", name)
	}
	return say(row.Value + "\n")
}

func setSetting(name, value string) int {
	code, err := settings.Set(name, value)
	if err != nil {
		return fail("%v", err)
	}
	return code
}

func pageNames() string {
	var names []string
	for _, page := range settings.Pages() {
		names = append(names, page.ID)
	}
	return strings.Join(names, " ")
}

func settingsCommand() int {
	switch word := env.Arg(2); word {
	case "list":
		return printed(settings.List())
	case "about":
		return printed(settings.Facts())
	case "get":
		return settingValue(env.Arg(3))
	case "set":
		return setSetting(env.Arg(3), env.Arg(4))
	case "", "-h", "--help":
		if word != "" {
			return say(settingsUsage + "\n")
		}
		return openSettings("about")
	default:
		if !settings.KnownPage(word) {
			return fail("no page named %s; the pages are %s\n%s", word, pageNames(), settingsUsage)
		}
		return openSettings(word)
	}
}
