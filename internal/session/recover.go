package session

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/privilege"
)

type Next int

const (
	Back Next = iota
	Again
	Restart
	LogOut
)

type Choice struct {
	Run   [][]string
	Say   string
	Wait  bool
	Then  Next
	Tried string
}

type item struct {
	label  string
	choice Choice
}

func crashed(ending Ending) string {
	said := "It crashed"
	if ending.Graphics {
		said += " while starting the graphics"
	}
	if ending.Frame != "" {
		said += ", in " + ending.Frame
	}
	return said + "."
}

func (newer Newer) said() string {
	came := "came"
	if newer.Came != "" {
		came += " on " + newer.Came + ","
	}
	return "A newer kernel, " + newer.Kernel + ", " + came + " after the last login that worked, which ran " + newer.Previous + "."
}

func (facts Facts) reason() string {
	lines := []string{"Hyprland stopped before it finished starting."}
	if facts.Ending.Crashed {
		lines = append(lines, crashed(facts.Ending))
	}
	if facts.Newer.Kernel != "" {
		lines = append(lines, facts.Newer.said())
	}
	if facts.Changed != "" {
		lines = append(lines, "Packages changed after the last login that worked, last on "+facts.Changed+".")
	}
	return strings.Join(lines, "\n") + "\n"
}

func (facts Facts) menuSteps() string {
	open, pick := ", open Advanced options", "the entry with "+facts.Newer.Previous
	if facts.Entry.ID != "" {
		open, pick = "", facts.Entry.Title
		if facts.Entry.Submenu != "" {
			open = ", open " + facts.Entry.Submenu
		}
	}
	return "Restart, press Esc right after the laptop's logo" + open + "\nand pick " + pick + ".\n"
}

func (facts Facts) previousKernel() Choice {
	tried := facts.Newer.Previous
	if facts.Writable && facts.Entry.ID != "" {
		return Choice{Run: [][]string{privilege.AsRoot("grub-reboot", facts.Entry.ID)}, Then: Restart, Tried: tried}
	}
	return Choice{Say: facts.menuSteps(), Wait: true, Then: Restart, Tried: tried}
}

func (facts Facts) items() []item {
	items := []item{{"Try again", Choice{Then: Again}}}
	if facts.Newer.Previous != "" {
		items = append(items, item{"Start the previous kernel, " + facts.Newer.Previous + ", once", facts.previousKernel()})
	}
	if facts.Restore != "" {
		items = append(items, item{"Go back to before the last update (solanum restore " + facts.Restore + ")", Choice{Run: [][]string{{"solanum", "restore", facts.Restore}}}})
	}
	return append(items,
		item{"Check the system (sudo solanum doctor)", Choice{Run: [][]string{privilege.AsRoot("solanum", "doctor")}}},
		item{"Collect the logs for a bug report (solanum collect)", Choice{Run: [][]string{{"solanum", "collect"}}}},
		item{"Open a shell; exit comes back here", Choice{Run: [][]string{{env.Or("SHELL", "/bin/bash")}}}},
		item{"Restart", Choice{Then: Restart}},
		item{"Log out", Choice{Then: LogOut}},
	)
}

func (facts Facts) screen(items []item) string {
	var text strings.Builder
	text.WriteString("\nSolanumOS: the desktop didn't start\n\n" + facts.reason())
	if len(facts.Tail) > 0 {
		text.WriteString("\nHyprland's last lines, in ~/" + OwnLog + ":\n  " + strings.Join(facts.Tail, "\n  ") + "\n")
	}
	text.WriteString("\n")
	for index, each := range items {
		fmt.Fprintf(&text, "  %d  %s\n", index+1, each.label)
	}
	return text.String() + "\n"
}

func Recover(in *bufio.Reader, out io.Writer, facts Facts) Choice {
	items := facts.items()
	fmt.Fprint(out, facts.screen(items))
	for {
		fmt.Fprintf(out, "Type a number from 1 to %d and press Enter: ", len(items))
		line, err := in.ReadString('\n')
		if number, bad := strconv.Atoi(strings.TrimSpace(line)); bad == nil && number >= 1 && number <= len(items) {
			return items[number-1].choice
		}
		if err != nil {
			return Choice{Then: LogOut}
		}
	}
}
