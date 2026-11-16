package terminal

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/Ahmadlolbit/solanumos/internal/cleanup"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/term"
)

const usage = `Usage: solanum-terminal [--wait] COMMAND [ARGUMENT...]

Runs COMMAND in a floating terminal in the middle of the screen.
--wait keeps the terminal open when COMMAND ends, until a key is pressed.
`

const stoppedCode = 130

func inside(wait string, command []string) int {
	if wait != "1" {
		return run.Status(command[0], command[1:]...)
	}
	interrupted := make(chan os.Signal, 1)
	cleanup.NotifyInterrupt(interrupted)
	status := run.Status(command[0], command[1:]...)
	signal.Stop(interrupted)
	fmt.Println()
	switch status {
	case 0:
		fmt.Print("Done. Press any key to close.")
	case stoppedCode:
		fmt.Print("Stopped. Press any key to close.")
	default:
		fmt.Printf("It ended with an error (exit %d). Press any key to close.", status)
	}
	term.ReadKey()
	return status
}

func Main() {
	args := os.Args[1:]
	wait := "0"
	if len(args) > 0 {
		switch args[0] {
		case "--inside":
			if len(args) < 3 {
				fmt.Fprint(os.Stderr, usage)
				os.Exit(2)
			}
			os.Exit(inside(args[1], args[2:]))
		case "--wait":
			wait = "1"
			args = args[1:]
		case "-h", "--help":
			fmt.Print(usage)
			return
		}
	}
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	os.Exit(run.Report(run.Exec("kitty", append([]string{"--class", "solanum.floating", "-e", run.Self(), "--inside", wait}, args...)...)))
}
