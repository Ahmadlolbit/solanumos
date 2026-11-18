package main

import (
	"fmt"
	"slices"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/btrfs"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/snapshots"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
)

const (
	cleanGoal   = 85
	deleteLimit = 30 * time.Minute
	freeWindow  = 2
	topMount    = "/btrfsroot"
)

func stopped(deleted int, format string, args ...any) int {
	if deleted > 0 {
		say("Deleted " + snapshots.Count(deleted, "snapshot") + ".\n")
	}
	return fail(format, args...)
}

func freesLittle(uses []int) bool {
	last := len(uses) - 1
	return last >= freeWindow && uses[last-freeWindow]-uses[last] < 1
}

func oldRoots() string {
	listing, err := run.CaptureQuiet(run.Minute, "btrfs", "subvolume", "list", "/")
	if err != nil {
		return ""
	}
	roots := snapshots.OldRoots(btrfs.RootSubvolume(), listing)
	if len(roots) == 0 {
		return ""
	}
	text := "Old roots kept after a restore take space too. To delete them:\n  sudo mount -o subvolid=5 " + btrfs.RootDevice() + " " + topMount + "\n"
	for _, root := range roots {
		text += "  sudo btrfs subvolume delete " + topMount + "/" + root + "\n"
	}
	return text + "  sudo umount " + topMount + "\n"
}

func elsewhere(stalled bool) string {
	text := "The space is used by files outside the snapshots, most likely your own in /home.\n"
	if stalled {
		text = "Deleting more snapshots would free little, so the rest stay. " + text
	}
	return text + oldRoots()
}

func gone(number string) bool {
	all, err := snapshots.All()
	return err == nil && !slices.ContainsFunc(all, func(snapshot snapshots.Snapshot) bool { return snapshot.Number == number })
}

func deleteUntilClean(numbers []string, use int) int {
	deleted, uses := 0, []int{use}
	for _, number := range numbers {
		switch {
		case run.Within(deleteLimit, "snapper", "--no-dbus", "-c", "root", "delete", "--sync", number) == 0:
			deleted++
		case !gone(number):
			return stopped(deleted, "snapshot %s could not be deleted", number)
		}
		var ok bool
		if use, ok = snapshots.DiskUse(); !ok {
			return stopped(deleted, "cannot tell how full / is")
		}
		uses = append(uses, use)
		if use < cleanGoal || freesLittle(uses) {
			break
		}
	}
	text := fmt.Sprintf("Deleted %s. / is %d%% full.\n", snapshots.Count(deleted, "snapshot"), use)
	if use >= cleanGoal {
		text += elsewhere(len(uses) <= len(numbers))
	}
	return say(text)
}

func cleanUp(yes bool) int {
	if !installed("snapper") {
		return fail("snapper isn't installed, so there's nothing to clean up")
	}
	use, ok := snapshots.DiskUse()
	if !ok {
		return fail("cannot tell how full / is")
	}
	if use < cleanGoal {
		return say(fmt.Sprintf("/ is %d%% full, so there is nothing to clean up.\n", use))
	}
	all, err := snapshots.All()
	if err != nil {
		return fail("the snapshots can't be read")
	}
	numbers := snapshots.Disposable(all)
	if len(numbers) == 0 {
		return say(fmt.Sprintf("/ is %d%% full, but there are no old automatic snapshots to delete.\n", use) + elsewhere(false))
	}
	say(fmt.Sprintf("/ is %d%% full. Up to %s can go, oldest first, until it is under %d%%.\n"+
		"The newest three and the ones with your own description stay.\n", use, snapshots.Count(len(numbers), "old automatic snapshot"), cleanGoal))
	if !yes && !ui.CanAsk() {
		return say("Nothing was deleted. Run it in a terminal to answer, or with --yes.\n")
	}
	if !yes && !ui.Confirm("Delete them?") {
		return 0
	}
	return deleteUntilClean(numbers, use)
}

type cleanupPlan struct {
	Use       int               `json:"use"`
	Known     bool              `json:"known"`
	Goal      int               `json:"goal"`
	Snapshots int               `json:"snapshots"`
	Old       []snapshots.Point `json:"old"`
	Held      int64             `json:"held"`
	HeldKnown bool              `json:"heldKnown"`
	Unread    bool              `json:"unread"`
}

func planCleanup() int {
	plan := cleanupPlan{Goal: cleanGoal, Old: []snapshots.Point{}}
	plan.Use, plan.Known = snapshots.DiskUse()
	if !installed("snapper") {
		return printed(plan)
	}
	all, err := snapshots.All()
	if err != nil {
		plan.Unread = true
		return printed(plan)
	}
	numbers := snapshots.Disposable(all)
	plan.Snapshots = len(all)
	for _, snapshot := range all {
		if slices.Contains(numbers, snapshot.Number) {
			plan.Old = append(plan.Old, snapshot.Point())
		}
	}
	plan.Held, plan.HeldKnown = snapshots.Held(numbers)
	return printed(plan)
}

func cleanupCommand() int {
	if env.Arg(2) == "--plan" && env.Arg(3) == "--json" && env.Arg(4) == "" {
		return planCleanup()
	}
	yes := env.Arg(2) == "--yes"
	if (env.Arg(2) != "" && !yes) || env.Arg(3) != "" {
		return fail("usage: solanum cleanup [--yes]")
	}
	return asRoot("cleanup", alone(func() int { return cleanUp(yes) }))
}
