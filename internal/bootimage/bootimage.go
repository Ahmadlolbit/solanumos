package bootimage

import (
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/tpm"
)

const kernelPrefix = "vmlinuz-"

func found(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func Dracut() bool {
	return !found("update-initramfs") && found("dracut")
}

func Folder() string {
	return env.Or("SOLANUM_BOOT", "/boot")
}

func Kernels() []string {
	output, _ := run.CaptureQuiet(run.Minute, "ls", "-v", "--quoting-style=literal", Folder())
	var kernels []string
	for _, name := range run.Lines(output) {
		version, ok := strings.CutPrefix(name, kernelPrefix)
		if ok && !strings.Contains(version, ".dpkg-") && !strings.HasPrefix(version, tpm.Before("")) {
			kernels = append(kernels, version)
		}
	}
	return kernels
}

func Installed(kernel string) bool {
	return slices.Contains(Kernels(), kernel)
}

func Newest() string {
	kernels := Kernels()
	if len(kernels) == 0 {
		return ""
	}
	return kernels[len(kernels)-1]
}

func Image(kernel string) string {
	return filepath.Join(Folder(), "initrd.img-"+kernel)
}

func Rebuild(all bool) []string {
	switch {
	case Dracut():
		if kernel := Newest(); kernel != "" {
			return []string{"dracut", "--force", Image(kernel), kernel}
		}
		return []string{"dracut", "--force"}
	case all:
		return []string{"update-initramfs", "-u", "-k", "all"}
	}
	return []string{"update-initramfs", "-u"}
}
