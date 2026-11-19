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

func wrapped(text string) string {
	var out strings.Builder
	width := 0
	for _, word := range strings.Fields(text) {
		switch {
		case width == 0:
		case width+1+len(word) > screenWidth:
			out.WriteString("\n")
			width = 0
		default:
			out.WriteString(" ")
			width++
		}
		out.WriteString(word)
		width += len(word)
	}
	return out.String()
}

func crashed(ending Ending) string {
	said := "It crashed"
	if ending.Graphics {
		said += " while starting the graphics"
	}
	if ending.Frame != "" {
		said += ", in " + plain(ending.Frame)
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
	if facts.CameUp {
		lines = []string{"Hyprland stopped in its first minute, after the desktop had come up. The way back in is the login screen, with your password."}
	}
	if facts.Ending.Crashed {
		lines = append(lines, crashed(facts.Ending))
	}
	if facts.Newer.Kernel != "" {
		lines = append(lines, facts.Newer.said())
	}
	if facts.Changed != "" {
		lines = append(lines, "Packages changed after the last login that worked, last on "+facts.Changed+".")
	}
	for index, line := range lines {
		lines[index] = wrapped(line)
	}
	return strings.Join(lines, "\n") + "\n"
}

func (facts Facts) previousKernel() Choice {
	tried := facts.Newer.Previous
	if facts.Writable {
		return Choice{Run: [][]string{privilege.AsRoot("solanum", "kernel", "once", tried)}, Then: Restart, Tried: tried}
	}
	steps := wrapped("Restart, press Esc right after the laptop's logo, open Advanced options and pick the entry with " + tried + ".")
	return Choice{Say: steps + "\n", Wait: true, Then: Restart, Tried: tried}
}

func (facts Facts) items() []item {
	if facts.CameUp {
		return []item{{"Log out, back to the login screen", Choice{Then: LogOut}}, {"Restart", Choice{Then: Restart}}}
	}
	items := []item{{"Try again", Choice{Then: Again}}}
	if facts.Newer.Previous != "" {
		items = append(items, item{"Start the previous kernel, " + facts.Newer.Previous + ", once", facts.previousKernel()})
	}
	if facts.Restore != "" {
		items = append(items, item{"Go back to before the last update (solanum restore " + facts.Restore + ")", Choice{Run: [][]string{{"solanum", "restore", facts.Restore}}}})
	}
	return append(items,
		item{"Check the system (sudo solanum doctor)", Choice{Run: [][]string{privilege.AsRoot("solanum", "doctor")}}},
		item{"Collect the logs for a bug report (sudo solanum collect)", Choice{Run: [][]string{privilege.AsRoot("solanum", "collect")}}},
		item{"Open a shell; exit comes back here", Choice{Run: [][]string{{env.Or("SHELL", "/bin/bash")}}}},
		item{"Restart", Choice{Then: Restart}},
		item{"Log out", Choice{Then: LogOut}},
	)
}

func (facts Facts) screen(items []item) string {
	var text strings.Builder
	title := "the desktop didn't start"
	if facts.CameUp {
		title = "the desktop stopped"
	}
	text.WriteString("\nSolanumOS: " + title + "\n\n" + facts.reason())
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
