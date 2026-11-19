package setup

import (
	"path/filepath"

	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/runlog"
)

const (
	passwordHelper = "solanum-backup-password"
	backupPolicy   = "/usr/share/polkit-1/actions/org.solanumos.backup.policy"
)

func careWindowDir() string {
	return filepath.Join(env.ShareDir(), "care")
}

func careWindowReady(*runlog.Log) error {
	if err := fileThere(filepath.Join(careWindowDir(), "common", "Theme.qml")); err != nil {
		return err
	}
	if err := fileThere(env.Or("SOLANUM_BACKUP_POLICY", backupPolicy)); err != nil {
		return err
	}
	helper := filepath.Join(env.BinDir(), passwordHelper)
	return expect(fsutil.Runnable(helper), helper+" won't run")
}

func (context *Context) careWindowInstall(*runlog.Log) error {
	if err := context.Deploy.Tree("care", careWindowDir()); err != nil {
		return err
	}
	if err := context.Deploy.Link("../shell/common", filepath.Join(careWindowDir(), "common")); err != nil {
		return err
	}
	if err := context.Deploy.Tool(passwordHelper); err != nil {
		return err
	}
	return context.Deploy.File("system/polkit/org.solanumos.backup.policy", env.Or("SOLANUM_BACKUP_POLICY", backupPolicy), "HELPER="+filepath.Join(env.BinDir(), passwordHelper))
}
