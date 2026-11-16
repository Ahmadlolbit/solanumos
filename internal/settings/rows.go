package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/config"
	"github.com/Ahmadlolbit/solanumos/internal/detect"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/look"
	"github.com/Ahmadlolbit/solanumos/internal/options"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

type barRow struct {
	key      string
	label    string
	page     string
	section  string
	kind     string
	choices  string
	fallback string
	help     string
}

var barRows = []barRow{
	{"preset", "Bar layout", "bar", "Layout", Pick, "full:Full width|floating:Floating|minimal:Minimal|vertical:Down the left side", "full", "Full width, floating with rounded corners, minimal, or down the left side"},
	{"position", "Bar position", "bar", "Layout", Pick, "top:Top|bottom:Bottom", "top", "The bar at the top or the bottom of the screen"},
	{"style", "Bar style", "bar", "Layout", Pick, ":Like the windows|sharp:Sharp|soft:Soft", "", "The bar's corners: like the windows, sharp or soft"},
	{"clockFormat", "Clock details", "bar", "Clock", Text, "", "", "The clock's format in Qt's words, such as ddd d MMM HH:mm; empty keeps the default"},
	{"workspaces", "Workspaces shown", "bar", "Widgets", Text, "", "", "The workspace numbers the bar always shows, such as 1 2 3; empty shows the ones in use"},
	{"left", "Widgets on the left", "bar", "Widgets", Text, "", "workspaces window", "The bar's widgets on the left, in order"},
	{"center", "Widgets in the middle", "bar", "Widgets", Text, "", "media clock focus", "The bar's widgets in the middle, in order"},
	{"right", "Widgets on the right", "bar", "Widgets", Text, "", "privacy indicators updates snapshots weather tray cpu memory temperature network bluetooth audio brightness power-profile battery notifications", "The bar's widgets on the right, in order"},
	{"clock", "24-hour clock", "time", "Clock", Pick, "24h:24-hour|12h:12-hour", "24h", "24-hour or 12-hour time in the bar and on the lock screen"},
}

func shellFile() string {
	return filepath.Join(env.ConfigHome(), "solanum", "shell.json")
}

func barValue(key, fallback string) string {
	for _, path := range []string{shellFile(), env.Or("SOLANUM_SHELL_SYSTEM", "/etc/solanum/shell.json")} {
		var shell struct {
			Bar map[string]json.RawMessage `json:"bar"`
		}
		data, err := os.ReadFile(path)
		if err != nil || json.Unmarshal(data, &shell) != nil {
			continue
		}
		if raw, found := shell.Bar[key]; found {
			return plainJSON(raw)
		}
	}
	return fallback
}

func plainJSON(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var list []any
	if json.Unmarshal(raw, &list) != nil {
		return strings.TrimSpace(string(raw))
	}
	words := make([]string, 0, len(list))
	for _, item := range list {
		if number, ok := item.(float64); ok {
			words = append(words, strconv.FormatFloat(number, 'f', -1, 64))
			continue
		}
		words = append(words, fmt.Sprint(item))
	}
	return strings.Join(words, " ")
}

func themeNames() []string {
	names := options.Themes()
	for _, entry := range fsutil.Entries(filepath.Join(env.AppConfig(), "themes")) {
		if theme.ValidName(entry.Name()) && fsutil.IsFile(filepath.Join(env.AppConfig(), "themes", entry.Name(), "colors.toml")) {
			names = append(names, entry.Name())
		}
	}
	slices.Sort(names)
	return slices.Compact(names)
}

func themeChoices(first ...Choice) []Choice {
	for _, name := range themeNames() {
		first = append(first, Choice{name, theme.Pretty(name)})
	}
	return first
}

func fonts(arguments ...string) []Choice {
	output, _ := run.CaptureQuiet(run.Quick, "fc-list", append([]string{"--format", "%{family[0]}\n"}, arguments...)...)
	var names []string
	for _, line := range strings.Split(output, "\n") {
		if name := strings.TrimSpace(line); theme.ValidFont(name) {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return named(slices.Compact(names))
}

func yourChoices(key string) []Choice {
	switch key {
	case "light", "dark":
		return themeChoices()
	case "apps":
		return themeChoices(Choice{"follow", "Follow the theme"}, Choice{"off", "Keep their own look"})
	case "font":
		return fonts(":spacing=100")
	case "ui_font":
		return append([]Choice{{"default", "The default"}}, fonts(":")...)
	}
	return nil
}

func yourRow(row yours, settings *look.Store) Row {
	shown := Row{Name: row.key, Label: row.label, Page: row.page, Section: row.section, Kind: row.kind, Value: settings.Get(row.key), Whose: Yours, Help: row.help,
		Choices: choicesOf(row.choices), Hidden: row.shows != nil && !row.shows(settings.Get)}
	if shown.Choices == nil {
		shown.Choices = yourChoices(row.key)
	}
	switch row.key {
	case "night_warmth":
		shown.Min, shown.Max, shown.Step, shown.Unit = 1000, 6500, 100, "K"
	case "trackpoint_speed":
		shown.Min, shown.Max, shown.Step = -1, 1, 0.1
	case "reader":
		shown.After = "login"
	case "logo":
		if shown.Value != "animated" && shown.Value != "life" {
			shown.Value = "off"
		}
	}
	return shown
}

func List() Listing {
	settings := look.Open()
	var rows []Row
	for _, row := range barRows {
		rows = append(rows, Row{Name: "bar." + row.key, Label: row.label, Page: row.page, Section: row.section, Kind: row.kind, Value: barValue(row.key, row.fallback), Whose: Yours, Help: row.help, Choices: choicesOf(row.choices)})
	}
	for _, row := range yourRows {
		rows = append(rows, yourRow(row, settings))
	}
	current := config.Read(config.Path())
	facts := current.Facts(detect.NvidiaOnBus())
	for _, option := range options.All() {
		rows = append(rows, systemRow(option, facts))
	}
	for _, button := range buttons() {
		button.Kind, button.Whose = Button, Yours
		rows = append(rows, button)
	}
	return Listing{Pages: pages, Rows: ranked(paired(rows))}
}
