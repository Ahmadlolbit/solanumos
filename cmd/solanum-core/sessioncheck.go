package main

import (
	"io"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/bootimage"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/hypr"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/snapshots"
	"github.com/Ahmadlolbit/solanumos/internal/ui"
)

const kernelImage = "linux-image-"

var sessionPackages = []string{"hyprland", "libaquamarine*", "libhypr*", "libgbm1", "libegl*", "libgl1-mesa-dri", "nvidia-*", kernelImage + "*", "firmware-*"}

func sessionVersions() string {
	output, _ := run.CaptureQuiet(run.Minute, "dpkg-query", append([]string{"-W", "-f=${db:Status-Abbrev} ${Package} ${Version}\n"}, sessionPackages...)...)
	var installed []string
	for _, line := range run.Lines(output) {
		if strings.HasPrefix(line, "ii ") {
			installed = append(installed, strings.Join(strings.Fields(line), " "))
		}
	}
	slices.Sort(installed)
	return strings.Join(installed, "\n")
}

func kernelPackages(list string) []string {
	var kernels []string
	for _, line := range strings.Split(list, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		release, found := strings.CutPrefix(fields[1], kernelImage)
		if found && release != "" && release[0] >= '0' && release[0] <= '9' {
			kernels = append(kernels, strings.TrimSuffix(release, "-unsigned"))
		}
	}
	return kernels
}

func newKernels(before, after string) []string {
	old := kernelPackages(before)
	var fresh []string
	for _, kernel := range kernelPackages(after) {
		if !slices.Contains(old, kernel) && !slices.Contains(fresh, kernel) {
			fresh = append(fresh, kernel)
		}
	}
	return fresh
}

func nvidiaRestart() bool {
	loaded := strings.TrimSpace(fsutil.Optional(env.Or("SOLANUM_NVIDIA_VERSION", "/sys/module/nvidia/version")))
	installed := apt.Version("nvidia-driver")
	return loaded != "" && installed != "" && installed != loaded && !strings.HasPrefix(installed, loaded+"-")
}

func dkmsMissing() []string {
	output, err := run.CaptureQuiet(run.Minute, "dkms", "status")
	newest := bootimage.Newest()
	if err != nil || newest == "" {
		return nil
	}
	built := map[string]bool{}
	for _, line := range run.Lines(output) {
		if name, rest, found := strings.Cut(line, "/"); found {
			built[name] = built[name] || strings.Contains(rest, ", "+newest+",") && strings.HasSuffix(line, ": installed")
		}
	}
	var missing []string
	for name, ok := range built {
		if !ok {
			missing = append(missing, name)
		}
	}
	slices.Sort(missing)
	return missing
}

func verifier() account.Account {
	if person, helped := onBehalf(); helped {
		return person
	}
	return account.Account{}
}

func updatePoint() string {
	if !installed("snapper") {
		return ""
	}
	points, _ := snapshots.All()
	for _, point := range slices.Backward(points) {
		if point.Description == "before solanum update" {
			return point.Number
		}
	}
	return ""
}

func dontLogOut() int {
	restore := strings.TrimSpace("solanum restore " + updatePoint())
	ui.Print("Don't log out yet. The ways out:\n" +
		"  " + restore + " goes back to the snapshot taken before this update\n" +
		"  sudo solanum doctor checks the system\n" +
		"  solanum collect packs the logs for a bug report\n")
	return 1
}

func sessionCheck(before string) int {
	after := sessionVersions()
	if after == before {
		return 0
	}
	if ok, complaint := hypr.Accepts(io.Discard, verifier()); !ok {
		ui.Print("Hyprland doesn't take its configuration after this update:\n" + complaint)
		return dontLogOut()
	}
	if missing := dkmsMissing(); len(missing) > 0 {
		ui.Print("These drivers aren't built for the newest kernel, " + bootimage.Newest() + ": " + strings.Join(missing, ", ") + ". dkms status shows why.\n")
		return dontLogOut()
	}
	if nvidiaRestart() {
		ui.Print("NVIDIA's driver was updated: restart before you log out.\n")
		return 0
	}
	if fresh := newKernels(before, after); len(fresh) > 0 && keptNow() == "" {
		ui.Print("A new kernel was installed, " + strings.Join(fresh, ", ") + ". If the desktop doesn't come up after the restart, the login offers the previous kernel.\n")
	}
	return 0
}
