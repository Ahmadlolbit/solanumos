package disks

import (
	"slices"
	"strings"
)

var keySystems = []string{"vfat", "exfat", "ntfs", "ntfs3", "ext2", "ext3", "ext4", "btrfs", "f2fs"}

type Stick struct {
	Disk       Disk
	Path       string
	Mountpoint string
}

func (stick Stick) Name() string {
	if stick.Path == stick.Disk.Path {
		return stick.Disk.Name()
	}
	return stick.Disk.Name() + " · " + strings.TrimPrefix(stick.Path, "/dev/")
}

func firstMount(points []string) string {
	for _, point := range points {
		if point != "" && point != "[SWAP]" {
			return point
		}
	}
	return ""
}

func (disk Disk) holdsKeys(target string) bool {
	return disk.External() && !disk.ReadOnly && disk.Path != target && !disk.Holds("/run/live/medium") && !disk.Holds("/")
}

func Sticks(all []Disk, target string) []Stick {
	var found []Stick
	for _, disk := range all {
		if !disk.holdsKeys(target) {
			continue
		}
		if slices.Contains(keySystems, disk.FSType) {
			found = append(found, Stick{disk, disk.Path, firstMount(disk.Mountpoint)})
		}
		for _, part := range disk.Children {
			if slices.Contains(keySystems, part.FSType) {
				found = append(found, Stick{disk, part.Path, firstMount(part.Mountpoints)})
			}
		}
	}
	return found
}
