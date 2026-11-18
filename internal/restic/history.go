package restic

import (
	"bufio"
	"cmp"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Ahmadlolbit/solanumos/internal/run"
)

const WhenLayout = "2 Jan 2006 15.04"

type Entry struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Size int64  `json:"size"`
	Time string `json:"time"`
}

type node struct {
	Name   string    `json:"name"`
	Type   string    `json:"type"`
	Path   string    `json:"path"`
	Size   int64     `json:"size"`
	Mtime  time.Time `json:"mtime"`
	Struct string    `json:"struct_type"`
}

func kindOf(mode fs.FileMode) string {
	switch {
	case mode.IsDir():
		return "dir"
	case mode&fs.ModeSymlink != 0:
		return "link"
	}
	return "file"
}

func Sorted(entries []Entry) []Entry {
	slices.SortStableFunc(entries, func(first, second Entry) int {
		return cmp.Or(cmp.Compare(boolRank(first.Kind != "dir"), boolRank(second.Kind != "dir")), cmp.Compare(strings.ToLower(first.Name), strings.ToLower(second.Name)))
	})
	return entries
}

func boolRank(later bool) int {
	if later {
		return 1
	}
	return 0
}

func Listed(dir string) ([]Entry, error) {
	found, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	entries := []Entry{}
	for _, item := range found {
		if info, err := item.Info(); err == nil {
			entries = append(entries, Entry{item.Name(), kindOf(info.Mode()), info.Size(), info.ModTime().Format(time.RFC3339)})
		}
	}
	return Sorted(entries), nil
}

func Entries(output, folder string) []Entry {
	entries := []Entry{}
	lines := bufio.NewScanner(strings.NewReader(output))
	lines.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for lines.Scan() {
		var found node
		if json.Unmarshal(lines.Bytes(), &found) != nil || found.Struct != "node" || filepath.Dir(found.Path) != folder || found.Path == folder {
			continue
		}
		kind := map[string]string{"dir": "dir", "symlink": "link"}[found.Type]
		entries = append(entries, Entry{found.Name, cmp.Or(kind, "file"), found.Size, found.Mtime.Format(time.RFC3339)})
	}
	return Sorted(entries)
}

func Beside(path, words string, folder bool) string {
	dir, name := filepath.Split(path)
	extension := filepath.Ext(name)
	if folder || extension == name {
		extension = ""
	}
	stem := strings.TrimSuffix(name, extension)
	for count := 1; ; count++ {
		mark := words
		if count > 1 {
			mark += " " + strconv.Itoa(count)
		}
		candidate := filepath.Join(dir, stem+" ("+mark+")"+extension)
		if _, err := os.Lstat(candidate); err != nil {
			return candidate
		}
	}
}

func copied(source, target string) error {
	return run.Logged(os.Stderr, "cp", "-a", "--reflink=auto", "-T", "--", source, target)
}

func PutBack(source, target, when string, folder, asCopy bool) (string, string, error) {
	if asCopy {
		named := Beside(target, "from "+when, folder)
		return named, "", copied(source, named)
	}
	aside := ""
	if _, err := os.Lstat(target); err == nil {
		aside = Beside(target, "before the restore", folder)
		if err := run.Logged(os.Stderr, "mv", "-n", "-T", "--", target, aside); err != nil {
			return "", "", err
		}
	}
	if err := copied(source, target); err != nil {
		if aside != "" {
			run.Logged(os.Stderr, "mv", "-n", "-T", "--", aside, target)
		}
		return "", "", err
	}
	return target, aside, nil
}

func Inside(path, home string) bool {
	clean, top := filepath.Clean(path), filepath.Clean(home)
	return filepath.IsAbs(path) && clean != top && strings.HasPrefix(clean+"/", top+"/")
}
