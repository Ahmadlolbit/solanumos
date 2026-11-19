package settings

import (
	"strings"
)

type yours struct {
	key     string
	label   string
	page    string
	section string
	kind    string
	choices string
	help    string
	shows   func(get func(string) string) bool
}

var (
	accents  = "theme:The theme's own|red:Red|orange:Orange|yellow:Yellow|green:Green|cyan:Cyan|blue:Blue|magenta:Magenta"
	profiles = "power-saver:Power saver|balanced:Balanced|performance:Performance"
	filters  = "none:None|greyscale:Greyscale|deuteranopia:Green weak (deuteranopia)|protanopia:Red weak (protanopia)|tritanopia:Blue weak (tritanopia)|invert:Inverted colours"
	yourRows = []yours{
		{"accent", "Accent colour", "themes", "Colours and fonts", Pick, accents, "The theme's own accent, or another of its colours, for borders and highlights", nil},
		{"font", "Font", "themes", "Colours and fonts", Pick, "", "The terminal's, the bar's and the lock screen's font", nil},
		{"ui_font", "Interface font", "themes", "Colours and fonts", Pick, "", "The font of GTK and KDE apps", nil},
		{"glass", "Glass", "themes", "Windows", Toggle, "", "See-through terminals and panels, with the wallpaper blurred behind them", nil},
		{"style", "Window style", "themes", "Windows", Pick, "sharp:Sharp|soft:Soft", "Sharp corners with thin gaps, or soft rounded corners for windows, popups and the lock screen", nil},
		{"apps", "GTK and KDE apps", "themes", "Apps", Pick, "follow:Follow the theme|off:Keep their own look", "Apps follow the theme, keep one theme whatever the others are, or keep their own look", nil},
		{"editors", "Editors follow the theme", "themes", "Apps", Toggle, "", "VS Code, VSCodium and Zed switch with the theme, unless you picked another theme in them", nil},
		{"auto", "Automatic light and dark", "themes", "Light and dark", Pick, "off:Off|sun:By the sun|times:At set times", "The light theme by day and the dark one at night", nil},
		{"light", "Light theme", "themes", "Light and dark", Pick, "", "The theme for the day", isnt("auto", "off")},
		{"dark", "Dark theme", "themes", "Light and dark", Pick, "", "The theme for the night", isnt("auto", "off")},
		{"light_at", "Light from", "themes", "Light and dark", Time, "", "When the light theme starts", is("auto", "times")},
		{"dark_at", "Dark from", "themes", "Light and dark", Time, "", "When the dark theme starts", is("auto", "times")},
		{"daily", "Theme of the day", "themes", "Light and dark", Toggle, "", "The next theme every morning at 6", is("auto", "off")},
		{"wallpapers", "Wallpapers to show", "wallpaper", "Pictures", Pick, "both:The theme's and yours|theme:The theme's|mine:Yours", "The pictures the picker, next, rotation and the workspaces use: the theme's, with yours for it; yours for every theme; or both", nil},
		{"rotate", "Change the wallpaper every", "wallpaper", "Changing", Pick, "0:Never|15:15 minutes|30:30 minutes|60:An hour|120:2 hours|240:4 hours", "Move to the next picture every so many minutes", nil},
		{"workspaces", "A wallpaper per workspace", "wallpaper", "Pictures", Toggle, "", "Each workspace shows its own picture", nil},
		{"day_wallpaper", "Wallpaper of the day", "wallpaper", "Changing", Pick, "off:Off|drawn:Drawn in the theme's colours|nasa:NASA's picture of the day|folder:From your folder", "A new wallpaper every day", nil},
		{"wallpaper_folder", "Wallpaper folder", "wallpaper", "Changing", Text, "", "The folder of pictures, or a file that lists them one per line, for the wallpaper of the day", is("day_wallpaper", "folder")},
		{"logo", "Live wallpaper", "wallpaper", "Live wallpapers", Pick, "off:Off|animated:Animated logo|life:Game of Life", "Drawn in the theme's colours behind your windows; it rests while a window covers it, on battery and with the screens off", nil},
		{"video", "Video as the wallpaper", "wallpaper", "Live wallpapers", Text, "", "A video file played muted and looping, or off", nil},
		{"video_battery", "Video on battery", "wallpaper", "Live wallpapers", Pick, "pause:Pause|play:Play", "Whether a video wallpaper plays on battery", isnt("video", "off")},
		{"corners", "Rounded screen corners", "display", "Scale and shape", Toggle, "", "Round the corners of the screen, like a phone's", nil},
		{"night", "Night light", "display", "Night light", Pick, "off:Off|on:On|times:At night", "A warmer screen: never, always, or for the night hours", nil},
		{"night_from", "Night light from", "display", "Night light", Time, "", "When the screen warms up", is("night", "times")},
		{"night_to", "Night light until", "display", "Night light", Time, "", "When the screen goes back to normal", is("night", "times")},
		{"night_warmth", "Night light warmth", "display", "Night light", Slide, "", "How warm, in kelvin: lower is warmer", nil},
		{"screens_wait", "Keep a new layout for", "display", "Screens", Pick, "10:10 seconds|15:15 seconds|30:30 seconds|60:A minute", "How long Keep this layout? waits for an answer before the old layout comes back", nil},
		{"presentation", "Presentation mode", "display", "Presenting", Toggle, "", "No notifications, stay awake, a bigger cursor and the laptop mirrored to other screens, until you log in again", nil},
		{"chime", "Startup sound", "sound", "Startup sound", Toggle, "", "A short chime when you log in, or ~/.config/solanum/startup.ogg if it's there", nil},
		{"power", "Power saver on battery", "power", "Battery", Toggle, "", "Switch the power profile when you plug in or unplug", nil},
		{"power_ac", "On the charger", "power", "Battery", Pick, profiles, "The power profile on the charger", is("power", "yes")},
		{"power_battery", "On battery", "power", "Battery", Pick, profiles, "The power profile on battery", is("power", "yes")},
		{"lock_media", "Now playing on the lock screen", "lock", "Locking", Toggle, "", "The song or video that's playing, under the password field", nil},
		{"lock_battery", "Battery on the lock screen", "lock", "Locking", Toggle, "", "How full the battery is, in a corner", nil},
		{"trackpoint_speed", "TrackPoint speed", "mouse", "TrackPoint", Slide, "", "From slow (-1) to fast (1)", nil},
		{"trackpoint_accel", "TrackPoint acceleration", "mouse", "TrackPoint", Pick, "adaptive:Adaptive|flat:Flat", "Adaptive speeds up with faster moves; flat keeps one speed", nil},
		{"trackpoint_scroll", "Scroll with the middle button", "mouse", "TrackPoint", Toggle, "", "Hold the middle button and move the TrackPoint to scroll", nil},
		{"weather", "Weather in the bar", "bar", "Weather", Toggle, "", "From Open-Meteo for your location, which it sends there", nil},
		{"weather_units", "Weather units", "bar", "Weather", Pick, "celsius:Celsius and km/h|fahrenheit:Fahrenheit and mph", "The units the weather shows in", is("weather", "yes")},
		{"zen", "Zen mode", "bar", "Layout", Toggle, "", "No bar and no gaps", nil},
		{"quiet", "Quiet hours", "notifications", "Quiet hours", Toggle, "", "Do not disturb every night, for the hours below", nil},
		{"quiet_from", "Quiet from", "notifications", "Quiet hours", Time, "", "When notifications go quiet", is("quiet", "yes")},
		{"quiet_to", "Quiet until", "notifications", "Quiet hours", Time, "", "When notifications come back", is("quiet", "yes")},
		{"location", "Location", "time", "Location", Text, "", "zone for your time zone's main city, or LATITUDE,LONGITUDE; the sun and the weather use it", nil},
		{"text", "Text size", "accessibility", "Seeing", Pick, "normal:Normal|large:Large|larger:Larger", "Text everywhere a quarter (large) or half (larger) bigger", nil},
		{"zoom_follow", "Zoom follows the pointer", "accessibility", "Seeing", Pick, "edges:At the screen's edges|centre:In the middle", "Super with the two keys right of 0, or with the keypad's plus and minus, zooms in and out; Super+Alt+0 goes back to normal and takes the colour filter away too; the zoomed view moves when the pointer reaches its edge, or keeps the pointer in the middle", nil},
		{"filter", "Colour filter", "accessibility", "Seeing", Pick, filters, "Colours to tell apart or to rest your eyes; screenshots and recordings show them too", nil},
		{"animations", "Animations", "accessibility", "Motion", Toggle, "", "Windows, workspaces, popups and apps move as they open and close; off is reduced motion", nil},
		{"reader", "Screen reader", "accessibility", "Screen reader", Toggle, "", "Orca reads apps aloud, now and at every login; it can't read the bar, the menu or the launcher", nil},
	}
)

