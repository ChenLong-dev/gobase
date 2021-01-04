/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:11:47
 * @LastEditTime: 2020-12-17 09:11:47
 * @LastEditors: Chen Long
 * @Reference:
 */

package mbase

import (
	"flag"
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase/msys"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
	"github.com/ChenLong-dev/gobase/mlog"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

var mlogLevel string
var MLogLevel = mlog.DEBUG //日志级别
var SystemTest *bool
var gParams = &mlog.Params{}

func setupSigusr1Trap(logpath string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	go func() {
		for range c {
			ioutil.WriteFile(logpath+".stack", []byte(mutils.DumpStack(true)), 0666)
			//mlog.Debugf("%s", mutils.DumpStack(true))
		}
	}()
}
func cycleFreeMem() {
	for {
		time.Sleep(time.Hour)
		debug.FreeOSMemory()
	}
}

func SetParamsMaxBackups(backUps int) {
	if backUps >= 0 {
		gParams.MaxBackups = backUps
	}
}

func init() {
	flag.StringVar(&mlogLevel, "mlog_level", "debug", "mlog level, can all/trace/debug/info/warn/error")
	SystemTest = flag.Bool("systemTest", false, "Set to true when running system tests") //用于 gotest 系统测试的参数

	//	1.mbase.conf配置
	initMBaseConf()

	//	2.设置日志
	logDir := GetMBaseConf().LogDir + "/" + msys.ProcessName() + "/"
	os.MkdirAll(logDir, os.ModePerm)
	gParams.Path = logDir + msys.ProcessName() + ".log"
	gParams.Level = MLogLevel
	gParams.DisableStdOut = true
	gParams.DisableUdpLog = true
	gParams.WorkIpString = GetMBaseConf().LogDir
	gParams.ProcessId = msys.ProcessId()
	gParams.ProcessName = msys.ProcessName()
	mlog.Init(gParams)
	//	3.设置系统运行参数
	msys.SetMaxFdSize(1000000)
	msys.EnableCore()

	setupSigusr1Trap(gParams.Path) //	安装信号处理
	go cycleFreeMem()
}

func Init() error {

	enStdoutLog := flag.Bool("enableStdoutLog", false, "enable stdout log")
	pV := flag.Bool("v", false, "build time") /*放在最后*/

	flag.Parse()

	if *pV {
		fmt.Println("Build Time:", mutils.BuildTime())
		os.Exit(0) /*查看版本号直接退出*/
	}

	switch mlogLevel {
	case "all":
		MLogLevel = mlog.ALL
	case "trace":
		MLogLevel = mlog.TRACE
	case "debug":
		MLogLevel = mlog.DEBUG
	case "info":
		MLogLevel = mlog.INFO
	case "warn":
		MLogLevel = mlog.WARN
	case "error":
		MLogLevel = mlog.ERROR
	}

	mlog.SetLevel(MLogLevel)

	if *enStdoutLog {
		mlog.SwitchStdout(true)
	} else {
		//stderr := GetMBaseConf().LogDir + "/" + msys.ProcessName() + "/stderr/" + msys.ProcessName() + ".stderr"
		redirectPath := GetMBaseConf().LogDir + "/" + msys.ProcessName() + "/redirectStd/" + msys.ProcessName() + ".std"
		if err := mutils.RedirectStdToFile(redirectPath, 20); err != nil {
			mlog.Errorf("RedirectStdToFile failed. err=%s", err)
		}
	}

	return nil
}
