package look

import (
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
)

func OwnBackgrounds(theme string) string {
	return fsutil.Followed(filepath.Join(env.AppConfig(), "backgrounds", theme))
}

func Everywhere() string {
	return fsutil.Followed(filepath.Join(env.AppConfig(), "backgrounds"))
}

func (store *Store) Backgrounds(theme string) []string {
	themes := fsutil.Images(filepath.Join(env.ThemeState(), "theme", "backgrounds"), OwnBackgrounds(theme))
	mine := fsutil.Images(Everywhere())
	switch store.Get("wallpapers") {
	case "theme":
		return themes
	case "mine":
		if len(mine) > 0 {
			return mine
		}
	}
	return append(themes, mine...)
}
