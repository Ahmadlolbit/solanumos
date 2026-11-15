package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/Ahmadlolbit/solanumos/internal/apt"
	"github.com/Ahmadlolbit/solanumos/internal/disks"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/hypr"
	"github.com/Ahmadlolbit/solanumos/internal/run"
	"github.com/Ahmadlolbit/solanumos/internal/theme"
)

const (
	release   = "SolanumOS Chert"
	kibPerGiB = 1 << 20
	shortDate = "2 January 2006"
)

var (
	placeholders = []string{"", "none", "default string", "not applicable", "not specified", "o.e.m.", "oem", "system product name", "system version", "to be filled by o.e.m.", "type1productconfigid", "unknown"}
	makers       = map[string]string{"Intel Corporation": "Intel", "Advanced Micro Devices, Inc. [AMD/ATI]": "AMD", "NVIDIA Corporation": "NVIDIA"}
	graphicsKind = []string{"VGA compatible controller", "3D controller", "Display controller"}
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

func system(path string) string {
	return strings.TrimSpace(fsutil.Optional(os.Getenv("SOLANUM_SYSROOT") + path))
}

func field(text, key string) string {
	for _, line := range strings.Split(text, "\n") {
		if label, value, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(label) == key {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func given(value string) bool {
	return !slices.Contains(placeholders, strings.ToLower(value))
}

func model() string {
	name := system("/sys/class/dmi/id/product_name")
	if version := system("/sys/class/dmi/id/product_version"); strings.EqualFold(system("/sys/class/dmi/id/sys_vendor"), "lenovo") && given(version) {
		name = version
	}
	if !given(name) {
		return ""
	}
	return name
}

func processor() string {
	return strings.Join(strings.Fields(strings.NewReplacer("(R)", "", "(TM)", "", "(tm)", "").Replace(field(system("/proc/cpuinfo"), "model name"))), " ")
}

func memory() string {
	kib, err := strconv.ParseInt(strings.TrimSuffix(field(system("/proc/meminfo"), "MemTotal"), " kB"), 10, 64)
	if err != nil || kib <= 0 {
		return ""
	}
	return fmt.Sprintf("%d GiB", (kib+kibPerGiB-1)/kibPerGiB)
}

func disk() string {
	var stat syscall.Statfs_t
	if syscall.Statfs(os.Getenv("SOLANUM_SYSROOT")+"/", &stat) != nil || stat.Blocks == 0 {
		return ""
	}
	return disks.Size(int64(stat.Blocks)*int64(stat.Bsize)) + ", " + disks.Size(int64(stat.Blocks-stat.Bfree)*int64(stat.Bsize)) + " used"
}

func card(line string) string {
	fields := strings.Split(line, "\"")
	if len(fields) < 6 || !slices.Contains(graphicsKind, fields[1]) {
		return ""
	}
	maker, device := fields[3], fields[5]
	if short, ok := makers[maker]; ok {
		maker = short
	}
	if open := strings.LastIndex(device, "["); open >= 0 && strings.HasSuffix(device, "]") {
		device = device[open+1 : len(device)-1]
	}
	if strings.HasPrefix(device, maker) {
		return device
	}
	return maker + " " + device
}

func graphics() string {
	output, err := run.CaptureQuiet(run.Quick, "lspci", "-mm")
	if err != nil {
		return ""
	}
	var cards []string
	for _, line := range run.Lines(output) {
		if name := card(line); name != "" {
			cards = append(cards, name)
		}
	}
	return strings.Join(cards, ", ")
}

func screens() (string, string) {
	monitors, _ := hypr.QuietMonitors()
	var sizes []string
	for _, monitor := range monitors {
		wide, high := monitor.Width, monitor.Height
		if monitor.Transform%2 == 1 {
			wide, high = high, wide
		}
		sizes = append(sizes, fmt.Sprintf("%d × %d (%s)", wide, high, monitor.Name))
	}
	if len(sizes) > 1 {
		return "Screens", strings.Join(sizes, ", ")
	}
	return "Screen", strings.Join(sizes, "")
}

func kernel() string {
	if version := system("/proc/sys/kernel/osrelease"); version != "" {
		return "Linux " + version
	}
	return ""
}

func hyprland() string {
	version := strings.TrimSpace(apt.Version("hyprland"))
	if _, upstream, epoch := strings.Cut(version, ":"); epoch {
		version = upstream
	}
	if end := strings.IndexAny(version, "+~-"); end >= 0 {
		version = version[:end]
	}
	return version
}

func themeInUse() string {
	name, _ := fsutil.Text(filepath.Join(env.ThemeState(), "theme.name"))
	if name = strings.TrimSpace(name); name == "" {
		return ""
	}
	return theme.Pretty(name)
}

func installed() string {
	info, err := os.Stat(os.Getenv("SOLANUM_SYSROOT") + "/etc/machine-id")
	if err != nil {
		return ""
	}
	return info.ModTime().Format(shortDate)
}

func Facts() About {
	about := About{Name: release, Version: env.Version, Debian: system("/etc/debian_version"), Facts: []Fact{}}
	screen, sizes := screens()
	for _, fact := range []Fact{{"Model", model()}, {"Processor", processor()}, {"Memory", memory()}, {"Disk", disk()}, {"Graphics", graphics()},
		{screen, sizes}, {"Kernel", kernel()}, {"Hyprland", hyprland()}, {"Theme", themeInUse()}, {"Installed", installed()}} {
		if fact.Value != "" {
			about.Facts = append(about.Facts, fact)
		}
	}
	return about
}
