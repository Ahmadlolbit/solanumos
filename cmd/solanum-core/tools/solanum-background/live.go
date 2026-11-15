package background

import (
	"fmt"
	"os"

	"github.com/Ahmadlolbit/solanumos/internal/look"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
	"github.com/Ahmadlolbit/solanumos/internal/wallpaper"
)

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
	if err := look.Open().Put("logo", "off"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer look.ReloadShell()
	if found := backgrounds(); len(found) > 0 {
		return apply(found[0])
	}
	return drawAnother()
}
