package setup

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/bootimage"
	"github.com/Ahmadlolbit/solanumos/internal/config"
	"github.com/Ahmadlolbit/solanumos/internal/detect"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/grub"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/runlog"
	"github.com/Ahmadlolbit/solanumos/internal/service"
)

const (
	kernelPackage  = "linux-image-amd64"
	keptOption     = "KEPT_KERNEL"
	kernelSpec     = kernelPackage + "/backports"
	packageFiles   = "/var/lib/dpkg/info/linux-image-"
	nvidiaSettings = "/etc/modprobe.d/solanum-nvidia.conf"
	hybridPower    = " NVreg_DynamicPowerManagement=0x02"
	kernelKept     = "The backports kernel stays installed, because the setup never removes a kernel."
	nvidiaLegacy   = "This NVIDIA card is older than Maxwell, and Debian 13's NVIDIA driver no longer supports it, so it stays on nouveau, the open driver."
	kernelBack     = "To go back to Debian's own: sudo apt install linux-image-amd64/trixie linux-headers-amd64/trixie " +
		"(leave out the headers if you don't have them) and say yes to going back, restart into Debian's kernel from " +
		"Advanced options in the boot menu, then sudo apt remove the backports kernel: dpkg -l 'linux-image-*' | grep bpo shows its name."
)

var nvidiaUnits = []string{"nvidia-suspend.service", "nvidia-resume.service", "nvidia-hibernate.service"}

type nvidiaDriver struct {
	*Context
	hardware  detect.Hardware
	initramfs followUp
}

func (context *Context) kernelSpecs(specs ...string) []string {
	if context.Config.Get("KERNEL") != "stable" {
		return specs
	}
	plain := make([]string, len(specs))
	for index, spec := range specs {
		plain[index] = apt.Bare(spec)
	}
	return plain
}

func backportsKernel() bool {
	return strings.Contains(apt.Version(kernelPackage), "bpo")
}

func (context *Context) driverSpecs(specs ...string) []string {
	if backportsKernel() {
		return specs
	}
	return context.kernelSpecs(specs...)
}

func (context *Context) kernelInstall(log *runlog.Log) error {
	return apt.Install(log, context.kernelSpecs(kernelSpec)...)
}

func (context *Context) kernelNotes() []string {
	if context.Config.Get("KERNEL") != "stable" || !backportsKernel() {
		return nil
	}
	return []string{kernelKept, kernelBack}
}

func newestKernel() string {
	return bootimage.Newest()
}

func bootedAt() int64 {
	text, _ := fsutil.Text("/proc/uptime")
	whole, _, _ := strings.Cut(text, ".")
	seconds, _ := strconv.ParseInt(whole, 10, 64)
	return time.Now().Unix() - seconds
}

func changedSinceBoot(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && info.ModTime().Unix() > bootedAt()
}

func installedSinceBoot(kernel string) bool {
	return changedSinceBoot(packageFiles+kernel+".list") || changedSinceBoot(grub.Menu())
}

func kernelThere(kernel string) bool {
	return bootimage.Installed(kernel)
}

func keptGone(kept string) error {
	return fmt.Errorf("the kept kernel, %s, isn't installed: keep another with sudo solanum kernel keep, or start the newest with sudo solanum kernel newest", kept)
}

func keptBooted(log io.Writer, kept, newest, running string) error {
	switch {
	case running == kept && newest != kept:
		fmt.Fprintf(log, "this is the kept kernel; %s is newer, and sudo solanum kernel newest starts it\n", newest)
	case running == kept:
		fmt.Fprintln(log, "this is the kept kernel")
	case !kernelThere(kept):
		return keptGone(kept)
	case changedSinceBoot(grub.Menu()):
		fmt.Fprintf(log, "reboot to start the kept kernel, %s\n", kept)
	default:
		return errors.New("this boot did not use the kept kernel, " + kept + "; check the default GRUB entry")
	}
	return nil
}

func (context *Context) kernelBooted(log *runlog.Log) error {
	newest := newestKernel()
	running, _ := run.CaptureQuiet(run.Minute, "uname", "-r")
	fmt.Fprintf(log, "newest installed: %s, running: %s\n", cmp.Or(newest, "none"), running)
	if kept := cmp.Or(context.Config.Get(keptOption), grub.HandMade()); kept != "" {
		return keptBooted(log, kept, newest, running)
	}
	if newest == "" || newest == running {
		return nil
	}
	if !installedSinceBoot(newest) {
		return errors.New("this boot did not use the newest kernel; check the default GRUB entry")
	}
	fmt.Fprintf(log, "reboot to start %s\n", newest)
	return nil
}

