package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/look"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/screens"
)

const (
	tryLimit    = 120
	tryPoll     = 100 * time.Millisecond
	revertWord  = "revert"
	beingTried  = "a layout is being tried: keep it with solanum-look screens confirm, or put the old one back with solanum-look screens revert"
	nothingTold = "no layout is being tried"
)

func screensFile() string {
	return filepath.Join(env.ConfigHome(), "solanum", "screens.conf")
}

var procRoot = "/proc"

func trialFile() string {
	return stateHome("solanum", "screens-trying")
}

func beforeFile() string {
	return stateHome("solanum", "screens-before")
}

func monitorsNow() ([]screens.Monitor, bool) {
	output, err := run.Capture(run.Quick, "hyprctl", "monitors", "all", "-j")
	if err != nil {
		return nil, false
	}
	monitors, err := screens.ParseMonitors(output)
	if err == nil && len(monitors) == 0 {
		err = errors.New("hyprctl lists no screens")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, false
	}
	return monitors, true
}

func savedLayouts() ([]screens.Profile, string, bool) {
	data, err := os.ReadFile(screensFile())
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(os.Stderr, err)
		return nil, "", false
	}
	return screens.Parse(string(data)), string(data), true
}

func writeLayouts(text string) bool {
	path := screensFile()
	err := os.MkdirAll(filepath.Dir(path), 0o777)
	if err == nil {
		var unlock func()
		if unlock, err = look.Hold(filepath.Dir(path)); err == nil {
			defer unlock()
			err = putLayouts(path, text)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err == nil
}

func putLayouts(path, text string) error {
	if text == "" {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	return atomic.Write(path, []byte(text), 0o600)
}

func reloadScreens() {
	run.Quiet(run.Quick, "hyprctl", "reload")
}

func liveTrial(word string) bool {
	pid, err := strconv.Atoi(strings.TrimSuffix(word, " "+revertWord))
	if err != nil || pid <= 0 || syscall.Kill(pid, 0) != nil {
		return false
	}
	data, err := os.ReadFile(filepath.Join(procRoot, strconv.Itoa(pid), "cmdline"))
	args := strings.Split(strings.TrimRight(string(data), "\x00"), "\x00")
	return err == nil && strings.HasPrefix(filepath.Base(args[0]), "solanum-look") && slices.Contains(args, "--try")
}

func trying() bool {
	return liveTrial(stored(trialFile()))
}

func settleTrial() {
	word := stored(trialFile())
	if word == "" || liveTrial(word) {
		return
	}
	if before, err := os.ReadFile(beforeFile()); err == nil {
		writeLayouts(string(before))
		fmt.Fprintln(os.Stderr, "put back the layout from before a trial that never ended")
	}
	remove(beforeFile())
	remove(trialFile())
	reloadScreens()
}

func startTrial(before string) bool {
	if !mark(beforeFile(), before) {
		return false
	}
	if err := atomic.Stage(trialFile(), []byte(strconv.Itoa(os.Getpid())+"\n"), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, beingTried)
		remove(beforeFile())
		return false
	}
	return true
}

func endTrial() {
	remove(trialFile())
	remove(beforeFile())
}

func layoutNow() (screens.Profile, []screens.Monitor, []screens.Profile, string, bool) {
	monitors, ok := monitorsNow()
	if !ok {
		return screens.Profile{}, nil, nil, "", false
	}
	saved, before, ok := savedLayouts()
	if !ok {
		return screens.Profile{}, nil, nil, "", false
	}
	return screens.WithSaved(screens.Live(monitors), saved), monitors, saved, before, true
}

func trySeconds(args []string) (int, []string, bool) {
	if arg(args, 0) != "--try" {
		return 0, args, true
	}
	seconds, err := strconv.Atoi(arg(args, 1))
	if err != nil || seconds < 1 || seconds > tryLimit {
		fmt.Fprintf(os.Stderr, "--try takes seconds from 1 to %d\n", tryLimit)
		return 0, nil, false
	}
	return seconds, args[2:], true
}

func screensSave(args []string) int {
	seconds, args, ok := trySeconds(args)
	if !ok {
		return 2
	}
	if trying() {
		fmt.Fprintln(os.Stderr, beingTried)
		return 1
	}
	live, monitors, saved, before, ok := layoutNow()
	if !ok {
		return 1
	}
	profile, err := screens.Change(live, monitors, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if seconds > 0 && !startTrial(before) {
		return 1
	}
	if !writeLayouts(screens.Format(screens.Put(saved, profile))) {
		endTrial()
		return 1
	}
	reloadScreens()
	if seconds > 0 {
		return tryLayout(seconds, before)
	}
	fmt.Println("Saved the layout for these screens.")
	return 0
}

func mark(path, value string) bool {
	err := os.MkdirAll(filepath.Dir(path), 0o777)
	if err == nil {
		err = atomic.Write(path, []byte(value), 0o600)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return err == nil
}

func tryLayout(seconds int, before string) int {
	fmt.Printf("Keep this layout? Run solanum-look screens confirm within %d seconds, or the old one comes back.\n", seconds)
	if answered(time.Duration(seconds) * time.Second) {
		endTrial()
		fmt.Println("Kept the new layout.")
		return 0
	}
	writeLayouts(before)
	endTrial()
	reloadScreens()
	fmt.Println("Put the old layout back.")
	return 3
}

func answered(limit time.Duration) bool {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(stop)
	tick := time.NewTicker(tryPoll)
	defer tick.Stop()
	deadline := time.After(limit)
	for {
		select {
		case <-stop:
			return false
		case <-deadline:
			return false
		case <-tick.C:
			if word := stored(trialFile()); word == "" || strings.HasSuffix(word, " "+revertWord) {
				return word == ""
			}
		}
	}
}

func screensConfirm([]string) int {
	if !trying() {
		fmt.Fprintln(os.Stderr, nothingTold)
		return 1
	}
	remove(trialFile())
	return 0
}

func screensRevert([]string) int {
	if !trying() {
		fmt.Fprintln(os.Stderr, nothingTold)
		return 1
	}
	if !mark(trialFile(), stored(trialFile())+" "+revertWord+"\n") {
		return 1
	}
	return 0
}

func screensApply([]string) int {
	live, _, saved, _, ok := layoutNow()
	if !ok {
		return 1
	}
	reloadScreens()
	if screens.Find(saved, live) < 0 {
		fmt.Println("No layout is saved for these screens, so they have the default one.")
		return 0
	}
	fmt.Println("Put back the layout saved for these screens.")
	return 0
}

func screensForget([]string) int {
	if trying() {
		fmt.Fprintln(os.Stderr, beingTried)
		return 1
	}
	live, _, saved, _, ok := layoutNow()
	if !ok {
		return 1
	}
	kept, found := screens.Forget(saved, live)
	if !found {
		fmt.Println("No layout is saved for these screens.")
		return 0
	}
	if !writeLayouts(screens.Format(kept)) {
		return 1
	}
	reloadScreens()
	fmt.Println("Forgot the layout for these screens; they have the default one again.")
	return 0
}

func screensList([]string) int {
	live, _, saved, _, ok := layoutNow()
	if !ok {
		return 1
	}
	if screens.Find(saved, live) < 0 {
		fmt.Printf("%s (these screens, no layout saved)\n", live.Key)
	}
	for _, profile := range saved {
		note := ""
		if screens.SameScreens(profile, live) {
			note = " (these screens)"
		}
		fmt.Printf("%s%s\n", profile.Key, note)
		for _, screen := range profile.Screens {
			fmt.Println("  " + screens.Summary(screen))
		}
	}
	return 0
}

func screensNow([]string) int {
	live, monitors, saved, _, ok := layoutNow()
	if !ok {
		return 1
	}
	data, err := json.Marshal(screens.Shown(live, monitors, screens.Find(saved, live) >= 0, trying()))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(string(data))
	return 0
}

func screensWait(args []string) int {
	value := arg(args, 0)
	if value == "" {
		return show("screens_wait")
	}
	seconds, ok := look.WholeNumber(value)
	if !ok || seconds < 5 || seconds > tryLimit {
		fmt.Fprintf(os.Stderr, "screens-wait takes seconds from 5 to %d\n", tryLimit)
		return 2
	}
	return putThen("screens_wait", strconv.FormatInt(seconds, 10), nothing)()
}

func screensCommand(args []string) int {
	words := map[string]func([]string) int{
		"list": screensList, "now": screensNow, "save": screensSave, "apply": screensApply,
		"forget": screensForget, "confirm": screensConfirm, "revert": screensRevert, "settle": func([]string) int { return 0 },
	}
	word := arg(args, 0)
	if word == "" {
		word = "list"
	}
	command, known := words[word]
	if !known {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	settleTrial()
	return command(args[min(1, len(args)):])
}
