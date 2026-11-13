package look

import (
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
)

func OwnBackgrounds(theme string) string {
	return fsutil.Followed(filepath.Join(env.AppConfig(), "backgrounds", theme))
}
