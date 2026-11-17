package setup

import (
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/runlog"
)

func careWindowDir() string {
	return filepath.Join(env.ShareDir(), "care")
}

func careWindowReady(*runlog.Log) error {
	return fileThere(filepath.Join(careWindowDir(), "common", "Theme.qml"))
}

func (context *Context) careWindowInstall(*runlog.Log) error {
	if err := context.Deploy.Tree("care", careWindowDir()); err != nil {
		return err
	}
	return context.Deploy.Link("../shell/common", filepath.Join(careWindowDir(), "common"))
}
