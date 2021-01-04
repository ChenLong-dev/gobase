/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:26:45
 * @LastEditTime: 2020-12-26 15:47:30
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcom

import (
	"bytes"
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase"
	"github.com/ChenLong-dev/gobase/mbase/msys"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
	"github.com/ChenLong-dev/gobase/mlog"
	"github.com/ChenLong-dev/gobase/sspb"
	"github.com/golang/protobuf/proto"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"
)

type MdbgPower int

const (
	MdbgPowerAll  MdbgPower = 0
	MdbgPowerRoot MdbgPower = 1
)

type MdbgCallback func(cmdReq string) (result int32, cmdRsp string)
type MdbgMmlCallback func(mml *mutils.Mml) (result int32, cmdRsp string)

type mdbgCmd struct {
	name  string
	help  string
	limit MdbgPower
	cb    MdbgCallback
}

type Mdbg struct {
	srv   *MServerListener
	cmd   map[string]*mdbgCmd
	pprof net.Listener
}

func (md *Mdbg) init() {
	RegisterMdbg(MdbgPowerAll, "setlog", "setlog trace,debug,info,warn,error,fatal,all", mdbgSetLog)
	RegisterMdbgMml(MdbgPowerRoot, "pprof", "pprof cmd={start,stop,dump}[,port=?,dumpPath=%s]", mdbgPprof)
	RegisterMdbgMml(MdbgPowerRoot, "pprofCpu", "pprofCpu cmd={start,stop}[,savepath=%s]", mdbgPprofCpu)
	RegisterMdbg(MdbgPowerRoot, "freeMem", "free memory to OS", mdbgFreeMem)
	RegisterMdbg(MdbgPowerRoot, "stack", "stack", mdbgStack)
	RegisterMdbg(MdbgPowerAll, "help", "help", mdbgHelp)
	RegisterMdbg(MdbgPowerRoot, "panic", "panic", mdbgPanic)
	RegisterMdbg(MdbgPowerRoot, "exit", "exit", mdbgExit)
}

var sMdbg = &Mdbg{
	srv:   nil,
	cmd:   make(map[string]*mdbgCmd),
	pprof: nil,
}

func InitMdbg(listenAddr string) error {
	if sMdbg.srv != nil {
		return nil
	}

	if srv, err := MServerListen(NewMContext(), "tcp", listenAddr, nil, packetFunc, nil); err != nil {
		return err
	} else {
		sMdbg.srv = srv
	}
	sMdbg.init()
	return nil
}

func RegisterMdbg(limit MdbgPower, cmdName, cmdHelp string, cb MdbgCallback) {
	if _, ok := sMdbg.cmd[cmdName]; ok {
		return
	}

	cmd := &mdbgCmd{
		name:  cmdName,
		help:  cmdHelp,
		limit: limit,
		cb:    cb,
	}
	sMdbg.cmd[cmdName] = cmd
}
func RegisterMdbgMml(limit MdbgPower, cmdName, cmdHelp string, cb MdbgMmlCallback) {
	f := func(cmdReq string) (result int32, cmdRsp string) {
		mml, _ := mutils.NewMml(cmdReq, ",")
		return cb(&mml)
	}

	RegisterMdbg(limit, cmdName, cmdHelp, f)
}

func mdbgSetLog(cmdReq string) (int32, string) {

	level := cmdReq

	switch level {
	case "trace":
		mlog.SetLevel(mlog.TRACE)
	case "debug":
		mlog.SetLevel(mlog.DEBUG)
	case "info":
		mlog.SetLevel(mlog.INFO)
	case "warn":
		mlog.SetLevel(mlog.WARN)
	case "error":
		mlog.SetLevel(mlog.ERROR)
	case "fatal":
		mlog.SetLevel(mlog.FATAL)
	case "all":
		mlog.SetLevel(mlog.ALL)
	default:
		return 0, fmt.Sprintf("Not support level = %s!", level)
	}
	return 0, fmt.Sprintf("Set log to %s success!", level)
}

