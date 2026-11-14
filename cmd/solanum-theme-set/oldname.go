package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/oldname"
	"github.com/Ahmadlolbit/solanumos/internal/run"
)

func oldLinks() []string {
	data, config := env.DataHome(), env.ConfigHome()
	return []string{
		filepath.Join(data, "color-schemes", oldname.Back("Solanum.colors")),
		filepath.Join(data, "nvim", "site", "colors", oldname.Name+".lua"),
		filepath.Join(config, "walker", "themes", oldname.Name, "style.css"),
		filepath.Join(config, "btop", "themes", oldname.Name+".theme"),
	}
}

func oldFolders() []string {
	data, config := env.DataHome(), env.ConfigHome()
	return []string{
		filepath.Join(data, "themes", oldname.Prefix+"one"),
		filepath.Join(data, "themes", oldname.Prefix+"two"),
		filepath.Join(config, "zed", "themes", oldname.Name+".json"),
	}
}

func dropOld(path string, remove func(string) error) {
	if err := remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(os.Stderr, err)
	}
}

func forgetOldNames() {
	for _, link := range oldLinks() {
		if ourLink(link) {
			dropOld(link, os.Remove)
		}
	}
	os.Remove(filepath.Join(env.ConfigHome(), "walker", "themes", oldname.Name))
	for _, path := range oldFolders() {
		dropOld(path, os.RemoveAll)
	}
}

func dropOldFirefox(chrome string) {
	dropOld(filepath.Join(chrome, oldname.FirefoxStyle), os.Remove)
	rewrite(filepath.Join(chrome, "userChrome.css"), withoutLine(oldname.FirefoxImport))
}

func (editor codeEditor) dropOldExtension() {
	found, _ := filepath.Glob(filepath.Join(env.Home(), editor.extensions, oldname.Extension+"-*"))
	if len(found) > 0 && run.Quiet(run.Minute, editor.command, "--uninstall-extension", oldname.Extension) != 0 {
		fmt.Fprintf(os.Stderr, "%s kept the %s themes\n", editor.command, oldname.Name)
	}
}
