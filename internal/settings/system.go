package settings

import (
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/options"
)

type placed struct {
	page    string
	modules string
}

var (
	afterRestart = []string{"KERNEL", "GRAPHICS", "WEBCAM", "GREETER", "HIBERNATE", "BOOT_SPLASH", "BOOT_PROMPT", "BOOT_MENU_THEME", "CONSOLE_THEME"}
	systemRows   = map[string]placed{
		"PROFILE": {"apps", ""}, "FAILURE_MODE": {"updates", ""}, "DEBIAN_MIRROR": {"updates", "02"}, "KERNEL": {"hardware", "03"},
		"SYSTEM_NAME": {"startup", "02"}, "WEBCAM": {"hardware", "03"}, "GRAPHICS": {"hardware", "03,12"}, "SWAPFILE_GB": {"storage", "04"},
		"ZRAM_PERCENT": {"storage", "04"}, "SWAPPINESS": {"storage", "04"}, "HIBERNATE": {"power", "04"}, "SNAPSHOTS": {"updates", "05"},
		"SNAPSHOT_DAILY": {"updates", "05"}, "SNAPSHOT_WEEKLY": {"updates", "05"}, "SNAPSHOT_MONTHLY": {"updates", "05"}, "GREETER": {"lock", "07"},
		"SHELL_PROMPT": {"developer", "08"}, "APPS": {"apps", "10"}, "TEAMS_SOURCE": {"apps", "10"}, "DEV_TOOLS": {"developer", "11"},
		"MISE_TOOLS": {"developer", "11"}, "CONTAINER_ENGINE": {"developer", "11"}, "DOCKER_PORTS": {"developer", "11"}, "EXTRAS": {"apps", "10"},
		"DISPLAY_SCALE": {"display", "12"}, "NATURAL_SCROLL": {"mouse", "12"}, "TAP_TO_CLICK": {"mouse", "12"}, "ANIMATIONS": {"windows", "12"},
		"WINDOW_STYLE": {"themes", "14"}, "TEXT_SIZE": {"accessibility", "14"}, "SCREEN_READER": {"accessibility", "14"}, "STARTUP_SOUND": {"sound", "14"},
		"SCREENSAVER_MINUTES": {"lock", "12"}, "TIPS_TERMINAL": {"about", "12"}, "TIPS_WELCOME": {"about", "12"}, "TIPS_NOTIFY": {"notifications", "12"},
		"POWER_FOLLOW": {"power", "14"}, "CARE_NOTIFY": {"notifications", "12"}, "IDLE_LOCK_MINUTES": {"lock", "12"}, "IDLE_SUSPEND_MINUTES": {"power", "12"},
		"BAR_PRESET": {"bar", "13"}, "BAR_POSITION": {"bar", "13"}, "CLOCK_FORMAT": {"time", "12,13,14"}, "LOCK_MEDIA": {"lock", "14"},
		"LOCK_BATTERY": {"lock", "14"}, "THEME": {"themes", "14"}, "ACCENT_COLOR": {"themes", "14"}, "GLASS": {"themes", "14"},
		"ROUNDED_CORNERS": {"display", "14"}, "APPS_THEME": {"themes", "14"}, "EDITORS_THEME": {"themes", "14"}, "BROWSERS_THEME": {"themes", "14"},
		"AUTO_THEME": {"themes", "14"}, "LIGHT_THEME": {"themes", "14"}, "DARK_THEME": {"themes", "14"}, "LIGHT_AT": {"themes", "14"},
		"DARK_AT": {"themes", "14"}, "DAILY_THEME": {"themes", "14"}, "WALLPAPER_MINUTES": {"wallpaper", "14"}, "DAY_WALLPAPER": {"wallpaper", "14"},
		"WALLPAPER_FOLDER": {"wallpaper", "14"}, "WALLPAPER_PER_WORKSPACE": {"wallpaper", "14"}, "WEATHER": {"bar", "14"}, "WEATHER_UNITS": {"bar", "14"},
		"QUIET_HOURS": {"notifications", "14"}, "QUIET_FROM": {"notifications", "14"}, "QUIET_TO": {"notifications", "14"}, "NIGHT_LIGHT": {"display", "14"},
		"NIGHT_FROM": {"display", "14"}, "NIGHT_TO": {"display", "14"}, "BOOT_SPLASH": {"startup", "14"}, "BOOT_PROMPT": {"startup", "14"},
		"BOOT_MENU_THEME": {"startup", "14"}, "CONSOLE_THEME": {"startup", "14"}, "FIREWALL": {"privacy", "15"}, "AUTO_UPDATES": {"updates", "15"},
		"SSH_POLICY": {"sharing", "15"}, "NM_HANDOVER": {"network", "15"}, "FINGERPRINT": {"accounts", "12,15"},
	}
)

func Modules(key string) []string {
	if place, found := systemRows[key]; found && place.modules != "" {
		return strings.Split(place.modules, ",")
	}
	return nil
}

func Option(key string) (options.Option, bool) {
	for _, option := range options.All() {
		if option.Key == key {
			return option, true
		}
	}
	return options.Option{}, false
}

func systemKind(option options.Option) (string, float64, float64) {
	low, high, bounded := option.Range()
	switch {
	case option.Kind == options.YesNo:
		return Toggle, 0, 0
	case option.Kind == options.Choice || option.Kind == options.Multi:
		return Pick, 0, 0
	case bounded && high <= 1440:
		return Slide, float64(low), float64(high)
	}
	return Text, 0, 0
}

func systemRow(option options.Option, facts options.Facts) Row {
	kind, low, high := systemKind(option)
	row := Row{Name: option.Key, Label: option.Label, Page: systemRows[option.Key].page, Kind: kind, Value: facts.Value(option.Key), Whose: Computer,
		Help: option.Help, Hidden: !option.Shows(facts), Min: low, Max: high, Multiple: option.Kind == options.Multi}
	if kind == Slide {
		row.Step = 1
	}
	if kind == Pick {
		row.Choices = named(option.Items())
	}
	if slices.Contains(afterRestart, option.Key) {
		row.After = "restart"
	}
	return row
}
