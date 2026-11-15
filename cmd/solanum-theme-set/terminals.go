package main

import (
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
)

const alacrittyTheme = "alacritty.toml"

func linkAlacritty(current string) {
	if fsutil.IsFile(filepath.Join(current, alacrittyTheme)) {
		missed("Alacritty", atomic.Symlink(filepath.Join(filepath.Base(current), alacrittyTheme), filepath.Join(state(), alacrittyTheme)))
	}
}
