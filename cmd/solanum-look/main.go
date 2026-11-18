package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/look"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

const usage = `Usage: solanum-look
       solanum-look accent [theme|red|orange|yellow|green|cyan|blue|magenta|#rrggbb]
       solanum-look font [FAMILY|default]
       solanum-look ui-font [FAMILY|default]
       solanum-look text [normal|large|larger]
       solanum-look reader [on|off|toggle]
       solanum-look zoom [in|out|reset|toggle]
       solanum-look zoom-follow [edges|centre]
       solanum-look filter [none|greyscale|deuteranopia|protanopia|tritanopia|invert]
       solanum-look animations [on|off|toggle]
       solanum-look chime [on|off|toggle|play]
       solanum-look glass [on|off|toggle]
       solanum-look corners [on|off|toggle]
       solanum-look apps [follow|off|THEME]
       solanum-look editors [on|off|toggle]
       solanum-look auto [off|sun|times|toggle]
       solanum-look light|dark [THEME]
       solanum-look times [LIGHT_FROM DARK_FROM]
       solanum-look daily [on|off|toggle]
       solanum-look sun [DATE]
       solanum-look location [LATITUDE,LONGITUDE|zone]
       solanum-look rotate [MINUTES|off|toggle]
       solanum-look wallpaper-day [off|drawn|nasa|folder [PATH]]
       solanum-look workspaces [on|off|toggle]
       solanum-look wallpapers [theme|mine|both]
       solanum-look bar [full|floating|minimal|vertical]
       solanum-look position [top|bottom]
       solanum-look style [sharp|soft]
       solanum-look video [FILE|off]
       solanum-look video-battery [pause|play]
       solanum-look weather [on|off|toggle]
       solanum-look weather-units [celsius|fahrenheit]
       solanum-look coordinates
       solanum-look quiet [on|off|toggle]
       solanum-look quiet-hours [FROM TO]
       solanum-look night [on|off|times|toggle]
       solanum-look night-hours [FROM TO]
       solanum-look night-warmth [KELVIN]
       solanum-look power [on|off|toggle]
       solanum-look power-profiles [CHARGER BATTERY]
       solanum-look trackpoint-speed [-1 to 1]
       solanum-look trackpoint-accel [adaptive|flat]
       solanum-look trackpoint-scroll [on|off|toggle]
       solanum-look lock-media [on|off|toggle]
       solanum-look lock-battery [on|off|toggle]
       solanum-look zen [on|off|toggle]
       solanum-look presentation [on|off|toggle]
       solanum-look screens [list|now|apply|forget|confirm|revert]
       solanum-look screens save [--try SECONDS] [SCREEN SETTING=VALUE...]...
       solanum-look screens-wait [SECONDS]
       solanum-look get SETTING

Shows and changes how the desktop looks. Your settings are kept in
~/.config/solanum/look.conf and win over the setup menu's in
/etc/solanum/look.conf. With no value, a command prints the setting.
An accent named after a colour uses that colour of each theme. font is
the terminal's, the bar's and the lock screen's; ui-font is the apps'.
text makes the text everywhere a quarter (large) or half (larger) bigger.
reader turns Orca, the screen reader, on or off, now and at every login.
It reads apps, but not the bar, the menu or the launcher, and Qt apps for
sure only after you log in again. chime plays a short chime when you log
in, or ~/.config/solanum/startup.ogg (.oga, .wav, .flac, .mp3) if it's
there; play plays it now.
zoom magnifies the screen around the pointer a quarter more or less at a
time, up to eight times, until you log out; reset goes back to normal and
toggle switches between normal and twice the size. zoom-follow moves the
zoomed view when the pointer reaches its edges, or keeps the pointer in the
middle. filter changes the screen's colours: greyscale, for colour
blindness (deuteranopia is green weak, protanopia red weak, tritanopia blue
weak), or inverted; screenshots and recordings show it too. animations off
is reduced motion: windows, workspaces, the bar's popups and GTK and KDE
apps stop animating, at once.
Glass makes terminals and panels see-through, with blur behind them.
corners rounds the screen's corners. apps sets what GTK and KDE apps do:
follow the theme, keep one theme whatever the others are, or stay as they
are (off). editors makes VS Code, VSCodium and Zed switch with the theme,
unless you picked another theme in them.
auto switches to the light theme by day and the dark one at night, at
sunrise and sunset (sun) or at set times like 07:00 and 19:00 (times).
While it's on, a theme you pick becomes the light or the dark one.
sun prints the day's sunrise and sunset where your time zone's main city
is, or at the location you set. rotate changes the wallpaper every so
many minutes; toggle switches between off and every 30 minutes.
wallpaper-day gives you a new wallpaper every day: drawn in the theme's
colours, NASA's picture of the day (with its credit; a day without a
picture is drawn), or one from a folder or a list of pictures, one per
line.
wallpapers picks the pictures the picker, next, rotation and the
workspaces use: the theme's, with yours for it in
~/.config/solanum/backgrounds/THEME; yours for every theme, the pictures in
~/.config/solanum/backgrounds; or both, the default.
workspaces gives each workspace its own wallpaper. bar and position set
the bar's layout and edge in ~/.config/solanum/shell.json. style makes
windows, popups and the lock screen sharp or soft. video plays a video file as
the wallpaper, muted and looping; on battery it pauses, unless
video-battery is play. weather shows the weather in the bar from
Open-Meteo for your location (it sends that location there); coordinates
prints the location the sun and the weather use. quiet turns do not
disturb on for the quiet hours (22:00 to 07:00 unless you set others).
night warms the screen: on, off, or for the night hours (times, 20:00 to
07:00 unless you set others); night-warmth is how warm, 4500 by default.
power switches the power profile when you plug in or unplug, to the two
in power-profiles: balanced on the charger and power-saver on battery unless
you set others. The trackpoint commands set a TrackPoint's speed, its
acceleration, and scrolling with the middle button. lock-media and
lock-battery show what's playing and the battery on the lock screen. zen
hides the bar and the gaps. presentation turns notifications off, stays
awake, shows a bigger cursor and mirrors the laptop to other screens,
until you log in again.
screens keeps a layout for each set of screens you plug in, such as the
desk, home or a projector, in ~/.config/solanum/screens.conf, and puts it
back whenever the same screens are plugged in again; screens without one
get the setup's scale. list shows the layouts, and now prints the screens
as JSON. save keeps the screens as they are, with changes such as DP-1
mode=2560x1440@144 position=1920x0 scale=1.25 rotation=90 on=no main=yes
mirror=eDP-1; the main screen gets workspace 1 and the cursor. With --try,
the old layout comes back after that many seconds unless confirm comes
first, and revert brings it back at once. apply puts the saved layout
back, and forget drops it. screens-wait is how long Settings waits for
Keep before the old layout comes back, 15 seconds unless you set another.
`

