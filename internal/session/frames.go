package session

import (
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	ownModule   = "Hyprland"
	traceHeader = "Stack trace of thread"
)

var (
	dumpFrame     = regexp.MustCompile(`^\s*#\d+\s+0x[0-9a-f]+\s+(\S+)\s+\((\S+) \+ 0x[0-9a-f]+\)`)
	reportFrame   = regexp.MustCompile(`^\s*#\s*\d+\s*\|\s*([^\s(]+)\(([^+)]*)`)
	crashHandlers = []string{"getBacktrace", "CrashReporter", "handleUnrecoverableSignal", "handleCritSignal"}
	graphicsStart = []string{"initEGL", "CHyprOpenGLImpl::CHyprOpenGLImpl"}
)

type frame struct {
	name   string
	module string
}

func dumpFrames(info string) []frame {
	_, trace, found := strings.Cut(info, traceHeader)
	if !found {
		return nil
	}
	trace, _, _ = strings.Cut(trace, traceHeader)
	var frames []frame
	for _, line := range strings.Split(trace, "\n") {
		if match := dumpFrame.FindStringSubmatch(line); match != nil {
			frames = append(frames, frame{demangled(match[1]), match[2]})
		}
	}
	return frames
}

func reportFrames(text string) []frame {
	var frames []frame
	for _, line := range strings.Split(text, "\n") {
		if match := reportFrame.FindStringSubmatch(line); match != nil {
			frames = append(frames, frame{demangled(match[2]), match[1]})
		}
	}
	return frames
}

func handlesTheCrash(name string) bool {
	return slices.ContainsFunc(crashHandlers, func(handler string) bool { return strings.Contains(name, handler) })
}

func story(frames []frame) string {
	for _, each := range frames {
		if filepath.Base(each.module) == ownModule && each.name != "" && each.name != "n/a" && !handlesTheCrash(each.name) {
			return each.name
		}
	}
	return ""
}

func startingGraphics(frames []frame) bool {
	return slices.ContainsFunc(frames, func(each frame) bool {
		return slices.ContainsFunc(graphicsStart, func(marker string) bool { return strings.Contains(each.name, marker) })
	})
}

func lengthAt(text string) (int, int) {
	digits := len(text) - len(strings.TrimLeft(text, "0123456789"))
	size, err := strconv.Atoi(text[:digits])
	if err != nil || digits+size > len(text) {
		return 0, 0
	}
	return digits, size
}

func names(text string, nested bool) ([]string, string) {
	var parts []string
	for text != "" {
		digits, size := lengthAt(text)
		if size == 0 {
			break
		}
		parts = append(parts, text[digits:digits+size])
		text = text[digits+size:]
		if !nested {
			break
		}
	}
	return parts, text
}

func withConstructor(parts []string, rest string) []string {
	last := parts[len(parts)-1]
	switch {
	case strings.HasPrefix(rest, "C1"), strings.HasPrefix(rest, "C2"), strings.HasPrefix(rest, "C3"):
		return append(parts, last)
	case strings.HasPrefix(rest, "D0"), strings.HasPrefix(rest, "D1"), strings.HasPrefix(rest, "D2"):
		return append(parts, "~"+last)
	}
	return parts
}

func demangled(symbol string) string {
	rest, nested := strings.CutPrefix(symbol, "_ZN")
	if !nested {
		var mangled bool
		if rest, mangled = strings.CutPrefix(symbol, "_Z"); !mangled {
			return symbol
		}
	}
	parts, rest := names(strings.TrimLeft(rest, "rVKRO"), nested)
	switch {
	case len(parts) == 0:
		return symbol
	case nested:
		parts = withConstructor(parts, rest)
	}
	return strings.Join(parts, "::")
}
