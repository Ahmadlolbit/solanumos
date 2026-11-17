package login

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/config"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/session"
	"github.com/Ahmadlolbit/solanumos/internal/term"
)

const usage = `Usage: solanum-session [up]

The desktop session the login screen starts. It runs start-hyprland and
keeps what it says in ~/.local/state/solanum/session.log, and the last
login's in session.log.1. When Hyprland stops before it finishes starting,
it shows what happened and the ways out on the console, unless the setting
Recovery screen is off. up is what the desktop runs once it has come up: it
marks this login as one that worked, with the kernel it runs.
`

func marker() string {
	return filepath.Join(env.Or("XDG_RUNTIME_DIR", env.OwnState()), session.Marker)
}

func me() account.Account {
	return account.Account{Name: os.Getenv("USER"), UID: strconv.Itoa(os.Getuid()), Home: env.Home()}
}

func attach() {
	if !term.IsTerminal(os.Stdin) || term.IsTerminal(os.Stdout) {
		return
	}
	for _, target := range []int{1, 2} {
		if err := syscall.Dup3(int(os.Stdin.Fd()), target, 0); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func confirmed(input *bufio.Reader) bool {
	fmt.Print("Press Enter to restart, or type 0 and press Enter to go back: ")
	line, err := input.ReadString('\n')
	return strings.TrimSpace(line) != "0" && (err == nil || line != "")
}

func restart() session.Next {
	if run.Status("systemctl", "reboot") == 0 {
		return session.LogOut
	}
	fmt.Println("The restart didn't start, so here are the choices again.")
	return session.Back
}

func carry(choice session.Choice, input *bufio.Reader) session.Next {
	fmt.Print(choice.Say)
	if choice.Wait && !confirmed(input) {
		return session.Back
	}
	for _, command := range choice.Run {
		if run.Status(command[0], command[1:]...) != 0 {
			fmt.Println("That didn't work, so here are the choices again.")
			return session.Back
		}
	}
	if choice.Tried != "" {
		if err := session.NoteTried(env.Home(), choice.Tried); err != nil {
			fmt.Println(err)
		}
	}
	if choice.Then == session.Restart {
		return restart()
	}
	return choice.Then
}

func recovery(input *bufio.Reader) session.Next {
	for {
		next := carry(session.Recover(input, os.Stdout, session.Gather(me(), run.Minute)), input)
		if next != session.Back {
			return next
		}
	}
}

func desktop() int {
	off := config.Read(config.Path()).Get("SESSION_RECOVERY") == "off"
	input := bufio.NewReader(os.Stdin)
	for {
		if session.Start([]string{"start-hyprland"}, marker(), filepath.Join(env.Home(), session.OwnLog)) == session.Up || off {
			return 0
		}
		attach()
		if recovery(input) != session.Again {
			return 0
		}
	}
}

func up() int {
	if _, err := session.MarkUp(marker(), filepath.Join(env.Home(), session.GoodLogin)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func Main() {
	switch env.Arg(1) {
	case "":
		os.Exit(desktop())
	case "up":
		os.Exit(up())
	case "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
}
