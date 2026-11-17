package login

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/session"
)

const usage = `Usage: solanum-session [up]

The desktop session the login screen starts. It runs start-hyprland and
keeps what it says in ~/.local/state/solanum/session.log, and the last
login's in session.log.1. up is what the desktop runs once it has come
up: it marks this login as one that worked, with the kernel it runs.
`

func marker() string {
	return filepath.Join(env.Or("XDG_RUNTIME_DIR", env.OwnState()), session.Marker)
}

func up() int {
	if err := session.MarkUp(marker(), filepath.Join(env.Home(), session.GoodLogin)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func Main() {
	switch env.Arg(1) {
	case "up":
		os.Exit(up())
	case "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
}
