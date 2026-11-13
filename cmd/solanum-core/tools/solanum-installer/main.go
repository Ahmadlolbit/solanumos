package installer

import (
	"fmt"
	"os"

	"github.com/Ahmadlolbit/solanumos/internal/osinstall"
	"github.com/Ahmadlolbit/solanumos/internal/tui"
)

func Main() {
	if len(os.Args) > 1 {
		fmt.Println("solanum-installer opens the installer with no arguments, as root, on the Solanum live system")
		return
	}
	if os.Getenv("SOLANUM_ROOT") == "" {
		os.Setenv("SOLANUM_ROOT", osinstall.Clone())
	}
	os.Exit(tui.RunInstaller())
}