func (context *Context) kernelChosen() bool {
	return context.Config.Get(keptOption) != "" || grub.HandMade() != "" || grub.Kept() != ""
}

func (context *Context) takeOverKept(log *runlog.Log) error {
	byHand := grub.HandMade()
	if byHand == "" || context.Config.Get(keptOption) != "" {
		return nil
	}
	saved := config.Read(config.Path())
	saved.Set(keptOption, byHand)
	if err := saved.Save(config.Path()); err != nil {
		return err
	}
	context.Config.Set(keptOption, byHand)
	log.Line("info", "the kernel kept by hand, "+byHand+", is the Kept kernel setting now")
	return nil
}

func (context *Context) kernelStart(log *runlog.Log) error {
	if err := context.takeOverKept(log); err != nil {
		return err
	}
	kept := context.Config.Get(keptOption)
	switch {
	case kept == "":
		return grub.Newest(log)
	case !kernelThere(kept):
		return keptGone(kept)
	}
	return grub.Keep(log, kept)
}

func kernelFromBackports(*runlog.Log) error {
	version := apt.Version(kernelPackage)
	return expect(strings.Contains(version, "bpo"), kernelPackage+" is at version "+version)
}

func nvidiaBuilt(*runlog.Log) error {
	kernel := newestKernel()
	return quietly("there's no nvidia module for "+kernel, "modinfo", "-k", kernel, "nvidia")
}

func (driver *nvidiaDriver) chosen() bool {
	return driver.hardware.HasGPU("nvidia") && driver.Config.Get("GRAPHICS") == "proprietary"
}

func (driver *nvidiaDriver) wanted() bool {
	return driver.chosen() && !driver.hardware.NvidiaLegacy
}

func (driver *nvidiaDriver) install(log *runlog.Log) error {
	list := "nvidia-closed"
	if driver.hardware.NvidiaOpen {
		list = "nvidia-open"
	}
	listed, err := apt.List(log, list)
	if err != nil {
		return err
	}
	return apt.Install(log, driver.driverSpecs(listed...)...)
}

func (driver *nvidiaDriver) configure(log *runlog.Log) error {
	power := ""
	if driver.hardware.Hybrid() {
		power = hybridPower
	}
	changed, err := driver.Deploy.Changed("system/modprobe/nvidia.conf", nvidiaSettings, "POWER="+power)
	driver.initramfs.need(changed)
	if err != nil {
		return err
	}
	if err := service.Enable(log, nvidiaUnits...); err != nil {
		return err
	}
	rebuild := bootimage.Rebuild(true)
	return driver.initramfs.do(func() error { return run.Logged(log, rebuild[0], rebuild[1:]...) })
}

func (context *Context) webcamWanted() bool {
	return context.Config.Get("WEBCAM") == "auto" && context.hardware().IPU6
}

func (context *Context) webcamInstall(log *runlog.Log) error {
	listed, err := apt.List(log, "webcam")
	if err != nil {
		return err
	}
	return apt.Install(log, context.driverSpecs(listed...)...)
}

func webcamReady(*runlog.Log) error {
	return apt.Commands("cam", "gst-launch-1.0")
}

func kernelSteps(context *Context) {
	driver := nvidiaDriver{Context: context, hardware: context.hardware(), initramfs: followUp{name: "nvidia-initramfs"}}
	context.Run.Step("Install the "+context.Config.Get("KERNEL")+" kernel", context.kernelInstall)
	if context.kernelChosen() {
		context.Run.Step("Set the kernel the computer starts", context.kernelStart)
	}
	for _, note := range context.kernelNotes() {
		context.Run.Note(note)
	}
	if driver.chosen() && driver.hardware.NvidiaLegacy {
		context.Run.Note(nvidiaLegacy)
	}
	if driver.wanted() {
		context.Run.Step("Install the NVIDIA driver", driver.install)
		context.Run.Step("Configure the NVIDIA driver", driver.configure)
	}
	if context.webcamWanted() {
		context.Run.Step("Set up the IPU6 webcam", context.webcamInstall)
	}
}

func kernelChecks(context *Context) {
	driver := nvidiaDriver{Context: context, hardware: context.hardware()}
	if context.Config.Get("KERNEL") == "backports" {
		context.Run.Check("The kernel comes from backports", kernelFromBackports)
	}
	context.Run.Check("The newest kernel is running (or starts after a reboot)", context.kernelBooted)
	if driver.wanted() {
		context.Run.Check("The NVIDIA module is built for the newest kernel", nvidiaBuilt)
	}
	if context.webcamWanted() {
		context.Run.Check("The IPU6 webcam's tools are there", webcamReady)
	}
}
