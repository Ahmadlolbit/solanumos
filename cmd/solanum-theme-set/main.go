package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/cleanup"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/look"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
	"github.com/Ahmadlolbit/solanumos/internal/wallpaper"
)

const usage = `Usage: solanum-theme-set <theme>
       solanum-theme-set --list
       solanum-theme-set --previews
       solanum-theme-set --current
       solanum-theme-set --editors
       solanum-theme-set --make NAME BACKGROUND FOREGROUND ACCENT
       solanum-theme-set --copy THEME NAME [KEY=#RRGGBB ...]
       solanum-theme-set --change NAME KEY=#RRGGBB ...
       solanum-theme-set --palette THEME [KEY=#RRGGBB ...]
       solanum-theme-set --palette BACKGROUND FOREGROUND ACCENT

Applies a theme to Hyprland, the bar, kitty, the launcher, notifications, the
lock screen, btop, Neovim, VS Code, VSCodium, Zed, Firefox, GTK and KDE
apps, and the wallpaper. Themes come from /usr/share/solanum/themes and
~/.config/solanum/themes. Before a theme applies, its colours and files are
checked: what's missing is worked out, text too faint to read is made
readable, and files that can't be read are left out, each said in one line.
A theme that can't be read at all is never applied. --editors gives the
editors and Firefox's profiles the current theme again, for one installed or
made since. --make makes a theme of your own from a background, a foreground
and an accent: the rest of its colours are worked out, dark or light follows
from the background, and it gets a preview and contours as its first
background. --copy makes your own copy of a theme with any colours changed;
the theme you started from stays as it was. --change changes colours of a
theme of your own, and applies it again if it's in use. --palette prints a
theme's colours as JSON, each with its contrast, and what a switch would fix.
`

const (
	denied     = " alacritty.toml fastfetch.jsonc firefox.css foot.ini ghostty.conf hyprlock.conf kde.colors kitty.conf mako.ini vscode.json "
	wallpapers = "--wallpapers-soon"
)

var (
	ownFiles     = []string{"btop.theme", "chromium.theme", "colors.toml", "gtk3.css", "gtk4.css", "icons.theme", "light.mode", "look.toml", "walker.css"}
	previewKinds = []string{"jpg", "jpeg", "png", "webp"}
	systemRoot   = env.ThemeRoot()
	iconName     = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	settings     = look.Open()
)

func state() string {
	return env.ThemeState()
}

func userThemes() string {
	return filepath.Join(env.AppConfig(), "themes")
}

func themesIn(dir string) []string {
	var names []string
	for _, entry := range fsutil.Entries(fsutil.Followed(dir)) {
		if (entry.IsDir() || entry.Type()&os.ModeSymlink != 0) && theme.ValidName(entry.Name()) {
			names = append(names, entry.Name())
		}
	}
	return names
}

func themeNames() []string {
	names := append(themesIn(filepath.Join(systemRoot, "themes")), themesIn(userThemes())...)
	return slices.Compact(run.Sorted(names))
}

func previewOf(name string) string {
	own, shipped := filepath.Join(userThemes(), name), filepath.Join(systemRoot, "themes", name)
	for _, dir := range []string{own, shipped} {
		for _, kind := range previewKinds {
			if file := filepath.Join(dir, "preview."+kind); wallpaper.IsPicture(file) {
				return file
			}
		}
	}
	for _, dir := range []string{own, shipped} {
		for _, file := range fsutil.Images(fsutil.Followed(filepath.Join(dir, "backgrounds"))) {
			if wallpaper.IsPicture(file) {
				return file
			}
		}
	}
	return ""
}

func overrideAccent(colors theme.Colors) {
	if accent, ok := colors.Accent(settings.Get("accent")); ok {
		colors["theme_accent"], colors["accent"] = colors["accent"], accent
	}
}

func glassSettings(colors theme.Colors) {
	if settings.Get("glass") == "yes" {
		colors["glass"], colors["opacity"], colors["panel_alpha"] = "yes", "0.88", "cc"
		return
	}
	colors["glass"], colors["opacity"], colors["panel_alpha"] = "no", "1.0", "ee"
}

