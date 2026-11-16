package main

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/bootimage"
	"github.com/Ahmadlolbit/solanumos/internal/config"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/grub"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
)

const keptKey = "KEPT_KERNEL"

func runningKernel() string {
	running, _ := run.CaptureQuiet(run.Quick, "uname", "-r")
	return running
}

func keptNow() string {
	return cmp.Or(config.Read(config.Path()).Get(keptKey), grub.HandMade())
}

func kernelMarks(kernel, running, kept string) string {
	var marks []string
	if kernel == running {
		marks = append(marks, "running")
	}
	if kernel == kept {
		marks = append(marks, "kept")
	}
	return strings.Join(marks, ", ")
}

func showKernels() int {
	running, kept := runningKernel(), keptNow()
	var text strings.Builder
	text.WriteString("Kernels, newest last:\n")
	for _, kernel := range bootimage.Kernels() {
		text.WriteString(strings.TrimRight("  "+kernel+"  "+kernelMarks(kernel, running, kept), " ") + "\n")
	}
	if kept == "" {
		text.WriteString("The computer starts the newest kernel; sudo solanum kernel keep keeps the running one.\n")
	} else {
		text.WriteString("The computer starts the kept kernel, " + kept + "; sudo solanum kernel newest starts the newest again.\n")
	}
	return say(text.String())
}

func saveKept(release string) int {
	current := config.Read(config.Path())
	current.Set(keptKey, release)
	if err := current.Save(config.Path()); err != nil {
		return fail("%v", err)
	}
	return 0
}

func keepKernel(version string) int {
	release := cmp.Or(version, runningKernel())
	if !fsutil.IsFile(filepath.Join(bootimage.Folder(), "vmlinuz-"+release)) {
		return fail("there's no kernel %s in %s; see: solanum kernel", release, bootimage.Folder())
	}
	if err := grub.Keep(os.Stdout, release); err != nil {
		return fail("%v", err)
	}
	if code := saveKept(release); code != 0 {
		return code
	}
	return say("The computer starts " + release + " from now on, and apt keeps it installed, until you run sudo solanum kernel newest.\n")
}

func startNewest() int {
	if err := grub.Newest(os.Stdout); err != nil {
		return fail("%v", err)
	}
	if code := saveKept(""); code != 0 {
		return code
	}
	return say("The computer starts the newest kernel, " + bootimage.Newest() + ", from the next restart.\n")
}

func newerThanKept() []string {
	kept, kernels := keptNow(), bootimage.Kernels()
	at := slices.Index(kernels, kept)
	if kept == "" || at < 0 || at == len(kernels)-1 {
		return nil
	}
	newest := kernels[len(kernels)-1]
	return []string{
		"You keep the kernel " + kept + ", and " + newest + " is newer",
		"Try it: restart, press Esc right after the laptop's logo, open Advanced options and pick " + newest +
			". If it works, sudo solanum kernel newest starts the newest from then on.",
	}
}

func keptKernelCheck() {
	kept, news := keptNow(), newerThanKept()
	if kept == "" {
		return
	}
	if news == nil {
		ui.OK("The kept kernel, " + kept + ", is the newest")
		return
	}
	ui.Warn(news[0])
	ui.Note(news[1])
}

func kernelCommand() int {
	switch env.Arg(2) {
	case "":
		return showKernels()
	case "keep":
		if env.Arg(4) == "" {
			return asRoot("kernel keep", alone(func() int { return keepKernel(env.Arg(3)) }))
		}
	case "newest":
		if env.Arg(3) == "" {
			return asRoot("kernel newest", alone(startNewest))
		}
	}
	return fail("usage: solanum kernel [keep [VERSION] | newest]")
}

func keptKernelNews() {
	if news := newerThanKept(); news != nil {
		ui.Print(fmt.Sprintf("%s.\n%s\n", news[0], news[1]))
	}
}