const (
	accents    = " theme red orange yellow green cyan blue magenta "
	readAccess = 4
)

var (
	settings   = look.Open()
	systemRoot = env.ThemeRoot()
)

func stateHome(parts ...string) string {
	return filepath.Join(append([]string{env.StateHome()}, parts...)...)
}

func currentTheme() string {
	name, _ := fsutil.Text(filepath.Join(env.ThemeState(), "theme.name"))
	return name
}

func put(key, value string) bool {
	if err := settings.Put(key, value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	return true
}

func show(key string) int {
	fmt.Println(settings.Get(key))
	return 0
}

func arg(args []string, index int) string {
	if index < len(args) {
		return args[index]
	}
	return ""
}

func editorsAgain() int {
	return run.Within(run.Slow, run.Sibling("solanum-theme-set"), "--editors")
}

func editorsFollow(args []string) int {
	return onOff{"editors", putThen("editors", "yes", editorsAgain), putThen("editors", "no", nothing)}.run(args)
}

func reapply() int {
	name := currentTheme()
	if name == "" {
		return 0
	}
	return run.Within(run.Slow, run.Sibling("solanum-theme-set"), name)
}

func accent(args []string) int {
	value := strings.ToLower(arg(args, 0))
	if value == "" {
		return show("accent")
	}
	if !strings.Contains(accents, " "+value+" ") && !theme.IsHex(value) {
		fmt.Fprintf(os.Stderr, "not an accent: %s (theme, a colour name or #rrggbb)\n", arg(args, 0))
		return 2
	}
	return putThen("accent", value, reapply)()
}

func fontInstalled(name string) bool {
	if _, err := exec.LookPath("fc-list"); err != nil {
		return true
	}
	family := ":family=" + strings.ReplaceAll(name, "-", `\-`)
	output, _ := run.CaptureQuiet(run.Slow, "fc-list", "--format", `%{family[0]}\n`, family)
	return output != ""
}

func font(key, value string) int {
	if value == "" {
		return show(key)
	}
	if value == "default" {
		value, _ = look.Default(key)
	} else if !theme.ValidFont(value) || !fontInstalled(value) {
		fmt.Fprintf(os.Stderr, "no installed font family named %s (see: fc-list : family)\n", value)
		return 2
	}
	if !put(key, value) {
		return 1
	}
	return reapply()
}

type onOff struct {
	key     string
	on, off func() int
}

func (item onOff) run(args []string) int {
	switch strings.ToLower(arg(args, 0)) {
	case "":
		return show(item.key)
	case "on", "yes":
		return item.on()
	case "off", "no":
		return item.off()
	case "toggle":
		if settings.Get(item.key) == "yes" {
			return item.off()
		}
		return item.on()
	}
	fmt.Fprintln(os.Stderr, item.key+" takes on, off or toggle")
	return 2
}

func putThen(key, value string, then func() int) func() int {
	return func() int {
		if !put(key, value) {
			return 1
		}
		return then()
	}
}

func nothing() int {
	return 0
}

func reload() int {
	look.ReloadShell()
	return 0
}

func glass(args []string) int {
	return onOff{"glass", putThen("glass", "yes", reapply), putThen("glass", "no", reapply)}.run(args)
}

func daily(args []string) int {
	fresh := func() int {
		remove(lookState("daily"))
		remove(lookState("daily-retry"))
		return 0
	}
	return onOff{"daily", putThen("daily", "yes", fresh), putThen("daily", "no", nothing)}.run(args)
}

func corners(args []string) int {
	return onOff{"corners", putThen("corners", "yes", reload), putThen("corners", "no", reload)}.run(args)
}

func windowStyle(args []string) int {
	switch value := strings.ToLower(arg(args, 0)); value {
	case "":
		return show("style")
	case "sharp", "soft":
		return putThen("style", value, reapply)()
	}
	fmt.Fprintln(os.Stderr, "style takes sharp or soft")
	return 2
}

func textSize(args []string) int {
	switch value := strings.ToLower(arg(args, 0)); value {
	case "":
		return show("text")
	case "normal", "large", "larger":
		return putThen("text", value, reapply)()
	}
	fmt.Fprintln(os.Stderr, "text takes normal, large or larger")
	return 2
}

func background(args ...string) int {
	return run.Within(run.Slow, run.Sibling("solanum-background"), args...)
}

func flatWallpaper() int {
	shown, err := os.Readlink(filepath.Join(env.ThemeState(), "background"))
	if err != nil || shown == "" {
		return 0
	}
	return background(shown)
}

func workspaces(args []string) int {
	on := putThen("workspaces", "yes", func() int {
		look.ReloadShell()
		return background("workspaces")
	})
	off := putThen("workspaces", "no", func() int {
		look.ReloadShell()
		return flatWallpaper()
	})
	return onOff{"workspaces", on, off}.run(args)
}

func video(args []string) int {
	value := arg(args, 0)
	if value == "" {
		return show("video")
	}
	if value != "off" {
		file, ok := videoFile(value)
		if !ok {
			fmt.Fprintf(os.Stderr, "not a video file this can play: %s (mp4, webm, mkv, mov or m4v)\n", value)
			return 2
		}
		value = file
	}
	return putThen("video", value, reload)()
}

func videoFile(path string) (string, bool) {
	file, err := fsutil.RealPath(path)
	if err != nil || !fsutil.IsFile(file) || syscall.Access(file, readAccess) != nil {
		return "", false
	}
	switch strings.ToLower(filepath.Ext(file)) {
	case ".mp4", ".webm", ".mkv", ".mov", ".m4v":
		return file, true
	}
	return "", false
}

func videoBattery(args []string) int {
	value := arg(args, 0)
	switch value {
	case "":
		return show("video_battery")
	case "pause", "play":
		return putThen("video_battery", value, reload)()
	}
	fmt.Fprintln(os.Stderr, "video-battery takes pause or play")
	return 2
}

func auto(args []string) int {
	value := strings.ToLower(arg(args, 0))
	switch value {
	case "":
		return show("auto")
	case "off":
		return putThen("auto", "off", nothing)()
	case "sun", "times":
		return putThen("auto", value, retick)()
	case "toggle":
		if settings.Get("auto") == "off" {
			return auto([]string{"sun"})
		}
		return auto([]string{"off"})
	}
	fmt.Fprintln(os.Stderr, "auto takes off, sun, times or toggle")
	return 2
}

func themeChoice(key, value string) int {
	if value == "" {
		return show(key)
	}
	name := theme.Slug(value)
	if !theme.ValidName(name) || !themeExists(name) {
		fmt.Fprintf(os.Stderr, "no theme named %s; see solanum-theme-set --list\n", value)
		return 2
	}
	return putThen(key, name, retick)()
}

func appsTheme(args []string) int {
	value := arg(args, 0)
	if value == "" {
		return show("apps")
	}
	name := theme.Slug(value)
	if name != "follow" && name != "off" && (!theme.ValidName(name) || !themeExists(name)) {
		fmt.Fprintf(os.Stderr, "apps takes follow, off or a theme; see solanum-theme-set --list\n")
		return 2
	}
	return putThen("apps", name, reapply)()
}

func themeExists(name string) bool {
	return fsutil.IsDir(filepath.Join(systemRoot, "themes", name)) || fsutil.IsDir(filepath.Join(env.AppConfig(), "themes", name))
}

func validTimes(args []string) bool {
	if len(args) != 2 {
		return false
	}
	_, first := look.Minutes(args[0])
	_, second := look.Minutes(args[1])
	return first && second
}

func times(args []string) int {
	if len(args) == 0 {
		fmt.Println(settings.Get("light_at"), settings.Get("dark_at"))
		return 0
	}
	if !validTimes(args) {
		fmt.Fprintln(os.Stderr, "times takes two times like 07:00 19:00: when light starts, then dark")
		return 2
	}
	if !put("light_at", args[0]) || !put("dark_at", args[1]) {
		return 1
	}
	return retick()
}

func location(args []string) int {
	value := arg(args, 0)
	if value == "" {
		return show("location")
	}
	if !look.ValidLocation(value) {
		fmt.Fprintln(os.Stderr, "a location is LATITUDE,LONGITUDE in degrees, like 33.31,44.36, or zone")
		return 2
	}
	return putThen("location", value, weatherHere)()
}

func remember(args []string) int {
	name := theme.Slug(arg(args, 0))
	if !theme.ValidName(name) {
		return 2
	}
	key := "dark"
	if arg(args, 1) == "light" {
		key = "light"
	}
	return putThen(key, name, nothing)()
}

func rotate(args []string) int {
	every := strings.ToLower(arg(args, 0))
	switch every {
	case "":
		return show("rotate")
	case "off":
		every = "0"
	case "toggle":
		every = "30"
		if settings.Get("rotate") != "0" {
			every = "0"
		}
	}
	minutes, ok := look.WholeNumber(every)
	if !ok || minutes > 1440 {
		fmt.Fprintln(os.Stderr, "rotate takes minutes from 0 to 1440, off or toggle")
		return 2
	}
	remove(lookState("rotated"))
	return putThen("rotate", fmt.Sprint(minutes), func() int {
		rotateTick()
		return 0
	})()
}

func wallpapers(args []string) int {
	switch value := arg(args, 0); value {
	case "":
		return show("wallpapers")
	case "theme", "mine", "both":
		if !put("wallpapers", value) {
			return 1
		}
		shown, _ := os.Readlink(filepath.Join(env.ThemeState(), "background"))
		if found := settings.Backgrounds(currentTheme()); len(found) > 0 && !slices.Contains(found, shown) {
			return background("next")
		}
		return 0
	}
	fmt.Fprintln(os.Stderr, "wallpapers takes theme, mine or both")
	return 2
}

func getSetting(args []string) int {
	key := arg(args, 0)
	if _, known := look.Default(key); !known {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	return show(key)
}

func all() int {
	for _, key := range look.Keys {
		fmt.Printf("%s=%s\n", key, settings.Get(key))
	}
	return 0
}

func commands() map[string]func([]string) int {
	return map[string]func([]string) int{
		"get": getSetting, "accent": accent, "glass": glass, "editors": editorsFollow, "corners": corners, "apps": appsTheme, "sun": sun, "auto": auto, "times": times, "daily": daily,
		"tick": tick, "remember": remember, "rotate": rotate, "wallpapers": wallpapers, "wallpaper-day": dayWallpaper, "workspaces": workspaces, "bar": barLayout, "position": barPosition, "style": windowStyle, "text": textSize, "reader": reader, "zoom": zoom, "zoom-follow": zoomFollow, "filter": filter, "animations": animations, "chime": chimeCommand, "video": video,
		"video-battery": videoBattery, "weather": weather, "weather-units": weatherUnits, "coordinates": showCoordinates,
		"quiet": quiet, "quiet-hours": quietHours, "night": night, "night-hours": nightHours, "night-warmth": nightWarmth, "location": location,
		"power": powerFollow, "power-profiles": powerProfiles, "power-apply": powerApply,
		"trackpoint-speed": trackpointSpeed, "trackpoint-accel": trackpointAccel, "trackpoint-scroll": trackpointScroll,
		"zen": zen, "presentation": presentation, "screens": screensCommand, "screens-wait": screensWait, "lock-media": lockMedia, "lock-battery": lockBattery, "lock-line": lockLine,
		"font":    func(args []string) int { return font("font", arg(args, 0)) },
		"ui-font": func(args []string) int { return font("ui_font", arg(args, 0)) },
		"light":   func(args []string) int { return themeChoice("light", arg(args, 0)) },
		"dark":    func(args []string) int { return themeChoice("dark", arg(args, 0)) },
	}
}

func main() {
	name := env.Arg(1)
	if name == "" {
		os.Exit(all())
	}
	if name == "-h" || name == "--help" {
		fmt.Print(usage)
		return
	}
	command, known := commands()[name]
	if !known {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	os.Exit(command(os.Args[2:]))
}