func styleSettings(colors theme.Colors) {
	colors["rounding"] = "0"
	if settings.Get("style") == "soft" {
		colors["rounding"] = "14"
	}
}

func chooseFont(colors theme.Colors) {
	colors["font"] = settings.Get("font")
	if !theme.ValidFont(colors["font"]) {
		colors["font"] = theme.DefaultFont
	}
}

func themeColors(dir string) (theme.Colors, []string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "colors.toml"))
	if err != nil {
		return nil, nil, err
	}
	return theme.Checked(string(data), fsutil.IsFile(filepath.Join(dir, "light.mode")))
}

func readColors(next string) (theme.Colors, []string, error) {
	colors, fixes, err := themeColors(next)
	if err != nil {
		return nil, nil, err
	}
	overrideAccent(colors)
	fixes = append(fixes, colors.ReadableAccent()...)
	if err := atomic.Install(filepath.Join(next, "colors.toml"), []byte(colors.Canonical()), 0o644); err != nil {
		return nil, nil, err
	}
	chooseFont(colors)
	glassSettings(colors)
	styleSettings(colors)
	return colors, fixes, nil
}

func templates(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var paths []string
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if !strings.HasPrefix(entry.Name(), ".") && strings.HasSuffix(entry.Name(), ".tpl") && fsutil.IsFile(path) {
			paths = append(paths, path)
		}
	}
	return paths
}

func renderOne(template, target string, replacer *strings.Replacer, scale float64) error {
	data, err := os.ReadFile(template)
	if err != nil {
		missed(appOf(filepath.Base(target)), err)
		return nil
	}
	if err := atomic.Stage(target, []byte(theme.Sized(replacer.Replace(string(data)), scale)), 0o666); err != nil {
		return fmt.Errorf("the theme file for %s can't be written (%s)", appOf(filepath.Base(target)), reason(err))
	}
	return nil
}

func render(dir string, colors, apps theme.Colors) error {
	replacers := map[bool]*strings.Replacer{false: colors.Replacer(), true: apps.Replacer()}
	scale := theme.TextScale(settings.Get("text"))
	found := append(templates(filepath.Join(env.AppConfig(), "themed")), templates(filepath.Join(systemRoot, "themed"))...)
	for _, template := range found {
		base := strings.TrimSuffix(filepath.Base(template), ".tpl")
		if target := filepath.Join(dir, base); !fsutil.Exists(target) {
			if err := renderOne(template, target, replacers[slices.Contains(toolkitFiles, base)], scale); err != nil {
				return err
			}
		}
	}
	return nil
}

func swap(next, target string) error {
	old := target + ".old"
	if err := os.RemoveAll(old); err != nil {
		return err
	}
	if err := os.Rename(target, old); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.Rename(next, target); err != nil {
		os.Rename(old, target)
		return err
	}
	return os.RemoveAll(old)
}

func symlink(app, target, link string) {
	if linked, err := os.Readlink(link); err == nil && linked == target {
		return
	}
	err := os.MkdirAll(filepath.Dir(link), 0o777)
	if err == nil {
		err = atomic.Symlink(target, link)
	}
	missed(app, err)
}

func drawnBackground() string {
	return filepath.Join(state(), "drawn.jpg")
}