func is(key, value string) func(func(string) string) bool {
	return func(get func(string) string) bool { return get(key) == value }
}

func isnt(key, value string) func(func(string) string) bool {
	return func(get func(string) string) bool { return get(key) != value }
}

func choicesOf(text string) []Choice {
	var found []Choice
	for _, pair := range strings.Split(text, "|") {
		if value, label, ok := strings.Cut(pair, ":"); ok {
			found = append(found, Choice{value, label})
		}
	}
	return found
}

func named(values []string) []Choice {
	found := make([]Choice, 0, len(values))
	for _, value := range values {
		found = append(found, Choice{value, value})
	}
	return found
}

func lookCommand(key string) string {
	return strings.ReplaceAll(key, "_", "-")
}

func lookArgs(key, value string, get func(string) string) []string {
	pairs := map[string][3]string{
		"light_at": {"times", "", "dark_at"}, "dark_at": {"times", "light_at", ""},
		"quiet_from": {"quiet-hours", "", "quiet_to"}, "quiet_to": {"quiet-hours", "quiet_from", ""},
		"night_from": {"night-hours", "", "night_to"}, "night_to": {"night-hours", "night_from", ""},
		"power_ac": {"power-profiles", "", "power_battery"}, "power_battery": {"power-profiles", "power_ac", ""},
	}
	if pair, found := pairs[key]; found {
		if pair[1] != "" {
			return []string{pair[0], get(pair[1]), value}
		}
		return []string{pair[0], value, get(pair[2])}
	}
	switch key {
	case "day_wallpaper":
		return []string{"wallpaper-day", value}
	case "wallpaper_folder":
		return []string{"wallpaper-day", "folder", value}
	case "rotate":
		if value == "0" {
			value = "off"
		}
	}
	return []string{lookCommand(key), onOff(value)}
}

func onOff(value string) string {
	switch value {
	case "yes":
		return "on"
	case "no":
		return "off"
	}
	return value
}
