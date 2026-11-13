package wizard

import (
	"fmt"
	"os"

	"github.com/Ahmadlolbit/solanumos/internal/tui"
)

func Main() {
	if len(os.Args) > 1 {
		fmt.Println("solanum opens the setup with no arguments; for its commands see: solanum help")
		return
	}
	os.Exit(tui.Run())
}
