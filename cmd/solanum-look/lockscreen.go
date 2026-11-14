package main

import (
	"cmp"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/battery"
	"github.com/Ahmadlolbit/solanumos/internal/bus"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	trackPlaying   = "\U000f075a"
	trackPaused    = "\U000f03e4"
	batteryPlugged = "\U000f0084"
	lockLineRunes  = 60
	trackFormat    = "{{status}}\t{{title}}\t{{artist}}"
	playerName     = "org.mpris.MediaPlayer2."
)

var (
	batteryGlyphs = []string{"\U000f0083", "\U000f007a", "\U000f007b", "\U000f007c", "\U000f007d", "\U000f007e", "\U000f007f", "\U000f0080", "\U000f0081", "\U000f0082", "\U000f0079"}
	markup        = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
)

func shortened(text string) string {
	runes := []rune(text)
	if len(runes) <= lockLineRunes {
		return text
	}
	return strings.TrimSpace(string(runes[:lockLineRunes-1])) + "…"
}

func chosenTrack(output string) []string {
	var paused []string
	for _, line := range strings.Split(output, "\n") {
		fields := strings.SplitN(line, "\t", 3)
		if len(fields) < 3 || fields[1] == "" {
			continue
		}
		if fields[0] == "Playing" {
			return fields
		}
		if fields[0] == "Paused" && paused == nil {
			paused = fields
		}
	}
	return paused
}

func isPlayer(name string) bool {
	return strings.HasPrefix(name, playerName)
}

func noPlayer() bool {
	names, err := bus.Names()
	return err == nil && !slices.ContainsFunc(names, isPlayer)
}

func trackLine() string {
	if noPlayer() {
		return ""
	}
	output, err := run.CaptureQuiet(run.Quick, "playerctl", "--all-players", "metadata", "--format", trackFormat)
	track := chosenTrack(output)
	if err != nil || track == nil {
		return ""
	}
	glyph, text := trackPaused, track[1]
	if track[0] == "Playing" {
		glyph = trackPlaying
	}
	if track[2] != "" {
		text += " · " + track[2]
	}
	return glyph + " " + markup.Replace(shortened(text))
}

func batteryLine() string {
	reading, found := battery.Read()
	if !found {
		return ""
	}
	glyph, note := batteryPlugged, ""
	if !reading.Plugged {
		glyph = batteryGlyphs[(reading.Percent+5)/10]
	}
	if reading.AtLimit {
		note = " · charge limit reached"
	}
	return fmt.Sprintf("%s %d%%%s", glyph, reading.Percent, note)
}

func nameLine() string {
	me, err := user.Current()
	if err != nil {
		return ""
	}
	return markup.Replace(cmp.Or(me.Name, me.Username))
}

func capsLine() string {
	leds, _ := filepath.Glob(filepath.Join(env.Or("SOLANUM_LEDS", "/sys/class/leds"), "*::capslock", "brightness"))
	for _, led := range leds {
		text, _ := fsutil.Text(led)
		if value := strings.TrimSpace(text); value != "" && value != "0" {
			return "Caps Lock is on"
		}
	}
	return ""
}

func lockShown(key string, line func() string) int {
	if key == "" || settings.Get(key) == "yes" {
		fmt.Print(line())
	}
	return 0
}

func lockLine(args []string) int {
	switch arg(args, 0) {
	case "media":
		return lockShown("lock_media", trackLine)
	case "battery":
		return lockShown("lock_battery", batteryLine)
	case "name":
		return lockShown("", nameLine)
	case "caps":
		return lockShown("", capsLine)
	}
	fmt.Fprintln(os.Stderr, "lock-line takes media, battery, name or caps")
	return 2
}

func lockMedia(args []string) int {
	return onOff{"lock_media", putThen("lock_media", "yes", nothing), putThen("lock_media", "no", nothing)}.run(args)
}

func lockBattery(args []string) int {
	return onOff{"lock_battery", putThen("lock_battery", "yes", nothing), putThen("lock_battery", "no", nothing)}.run(args)
}
