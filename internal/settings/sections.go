package settings

import (
	"cmp"
	"slices"
)

var (
	systemSections = map[string]string{
		"PROFILE": "Apps", "APPS": "Apps", "TEAMS_SOURCE": "Apps", "EXTRAS": "Apps", "BROWSERS_THEME": "Apps", "THEME": "Theme",
		"FAILURE_MODE": "Updates", "DEBIAN_MIRROR": "Updates", "AUTO_UPDATES": "Updates",
		"SNAPSHOTS": "Snapshots", "SNAPSHOT_DAILY": "Snapshots", "SNAPSHOT_WEEKLY": "Snapshots", "SNAPSHOT_MONTHLY": "Snapshots",
		"HOME_SNAPSHOTS": "File history", "HOME_SNAPSHOT_HOURLY": "File history", "HOME_SNAPSHOT_DAILY": "File history", "BACKUP_ON_BATTERY": "File history",
		"KERNEL": "Kernel", "KEPT_KERNEL": "Kernel", "WEBCAM": "Webcam", "GRAPHICS": "Graphics", "SWAPFILE_GB": "Swap", "ZRAM_PERCENT": "Swap", "SWAPPINESS": "Swap",
		"SYSTEM_NAME": "The system's name", "BOOT_SPLASH": "Boot", "BOOT_PROMPT": "Boot", "BOOT_MENU_THEME": "Boot", "CONSOLE_THEME": "Text consoles",
		"HIBERNATE": "Sleep", "IDLE_SUSPEND_MINUTES": "Sleep", "IDLE_LOCK_MINUTES": "Locking", "SCREENSAVER_MINUTES": "Screensaver", "GREETER": "Login screen", "SESSION_RECOVERY": "Login screen",
		"SHELL_PROMPT": "Terminal", "DEV_TOOLS": "Tools", "MISE_TOOLS": "Tools", "CONTAINER_ENGINE": "Containers", "DOCKER_PORTS": "Containers",
		"DISPLAY_SCALE": "Scale and shape", "NATURAL_SCROLL": "Touchpad", "TAP_TO_CLICK": "Touchpad", "ANIMATIONS": "Animations",
		"TIPS_TERMINAL": "Tips", "TIPS_WELCOME": "Tips", "TIPS_NOTIFY": "Tips and warnings", "CARE_NOTIFY": "Tips and warnings", "CRASH_NOTIFY": "Tips and warnings",
		"FIREWALL": "Firewall", "SSH_POLICY": "SSH", "NM_HANDOVER": "Connections", "FINGERPRINT": "Fingerprint",
		"RECOVERY_KEY_REMINDER": "Recovery key",
	}
	sectionOrder = map[string][]string{
		"themes":        {"Colours and fonts", "Windows", "Apps", "Light and dark", "Theme"},
		"wallpaper":     {"Pictures", "Changing", "Live wallpapers"},
		"bar":           {"Layout", "Widgets", "Clock", "Weather"},
		"display":       {"Night light", "Scale and shape", "Brightness", "Presenting"},
		"power":         {"Battery", "Sleep"},
		"mouse":         {"Touchpad", "TrackPoint"},
		"lock":          {"Locking", "Screensaver", "Login screen"},
		"notifications": {"Quiet hours", "Tips and warnings"},
		"updates":       {"Updates", "Snapshots", "File history"},
		"startup":       {"Boot", "Text consoles", "The system's name"},
		"developer":     {"Tools", "Containers", "Terminal"},
	}
)

func pageAt(row Row) int {
	return slices.IndexFunc(pages, func(page Page) bool { return page.ID == row.Page })
}

func sectionAt(row Row) int {
	if row.Whose == Computer && row.Twin != "" {
		return len(sectionOrder[row.Page])
	}
	return slices.Index(sectionOrder[row.Page], row.Section)
}

func ranked(rows []Row) []Row {
	slices.SortStableFunc(rows, func(one, other Row) int {
		return cmp.Or(cmp.Compare(pageAt(one), pageAt(other)), cmp.Compare(sectionAt(one), sectionAt(other)))
	})
	return rows
}