func drawBackground(name string, colors theme.Colors, link string) {
	data, err := wallpaper.JPEG(wallpaper.Contours(colors, wallpaper.Seed(name), wallpaper.Width, wallpaper.Height))
	if err == nil {
		err = atomic.Write(drawnBackground(), data, 0o666)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	symlink("the wallpaper", drawnBackground(), link)
}

type paper int

const (
	paperKept paper = iota
	paperShown
	paperReloaded
)

func backgroundLink() string {
	return filepath.Join(state(), "background")
}

func shownBackground() string {
	target, _ := os.Readlink(backgroundLink())
	return target
}

func paperChange(otherTheme bool, before string) paper {
	switch {
	case shownBackground() != before:
		return paperShown
	case otherTheme:
		return paperReloaded
	}
	return paperKept
}

func liveStill(kind string, colors theme.Colors, link string) (string, error) {
	if kind == "animated" {
		return wallpaper.LogoStill(colors)
	}
	seed, ok := wallpaper.LifeSeed(link)
	if !ok {
		seed = wallpaper.NewSeed()
	}
	return wallpaper.LifeStill(colors, seed)
}

func liveShown(colors theme.Colors, link string) bool {
	unlock, err := look.Hold(state())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	defer unlock()
	kind := look.Open().Get("logo")
	if kind != "animated" && kind != "life" {
		return false
	}
	still, err := liveStill(kind, colors, link)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	symlink("the wallpaper", still, link)
	wallpaper.Tidy(still)
	return true
}

func pickBackground(name, previous string, colors theme.Colors) {
	link := backgroundLink()
	if liveShown(colors, link) || (previous == name && fsutil.Exists(link)) {
		return
	}
	found := settings.Backgrounds(name)
	target, _ := os.Readlink(link)
	switch {
	case settings.Get("wallpapers") == "mine" && slices.Contains(found, target):
		return
	case len(found) > 0:
		symlink("the wallpaper", found[0], link)
	case !fsutil.Exists(link) || target == drawnBackground():
		drawBackground(name, colors, link)
	}
}

func linkApps() {
	current := filepath.Join(state(), "theme")
	if file := filepath.Join(current, "walker.css"); fsutil.IsFile(file) {
		symlink("the launcher", file, filepath.Join(env.ConfigHome(), "walker", "themes", "solanum", "style.css"))
	}
	if file := filepath.Join(current, "btop.theme"); fsutil.IsFile(file) {
		symlink("btop", file, filepath.Join(env.ConfigHome(), "btop", "themes", "solanum.theme"))
	}
	if file := filepath.Join(current, "fastfetch.jsonc"); fsutil.IsFile(file) {
		linkUnlessYours("fastfetch", file, filepath.Join(env.ConfigHome(), "fastfetch", "config.jsonc"))
	}
	linkNeovim(current)
	linkAlacritty(current)
	followFirefox(current)
	linkToolkits(current)
}

func desktop(action string, args ...string) {
	args = append([]string{action, "org.gnome.desktop.interface"}, args...)
	if os.Getenv("DBUS_SESSION_BUS_ADDRESS") != "" {
		run.Within(run.Quick, "gsettings", args...)
		return
	}
	run.Quiet(run.Quick, "dbus-run-session", append([]string{"--", "gsettings"}, args...)...)
}

func desktopValues() map[string]string {
	name, args := "gsettings", []string{"list-recursively", "org.gnome.desktop.interface"}
	if os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
		name, args = "dbus-run-session", append([]string{"--", "gsettings"}, args...)
	}
	output, _ := run.CaptureQuiet(run.Quick, name, args...)
	values := map[string]string{}
	for _, line := range strings.Split(output, "\n") {
		if fields := strings.SplitN(line, " ", 3); len(fields) == 3 && fields[0] == "org.gnome.desktop.interface" {
			values[fields[1]] = strings.TrimSpace(fields[2])
		}
	}
	return values
}

func change(current map[string]string, key, value string) {
	if current[key] != "'"+value+"'" {
		desktop("set", key, value)
	}
}

func scaleText(current map[string]string, scale float64) {
	factor, err := strconv.ParseFloat(current["text-scaling-factor"], 64)
	switch {
	case scale != 1 && (err != nil || factor != scale):
		desktop("set", "text-scaling-factor", strconv.FormatFloat(scale, 'f', -1, 64))
	case scale == 1 && err == nil && slices.Contains([]float64{theme.TextScale("large"), theme.TextScale("larger")}, factor):
		desktop("reset", "text-scaling-factor")
	}
}

