package upgradeAgent

import (
	"github.com/ChenLong-dev/gobase/mlog"
	"os"
)

func mkUpgradeDir(workdir string, p *PackageInfo) string {
	return workdir + p.Ver +"_" + p.SHA256 + "/"
}
func rmUpgradeDir(dir string) error {
	_, err := Exec("rm -rf \"" + dir + "\"")
	return err
}
func createUpgradePackagePath(dir string) string {
	rmUpgradeDir(dir)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		mlog.Warnf("mkdir(%s) error:%v", dir, err)
	}

	return dir + "package.tgz"
}