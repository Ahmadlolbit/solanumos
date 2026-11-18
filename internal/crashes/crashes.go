package crashes

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	MessageID = "fc2e22bc6ee647b6b90729ab34a250b1"
	Hint      = "Click to see what happened, or middle-click to write a report."
	Window    = time.Hour
	deleted   = " (deleted)"
	longest   = 60
)

var fields = []string{"COREDUMP_PID", "COREDUMP_COMM", "COREDUMP_EXE"}

type Crash struct {
	PID     string
	Program string
}

type Limit struct {
	shown  map[string]time.Time
	missed map[string]int
}

func Journal(matches ...string) []string {
	args := []string{"--no-pager", "--quiet", "--output=json", "--output-fields=" + strings.Join(fields, ","), "MESSAGE_ID=" + MessageID}
	return append(args, matches...)
}

func value(entry map[string]json.RawMessage, name string) string {
	var text string
	if json.Unmarshal(entry[name], &text) != nil {
		return ""
	}
	return text
}

func printable(text string) string {
	text = strings.Map(func(letter rune) rune {
		if unicode.IsControl(letter) {
			return -1
		}
		return letter
	}, text)
	if runes := []rune(text); len(runes) > longest {
		return string(runes[:longest]) + "…"
	}
	return text
}

func program(exe, comm string) string {
	if exe = strings.TrimSuffix(exe, deleted); exe != "" {
		return printable(filepath.Base(exe))
	}
	return printable(comm)
}

func Parse(line string) (Crash, bool) {
	var entry map[string]json.RawMessage
	if json.Unmarshal([]byte(line), &entry) != nil {
		return Crash{}, false
	}
	crash := Crash{PID: value(entry, "COREDUMP_PID"), Program: program(value(entry, "COREDUMP_EXE"), value(entry, "COREDUMP_COMM"))}
	pid, err := strconv.Atoi(crash.PID)
	return crash, err == nil && pid > 0 && crash.Program != ""
}

func NewLimit() *Limit {
	return &Limit{shown: map[string]time.Time{}, missed: map[string]int{}}
}

func (limit *Limit) Take(program string, now time.Time) (bool, int) {
	if last, seen := limit.shown[program]; seen && now.Sub(last) < Window {
		limit.missed[program]++
		return false, 0
	}
	missed := limit.missed[program]
	limit.shown[program] = now
	delete(limit.missed, program)
	return true, missed
}

func Summary(program string) string {
	return program + " stopped unexpectedly"
}

func Body(missed int) string {
	switch {
	case missed == 1:
		return "It stopped once more since the last notice. " + Hint
	case missed > 1:
		return fmt.Sprintf("It stopped %d more times since the last notice. ", missed) + Hint
	}
	return Hint
}
