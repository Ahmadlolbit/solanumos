package followtheme

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
)

func menuFace() string {
	for _, size := range []string{"16", "20", "28"} {
		if !fsutil.IsFile(filepath.Join(grubDir, "jetbrains-"+size+".pf2")) {
			return "DejaVu Sans Mono"
		}
	}
	return "JetBrains Mono"
}

func menu(current theme) bool {
	_, err := render(filepath.Join(templates, "theme.txt"), filepath.Join(grubDir, "theme.txt"), "BACKGROUND="+current.background,
		"FOREGROUND="+current.foreground, "ACCENT="+current.accent, "MUTED="+current.muted, "FONT="+menuFace())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err == nil
}
