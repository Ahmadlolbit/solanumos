package move

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/account"
	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/oldname"
)

var tools = []string{"core", "look", "screenshot", "theme-set", "theme-install", "background", "care", "follow-theme", "installer", "pick", "pkg", "record", "screensaver", "sensors", "start", "terminal", "tip", "weather", "webapp"}

func swaps() [][2]string {
	name := oldname.Name
	pairs := [][2]string{
		{"/usr/share/" + name + "/hypr/" + name + ".lua", "/usr/share/solanum/hypr/solanum.lua"},
		{"/usr/share/" + name + "/zsh/" + name + ".zsh", "/usr/share/solanum/zsh/solanum.zsh"},
		{"/usr/share/" + name + "/", "/usr/share/solanum/"},
		{"/etc/" + name + "/", "/etc/solanum/"},
		{".local/state/" + name + "/", ".local/state/solanum/"},
		{".config/" + name + "/", ".config/solanum/"},
		{oldname.AddressKey, "X-Solanum-Address="},
		{`"update.` + name + `"`, `"update.solanum"`},
		{`colorscheme, "` + name + `"`, `colorscheme, "solanum"`},
	}
	for _, tool := range tools {
		pairs = append(pairs, [2]string{oldname.Prefix + tool, "solanum-" + tool})
	}
	return pairs
}

func commandSwaps() [][2]string {
	name := oldname.Name
	return append(swaps(), [2]string{" " + name + " ", " solanum "}, [2]string{`"` + name + " ", `"solanum `}, [2]string{"'" + name + " ", "'solanum "})
}

func replacer(pairs [][2]string, back bool) *strings.Replacer {
	var flat []string
	for _, pair := range pairs {
		if back {
			flat = append(flat, pair[1], pair[0])
		} else {
			flat = append(flat, pair[0], pair[1])
		}
	}
	return strings.NewReplacer(flat...)
}

func PersonText(text string) string {
	return replacer(swaps(), false).Replace(text)
}

func menuText(text string) string {
	return replacer(commandSwaps(), false).Replace(text)
}

func fingerprint(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func readable(path string) (string, string, bool) {
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", "", false
	}
	info, err := os.Lstat(real)
	if err != nil || !info.Mode().IsRegular() || info.Size() > textLimit {
		return "", "", false
	}
	data, err := os.ReadFile(real)
	return real, string(data), err == nil
}

func (work *mover) rewrite(person account.Account, path string, change func(string) string) {
	real, before, ok := readable(path)
	if !ok {
		return
	}
	after := change(before)
	if after == before {
		return
	}
	if work.writeAs(person, real, after) {
		work.rehash(path, before, after)
	}
}

func (work *mover) writeAs(person account.Account, path, text string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	part := path + ".solanum-part"
	if err := person.Write(work.out, part, text); err != nil {
		work.note("%s keeps the old names: %v", path, err)
		return false
	}
	return work.asPerson(person, "chmod", strconv.FormatUint(uint64(info.Mode().Perm()), 8), part) && work.asPerson(person, "mv", "-fT", part, path)
}

func (work *mover) rehash(path, before, after string) {
	database := filepath.Join(stateDir(), "files.db")
	lines, err := fsutil.Lines(database)
	if err != nil {
		return
	}
	plain := strings.TrimPrefix(path, strings.TrimSuffix(at("/"), "/"))
	entry := plain + "\t" + fingerprint(before)
	changed := false
	for index, line := range lines {
		if line == entry {
			lines[index], changed = plain+"\t"+fingerprint(after), true
		}
	}
	if changed {
		atomic.Write(database, []byte(strings.Join(lines, "\n")+"\n"), databaseMode)
	}
}

func (work *mover) rewriteFiles(person account.Account) {
	for _, rel := range personFiles {
		change := PersonText
		if filepath.Base(rel) == "menu.json" {
			change = menuText
		}
		work.rewrite(person, at(filepath.Join(person.Home, rel)), change)
	}
	apps, _ := filepath.Glob(at(filepath.Join(person.Home, ".local/share/applications/webapp-*.desktop")))
	for _, app := range apps {
		work.rewrite(person, app, PersonText)
	}
	plugins := at(filepath.Join(person.Home, ".config/nvim/lua/plugins"))
	old, fresh := filepath.Join(plugins, oldname.Name+".lua"), filepath.Join(plugins, "solanum.lua")
	if fsutil.IsFile(old) && !fsutil.Exists(fresh) {
		work.rewrite(person, old, PersonText)
		work.moveOwn(person, old, fresh)
	}
}
