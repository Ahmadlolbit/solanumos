package background

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/cleanup"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/look"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
	"github.com/Ahmadlolbit/solanumos/internal/wallpaper"
)

const largestSoup = 1 << 20

var (
	liveStills   = map[string]func(string) bool{"animated": wallpaper.IsLogo, "life": wallpaper.IsLife}
	liveCommands = map[string]func() int{"logo": logoOn, "life": lifeOn, "off": liveOff}
)

func liveFollows(file string) {
	settings := look.Open()
	still, live := liveStills[settings.Get("logo")]
	if !live || still(file) {
		return
	}
	if err := settings.Put("logo", "off"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	look.ReloadShell()
}

func liveOn(kind string, draw func(theme.Colors) (string, error)) int {
	still, err := draw(wallpaper.CurrentColors())
	if err == nil {
		err = look.Open().Put("logo", kind)
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

func logoOn() int {
	return liveOn("animated", wallpaper.LogoStill)
}

func lifeOn() int {
	return liveOn("life", func(colors theme.Colors) (string, error) {
		return wallpaper.LifeStill(colors, wallpaper.NewSeed())
	})
}

func liveOff() int {
	settings := look.Open()
	if _, live := liveStills[settings.Get("logo")]; !live {
		fmt.Println("The live wallpaper is already off")
		return 0
	}
	if err := settings.Put("logo", "off"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer look.ReloadShell()
	if found := backgrounds(); len(found) > 0 {
		return apply(found[0])
	}
	return drawAnother()
}

func soupSize(args []string) (int, int, bool) {
	if len(args) != 2 {
		return 0, 0, false
	}
	cols, wide := strconv.Atoi(args[0])
	rows, high := strconv.Atoi(args[1])
	fits := cols > 0 && rows > 0 && cols <= largestSoup && rows <= largestSoup && cols*rows <= largestSoup
	return cols, rows, wide == nil && high == nil && fits
}

func plainName(name string) bool {
	return name != "" && len(name) <= 64 && strings.Trim(name, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.-") == "" && !strings.HasPrefix(name, ".")
}

func play(args []string) int {
	cols, rows, ok := 0, 0, len(args) == 3 && plainName(args[0])
	if ok {
		cols, rows, ok = soupSize(args[1:])
	}
	if !ok {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	seed, found := wallpaper.LifeSeed(link())
	if !found {
		seed = wallpaper.NewSeed()
	}
	prefix := framePrefix(args[0])
	defer forgetFrames(prefix)
	defer cleanup.OnSignal(func() { forgetFrames(prefix) })()
	fmt.Println(prefix)
	return playing(wallpaper.NewGame(seed, cols, rows), prefix, os.Stdin)
}

func framePrefix(name string) string {
	return filepath.Join(env.Or("XDG_RUNTIME_DIR", state()), fmt.Sprintf("solanum-life-%s-%d", name, os.Getpid()))
}

func frameFile(prefix string, shown int) string {
	return fmt.Sprintf("%s-%d.ppm", prefix, shown%2)
}

func forgetFrames(prefix string) {
	os.Remove(frameFile(prefix, 0))
	os.Remove(frameFile(prefix, 1))
}

func playing(game *wallpaper.Game, prefix string, input io.Reader) int {
	colors := wallpaper.CurrentColors()
	lines := bufio.NewScanner(input)
	for shown := 0; ; shown++ {
		os.Remove(frameFile(prefix, shown))
		if err := atomic.Scratch(frameFile(prefix, shown), game.Frame(colors), 0o600); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Println(shown, game.Generation, game.Soups, game.Population)
		if !lines.Scan() {
			return 0
		}
		switch lines.Text() {
		case "step":
			game.Step()
		case "sow":
			game.Sow(wallpaper.NewSeed())
		case "paint":
			colors = wallpaper.CurrentColors()
		}
	}
}