func iconInstalled(name string) bool {
	for _, dir := range []string{filepath.Join(env.DataHome(), "icons"), filepath.Join(env.Home(), ".icons"), "/usr/local/share/icons", "/usr/share/icons"} {
		if fsutil.IsDir(filepath.Join(dir, name)) {
			return true
		}
	}
	return false
}

func iconTheme(mode string) string {
	first, _, _ := strings.Cut(fsutil.Optional(filepath.Join(state(), "theme", "icons.theme")), "\n")
	icons := strings.TrimSpace(strings.TrimPrefix(first, "\ufeff"))
	if iconName.MatchString(icons) && iconInstalled(icons) {
		return icons
	}
	if mode == "dark" {
		return "Yaru-dark"
	}
	return "Yaru"
}

func desktopSettings(colors, apps theme.Colors) {
	if _, err := exec.LookPath("gsettings"); err != nil {
		return
	}
	current := desktopValues()
	mode := apps["mode"]
	if settings.Get("apps") == appsOff {
		releaseGTK(mode)
	} else {
		rememberGTK(current)
		change(current, "color-scheme", "prefer-"+mode)
		desktop("set", "gtk-theme", gtkTheme(mode))
		change(current, "icon-theme", iconTheme(mode))
	}
	change(current, "monospace-font-name", colors["font"]+" 11")
	if ui := settings.Get("ui_font"); ui == "default" {
		desktop("reset", "font-name")
	} else {
		change(current, "font-name", ui+" 11")
	}
	scaleText(current, theme.TextScale(settings.Get("text")))
	change(current, "cursor-theme", "Bibata-Modern-Ice")
}

func rememberPick(name, mode string) {
	if os.Getenv("SOLANUM_LOOK_AUTO") != "" || settings.Get("auto") == "off" {
		return
	}
	key := "dark"
	if mode == "light" {
		key = "light"
	}
	if err := settings.Put(key, name); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func prepare(name, next string) (theme.Colors, theme.Colors, []string, error) {
	notes, err := stage(name, next)
	if err != nil {
		return nil, nil, nil, err
	}
	if !fsutil.IsFile(filepath.Join(next, "colors.toml")) {
		os.RemoveAll(next)
		return nil, nil, nil, cantUse(notes)
	}
	colors, fixes, err := readColors(next)
	if err != nil {
		os.RemoveAll(next)
		return nil, nil, nil, err
	}
	fixes = append(checkFiles(name, next, colors["mode"], notes), fixes...)
	apps := appColors(colors)
	if err := render(next, colors, apps); err != nil {
		os.RemoveAll(next)
		return nil, nil, nil, err
	}
	if colors["glass"] != "yes" {
		os.Remove(filepath.Join(next, "glass.lua"))
	}
	return colors, apps, fixes, nil
}

func reason(err error) string {
	var path *fs.PathError
	if errors.As(err, &path) {
		return path.Err.Error()
	}
	return err.Error()
}

func refused(name string, err error) {
	stays := "Nothing changed."
	if previous, _ := fsutil.Text(filepath.Join(state(), "theme.name")); previous != "" {
		stays = theme.Pretty(previous) + " stays."
	}
	fmt.Fprintf(os.Stderr, "Couldn't apply %s: %s. %s\n", theme.Pretty(name), reason(err), stays)
}

func apply(name string) (paper, bool) {
	next := filepath.Join(filepath.Dir(state()), "next-theme")
	var moving sync.Mutex
	swapping := false
	defer cleanup.OnSignal(func() {
		moving.Lock()
		if !swapping {
			os.RemoveAll(next)
		}
	})()
	previous, _ := fsutil.Text(filepath.Join(state(), "theme.name"))
	before := shownBackground()
	colors, apps, fixes, err := prepare(name, next)
	if err == nil {
		err = atomic.SyncFolders(next)
	}
	if err != nil {
		refused(name, err)
		return paperKept, false
	}
	moving.Lock()
	swapping = true
	moving.Unlock()
	if err := swap(next, filepath.Join(state(), "theme")); err != nil {
		refused(name, err)
		return paperKept, false
	}
	writeChanged(filepath.Join(state(), "theme.name"), []byte(name+"\n"))
	pickBackground(name, previous, colors)
	linkApps()
	desktopSettings(colors, apps)
	forgetOldNames()
	rememberPick(name, colors["mode"])
	if len(fixes) > 0 {
		fmt.Println(strings.Join(fixes, "\n"))
	}
	return paperChange(previous != name, before), true
}

func detached(args ...string) {
	run.Quiet(run.Quick, "setsid", append([]string{"-f"}, args...)...)
}

func restartPaper() {
	if run.Within(run.Quick, "pkill", "-x", "hyprpaper") == 0 {
		run.Quiet(3*time.Second, "pidwait", "-x", "hyprpaper")
	}
	detached("solanum-start", "hyprpaper")
	if self, err := os.Executable(); err == nil {
		detached(self, wallpapers)
	}
}

func refreshPaper(change paper) {
	switch change {
	case paperShown:
		run.Within(run.Quick, "solanum-background", shownBackground())
	case paperReloaded:
		restartPaper()
	}
}

func retint(change paper) {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
		return
	}
	run.Quiet(run.Quick, "hyprctl", "reload")
	refreshPaper(change)
	run.Quiet(run.Quick, "makoctl", "reload")
	tellKDEApps()
	look.ReloadShell()
	run.Within(run.Quick, "pkill", "-USR1", "-x", "kitty")
	run.Within(run.Quick, "pkill", "-USR2", "-x", "btop")
	reloadNeovim()
	if run.Within(run.Quick, "pkill", "-x", "walker") == 0 {
		detached("solanum-start", "walker")
	}
}

