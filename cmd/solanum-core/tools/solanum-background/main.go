package background

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/burst"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/hypr"
	"github.com/Ahmadlolbit/solanumos/internal/look"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/wallpaper"
)

const usage = `Usage: solanum-background <file>
       solanum-background next
       solanum-background draw
       solanum-background keep
       solanum-background logo
       solanum-background add [--every] [FILE ...]
       solanum-background remove FILE
       solanum-background --list
       solanum-background --current
       solanum-background --frames
       solanum-background workspaces
       solanum-background workspace MONITOR WORKSPACE

Sets the wallpaper. Backgrounds come from the current theme, from yours for
that theme in ~/.config/solanum/backgrounds/<theme>, and from yours for every
theme, the pictures in ~/.config/solanum/backgrounds itself; solanum-look
wallpapers picks the theme's, yours or both. add copies pictures into yours
for the current theme, or for every theme with --every; without files it
opens the file chooser. remove moves one of yours to the Trash. draw draws a
new one in the theme's colours, in a style and with a seed picked at random,
and keeps only the last one drawn; keep keeps the one drawn last in yours for
the current theme. logo shows the burst logo, animated by the bar in the
theme's colours, over the still logo; it rests while a window is on the
screen or on battery, and stays across themes until another wallpaper is
set. --frames prints its frames. With a wallpaper per workspace on
(solanum-look workspaces on), each workspace shows the next one along,
workspaces shows each monitor's, and workspace shows one monitor's for the
workspace it just switched to.
`

func state() string {
	return env.ThemeState()
}

func link() string {
	return filepath.Join(state(), "background")
}

func current() string {
	target, _ := os.Readlink(link())
	return target
}

func backgrounds() []string {
	return look.Open().Backgrounds(themeName())
}

func next() (string, bool) {
	found := backgrounds()
	if len(found) == 0 {
		return "", false
	}
	return found[(slices.Index(found, current())+1)%len(found)], true
}

func perWorkspace() bool {
	settings := look.Open()
	return settings.Get("workspaces") == "yes" && settings.Get("day_wallpaper") == "off"
}

func inHyprland() bool {
	return os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != ""
}

func workspaces() bool {
	if !inHyprland() {
		return false
	}
	found := backgrounds()
	if len(found) == 0 {
		return current() == "" || overIPC(current())
	}
	monitors, ok := hypr.QuietMonitors()
	if !ok {
		return false
	}
	for _, monitor := range monitors {
		if monitor.ActiveWorkspace.ID == nil {
			continue
		}
		if !hypr.Wallpaper(monitor.Name, forWorkspace(found, *monitor.ActiveWorkspace.ID)) {
			return false
		}
	}
	return true
}

func forWorkspace(found []string, id int) string {
	offset := slices.Index(found, current())
	if offset < 0 && current() != "" {
		return current()
	}
	return found[(max(id, 1)-1+max(offset, 0))%len(found)]
}

func oneWorkspace(monitor, workspace string) int {
	id, err := strconv.Atoi(workspace)
	if err != nil || !hypr.ValidMonitor(monitor) {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	if !inHyprland() || !perWorkspace() {
		return 0
	}
	found := backgrounds()
	if len(found) == 0 || hypr.Wallpaper(monitor, forWorkspace(found, id)) {
		return 0
	}
	return 1
}

func overIPC(file string) bool {
	if !inHyprland() {
		return false
	}
	monitors, ok := hypr.QuietMonitors()
	if !ok || len(monitors) == 0 {
		return false
	}
	for _, monitor := range monitors {
		if !hypr.Wallpaper(monitor.Name, file) {
			return false
		}
	}
	return true
}

func restartPaper() {
	if !inHyprland() {
		return
	}
	if run.Within(run.Quick, "pkill", "-x", "hyprpaper") == 0 {
		run.Quiet(3*time.Second, "pidwait", "-x", "hyprpaper")
	}
	run.Quiet(run.Quick, "setsid", "-f", "solanum-start", "hyprpaper")
}

func apply(file string) int {
	if !fsutil.IsFile(file) {
		fmt.Fprintf(os.Stderr, "no such background: %s\n", file)
		return 1
	}
	err := os.MkdirAll(state(), 0o777)
	if err == nil {
		err = atomic.Symlink(file, link())
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	logoFollows(file)
	if perWorkspace() {
		if workspaces() {
			return 0
		}
	} else if overIPC(file) {
		return 0
	}
	restartPaper()
	return 0
}

func reason(err error) string {
	var number syscall.Errno
	if !errors.As(err, &number) {
		return err.Error()
	}
	text := number.Error()
	return strings.ToUpper(text[:1]) + text[1:]
}

func set(argument string) int {
	if listed, err := filepath.Abs(argument); err == nil && slices.Contains(backgrounds(), listed) {
		return apply(listed)
	}
	file, err := fsutil.RealPath(argument)
	if err != nil {
		fmt.Fprintf(os.Stderr, "realpath: %s: %s\n", argument, reason(err))
	}
	return apply(file)
}

func moveOn() int {
	file, ok := next()
	if !ok {
		fmt.Fprintln(os.Stderr, "the current theme has no backgrounds")
		return 1
	}
	return apply(file)
}

func logoFollows(file string) {
	settings := look.Open()
	if wallpaper.IsLogo(file) || settings.Get("logo") == "off" {
		return
	}
	if err := settings.Put("logo", "off"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	look.ReloadShell()
}

func logoOn() int {
	still, err := wallpaper.LogoStill(wallpaper.CurrentColors())
	if err == nil {
		err = look.Open().Put("logo", "animated")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	code := apply(still)
	if code == 0 {
		wallpaper.Tidy(still)
	}
	look.ReloadShell()
	return code
}

func drawAnother() int {
	drawn, err := wallpaper.Another(wallpaper.CurrentColors())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	code := apply(drawn)
	if code == 0 {
		wallpaper.Tidy(drawn)
	}
	return code
}

func showCurrent() int {
	target, err := os.Readlink(link())
	if err != nil {
		return 1
	}
	fmt.Println(target)
	return 0
}

func showWorkspaces() int {
	if perWorkspace() && !workspaces() {
		return 1
	}
	return 0
}

func Main() {
	command := env.Arg(1)
	switch command {
	case "":
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	case "-h", "--help":
		fmt.Print(usage)
	case "--list":
		for _, file := range backgrounds() {
			fmt.Println(file)
		}
	case "--current":
		os.Exit(showCurrent())
	case "workspaces":
		os.Exit(showWorkspaces())
	case "workspace":
		os.Exit(oneWorkspace(env.Arg(2), env.Arg(3)))
	case "next":
		os.Exit(moveOn())
	case "draw":
		os.Exit(drawAnother())
	case "logo":
		os.Exit(logoOn())
	case "--frames":
		fmt.Print(burst.Flipbook())
	default:
		os.Exit(other(command))
	}
}

func other(command string) int {
	if yours, found := yourCommands[command]; found {
		return yours(os.Args[2:])
	}
	return set(command)
}
