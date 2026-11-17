package setup

import (
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/runlog"
)

func softwareDir() string {
	return filepath.Join(env.ShareDir(), "software")
}

func softwareReady(*runlog.Log) error {
	return fileThere(filepath.Join(softwareDir(), "common", "Theme.qml"))
}

func (context *Context) softwareInstall(*runlog.Log) error {
	if err := context.Deploy.Tree("software", softwareDir()); err != nil {
		return err
	}
	return context.Deploy.Link("../shell/common", filepath.Join(softwareDir(), "common"))
}