func lock() (*os.File, error) {
	if err := os.MkdirAll(state(), 0o777); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(filepath.Dir(state()), "theme.lock"), os.O_RDONLY|os.O_CREATE, 0o666)
	if err != nil {
		return nil, err
	}
	for {
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
		if !errors.Is(err, syscall.EINTR) {
			return file, err
		}
	}
}

func known(argument, name string) bool {
	if theme.ValidName(name) && (fsutil.IsDir(filepath.Join(systemRoot, "themes", name)) || fsutil.IsDir(filepath.Join(userThemes(), name))) {
		return true
	}
	if own := filepath.Join(userThemes(), name); theme.ValidName(name) && fsutil.IsLink(own) {
		refused(name, errors.New(nowhere(own)))
		return false
	}
	fmt.Fprintf(os.Stderr, "no theme named %s; see solanum-theme-set --list\n", argument)
	return false
}

func set(argument string) int {
	name := theme.Slug(argument)
	if !known(argument, name) {
		return 1
	}
	held, err := lock()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	change, ok := apply(name)
	if !ok {
		return 1
	}
	held.Close()
	retint(change)
	return refreshEditors()
}

func list() {
	for _, name := range themeNames() {
		fmt.Println(theme.Pretty(name))
	}
}

func previews() {
	for _, name := range themeNames() {
		fmt.Printf("%s\t%s\t%s\n", name, theme.Pretty(name), previewOf(name))
	}
}

func current() int {
	file := filepath.Join(state(), "theme.name")
	if !fsutil.IsFile(file) {
		return 1
	}
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Print(theme.Pretty(string(data)))
	return 0
}

func main() {
	argument := env.Arg(1)
	switch argument {
	case "":
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	case "-h", "--help":
		fmt.Print(usage)
	case "--list":
		list()
	case "--previews":
		previews()
	case "--current":
		os.Exit(current())
	case "--editors":
		os.Exit(refreshEditors())
	case wallpapers:
		time.Sleep(time.Second)
		os.Exit(run.Report(run.Exec("solanum-background", "workspaces")))
	default:
		if command, found := ownCommands[argument]; found {
			os.Exit(command(os.Args[2:]))
		}
		os.Exit(set(argument))
	}
}
