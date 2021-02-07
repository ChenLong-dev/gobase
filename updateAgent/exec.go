package upgradeAgent

import (
	"github.com/ChenLong-dev/gobase/mlog"
	"os/exec"
	"path/filepath"
)

func Exec(cmdline string) (output string, err error) {
	cmd := exec.Command("/bin/bash", "-c", cmdline)
	o, e := cmd.CombinedOutput()
	if len(o) > 0 {
		return string(o), e
	}
	return "", e
}

func ExecUpgrade(packagePath string, cmdline string) (err error) {
	mlog.Infof("now exec upgrade command(%s), packagePath(%s)", cmdline, packagePath)
	workdir := filepath.Dir(packagePath)
	packageName := filepath.Base(packagePath)
	cdWorkdir := "cd " +  workdir + ";"
	if cmdline == "" {
		cmdline = cdWorkdir + "tar -xzf ./" + packageName + "; cd package; ./install.sh"
	} else {
		cmdline = cdWorkdir + cmdline
	}

	output, oerr := Exec(cmdline)
	mlog.Infof("exec(%s) result:%v,output:%s", cmdline, oerr, output)
	return oerr
}