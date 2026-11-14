package settings

import (
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const (
	release  = "SolanumOS Chert"
	hardware = "host:kernel:cpu:gpu:memory:swap:disk:display:battery"
)

type Fact struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type About struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Debian  string `json:"debian"`
	Facts   []Fact `json:"facts"`
}

func facts(text string) []Fact {
	var found []Fact
	for _, line := range strings.Split(text, "\n") {
		label, value, ok := strings.Cut(strings.TrimSpace(line), ": ")
		if ok && label != "" && strings.TrimSpace(value) != "" {
			found = append(found, Fact{label, strings.TrimSpace(value)})
		}
	}
	return found
}

func plainFacts() []Fact {
	var found []Fact
	if kernel := strings.TrimSpace(fsutil.Optional(env.Or("SOLANUM_KERNEL_RELEASE", "/proc/sys/kernel/osrelease"))); kernel != "" {
		found = append(found, Fact{"Kernel", "Linux " + kernel})
	}
	for _, line := range strings.Split(fsutil.Optional(env.Or("SOLANUM_CPUINFO", "/proc/cpuinfo")), "\n") {
		if label, value, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(label) == "model name" {
			return append(found, Fact{"CPU", strings.TrimSpace(value)})
		}
	}
	return found
}

func Facts() About {
	about := About{Name: release, Version: env.Version, Debian: strings.TrimSpace(fsutil.Optional(env.Or("SOLANUM_DEBIAN_VERSION", "/etc/debian_version")))}
	output, err := run.CaptureQuiet(run.Slow, "fastfetch", "--config", "none", "--logo", "none", "--pipe", "--structure", hardware)
	if about.Facts = facts(output); err != nil || len(about.Facts) == 0 {
		about.Facts = plainFacts()
	}
	return about
}