func mdbgPprofDump(dumpPath string) (int32, string) {
	err := os.MkdirAll( /*"/home/moa/log/pprof"*/ dumpPath, 0777)
	if err != nil {
		return 0, fmt.Sprintln(err)
	}

	/*params := strings.Split(os.Args[0], "/")
	fileName := fmt.Sprintf("/home/moa/log/pprof/%s_%s.mprof", params[len(params)-1],
		time.Now().Format("20060102150405"))*/

	mlog.Trace(dumpPath)

	fm, err := os.OpenFile(dumpPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		mlog.Warn(err)
	}

	runtime.GC()

	pprof.WriteHeapProfile(fm)
	fm.Close()

	return 0, fmt.Sprintf("Dump heap profilw succss!")
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
func mdbgPprofListenStart(port int32) (int32, string) {

	if sMdbg.pprof != nil {
		return 0, fmt.Sprintf("pprof has been listenning, [%s]", sMdbg.pprof.Addr())
	}

	if port == 0 {
		port = 19999
	}

	pprofAddr := fmt.Sprintf("localhost:%d", port)

	pprof, err := net.Listen("tcp", pprofAddr)
	if err != nil {
		return 0, fmt.Sprintf("listen and server failed, addr=%s, err=%s", pprofAddr, err)
	}

	sMdbg.pprof = pprof

	go func() {
		srv := &http.Server{Addr: pprofAddr, Handler: nil}
		if err := srv.Serve(tcpKeepAliveListener{sMdbg.pprof.(*net.TCPListener)}); err != nil {
			sMdbg.pprof = nil
			mlog.Debugf("listen and server failed, addr=%s, err=", pprofAddr, err)
		}
	}()

	return 0, fmt.Sprintf("pprof is listenning [%s]", sMdbg.pprof.Addr())
}

func mdbgPprofListenStop() (int32, string) {
	if sMdbg.pprof == nil {
		return 0, "not pprof is listenning"
	}

	if err := sMdbg.pprof.Close(); err != nil {
		return 9, fmt.Sprintln("close pprof failed, addr=", sMdbg.pprof.Addr(), "err=", err)
	}

	return 0, fmt.Sprintf("stop pprof listenning, [%s]", sMdbg.pprof.Addr())
}

func mdbgPprof(mml *mutils.Mml) (int32, string) {

	cmd := mml.GetString("cmd", "unknow")
	port := mml.GetInt32("port", 0)
	dumpPath := mml.GetString("dumpPath", "./"+msys.ProcessName()+".mprof")

	switch cmd {
	case "dump":
		return mdbgPprofDump(dumpPath)
	case "start":
		return mdbgPprofListenStart(port)
	case "stop":
		return mdbgPprofListenStop()
	default:
		return 0, fmt.Sprintf("Not support cmd =%s", cmd)
	}
}

var mdbgPprofCpuFile *os.File = nil
var mdbgPprofCpuStartTime time.Time

func mdbgPprofCpuStart(savepath string) (int32, string) {
	if mdbgPprofCpuFile != nil {
		return -1, fmt.Sprintf("alread start at %v", mdbgPprofCpuStartTime)
	}
	f, err := os.OpenFile(savepath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return -1, fmt.Sprintf("open(%s) error:%v", savepath, err)
	}

	err = pprof.StartCPUProfile(f)
	if err != nil {
		f.Close()
		return -1, fmt.Sprintf("StartCPUProfile error:%v", err)
	}

	mdbgPprofCpuFile = f
	mdbgPprofCpuStartTime = time.Now()
	return 0, fmt.Sprintf("ok, start cpu pprof to %s [%s]", savepath, mdbgPprofCpuStartTime)
}
func mdbgPprofCpuStop() (int32, string) {
	if mdbgPprofCpuFile == nil {
		return -1, fmt.Sprintf("error: no cpu pprof started!")
	}
	pprof.StopCPUProfile()
	savepath := mdbgPprofCpuFile.Name()
	now := time.Now()

	mdbgPprofCpuFile.Close()
	mdbgPprofCpuFile = nil

	return 0, fmt.Sprintf("ok, stop cpu pprof to %s. [%ds: %v]", savepath, now.Sub(mdbgPprofCpuStartTime)/time.Second, mdbgPprofCpuStartTime)
}
func mdbgPprofCpu(mml *mutils.Mml) (int32, string) {
	cmd := mml.GetString("cmd", "unknow")
	savepath := mml.GetString("savepath", "./"+msys.ProcessName()+".pprof")

	switch cmd {
	case "start":
		return mdbgPprofCpuStart(savepath)
	case "stop":
		return mdbgPprofCpuStop()
	default:
		return 0, fmt.Sprintf("Not support cmd =%s", cmd)
	}
}

func mdbgFreeMem(cmdReq string) (int32, string) {
	debug.FreeOSMemory()
	return 0, "done!"
}

func mdbgStack(cmdReq string) (int32, string) {

	buf := make([]byte, 1024*1024)
	n := runtime.Stack(buf, true)
	s := string(buf[:n])
	s += "\n"

	return 0, s
}

func mdbgHelp(cmdReq string) (int32, string) {

	var buf bytes.Buffer

	limit_str := func(l MdbgPower) string {
		switch l {
		case MdbgPowerRoot:
			return "root"
		case MdbgPowerAll:
			return "all"
		}
		return "unknow"
	}

	for _, v := range sMdbg.cmd {
		buf.WriteString(fmt.Sprintf("cmd = %s, help = %s, limit = %s\n", v.name, v.help, limit_str(v.limit)))
	}

	return 0, buf.String()
}

func mdbgPanic(cmq string) (int32, string) {
	panic("panic")
	return 0, "success"
}

func mdbgExit(cmq string) (int32, string) {
	mbase.MKrtEixt(0)
	return 0, "success"
}

func packetFunc(mctx *MContext, msrv *MServer, data []byte) {
	mctx.Tracef("len(data)=%d", len(data))
	if mctx.OpCode() != int32(sspb.PBOpCode_OPC_MBASE_MDBG_REQ) {
		mlog.Debugf("Not support cmd %v", mctx)
		return
	}

	req := &sspb.PBMDbgReq{}
	if err := proto.Unmarshal(data, req); err != nil {
		mctx.Tracef("unmarshal error:%v", err)
		return
	}

	mctx.Tracef("req=%v", req)

	var cmd *mdbgCmd
	var ok bool
	var result int32
	var cmdRsp string

	if cmd, ok = sMdbg.cmd[req.GetCmdname()]; !ok {
		result = 0
		cmdRsp = fmt.Sprintf("Not find cmd %s!!", req.GetCmdname())
	} else {
		result, cmdRsp = cmd.cb(req.GetCmdreq())
	}

	//do response
	rsp := &sspb.PBMDbgRsp{
		Result: result,
		Cmdrsp: []byte(cmdRsp),
	}

	mctx.Tracef("rsp=%v", rsp)

	msrv.Reply(mctx, nil, rsp)
}
