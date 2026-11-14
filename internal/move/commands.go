package move

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/oldname"
)

const programMode = 0o755

func (work *mover) place(name, target string) error {
	if fsutil.Exists(target) {
		return nil
	}
	if name == "solanum" {
		return atomic.Symlink(filepath.Join(env.Root(), "solanum"), target)
	}
	source := filepath.Join(env.ToolsDir(), name)
	if program, err := os.Readlink(source); err == nil {
		if err := work.place(program, filepath.Join(filepath.Dir(target), program)); err != nil {
			return err
		}
		return atomic.Symlink(program, target)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return atomic.Install(target, data, programMode)
}

func (work *mover) moveCommand(bin, name string) {
	old := filepath.Join(bin, name)
	if !ours(strings.TrimPrefix(old, strings.TrimSuffix(at("/"), "/"))) {
		return
	}
	fresh := oldname.Swap(name)
	if err := work.place(fresh, filepath.Join(bin, fresh)); err != nil {
		work.note("Left %s as it is: %s couldn't take its place (%v)", old, fresh, err)
		return
	}
	if target, err := os.Readlink(old); err == nil && target == fresh {
		return
	}
	if err := atomic.Symlink(fresh, old); err != nil {
		work.note("Left %s as it is: %v", old, err)
	}
}

func (work *mover) moveCommands() {
	bin := at("/usr/local/bin")
	for _, entry := range fsutil.Entries(bin) {
		if name := entry.Name(); name == oldname.Name || strings.HasPrefix(name, oldname.Prefix) {
			work.moveCommand(bin, name)
		}
	}
	fmt.Fprintln(work.out, "the old command names now run the new ones")
}
