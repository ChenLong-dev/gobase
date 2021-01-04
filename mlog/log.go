/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 12:34:40
 * @LastEditTime: 2020-12-16 12:36:30
 * @LastEditors: Chen Long
 * @Reference:
 */

package mlog

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"runtime"
)

//日志级别
const (
	ALL   = iota
	TRACE // 跟踪, 1
	DEBUG // 调试， 2
	INFO  // 信息，3
	WARN  // 警告，4
	ERROR // 一般错误,5
	FATAL // 致命错误，6
)

var (
	level        int
	enablestdout bool
	enableUdpLog bool
)

func getLevelStr(level int) string {
	switch level {
	case ALL:
		return "A"
	case TRACE:
		return "T"
	case DEBUG:
		return "D"
	case INFO:
		return "I"
	case WARN:
		return "W"
	case ERROR:
		return "E"
	case FATAL:
		return "F"
	default:
		return "NAN"
	}
}
func getPrefix(level int, calldep int) string {
	/*
		_, file, line, ok := runtime.Caller(2)
		if !ok {
			return fmt.Sprintf("[%d] ******cannot get statck******", level)
		}
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
	*/
	getRealName := func(file string) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		return short
	}

	pc, file, line, _ := runtime.Caller(calldep)
	f := runtime.FuncForPC(pc)

	//return fmt.Sprintf("[%s] (%s) ", getLevelStr(level), f.Name())
	return fmt.Sprintf("[%s] %s:%d (%s): ", getLevelStr(level), getRealName(file), line, getRealName(f.Name()))
}

type Params struct {
	Path          string //	路径
	MaxSize       int    //	MB
	MaxBackups    int    //	备份个数
	MaxAge        int    //	保存时间,天
	DisableStdOut bool   //	是否标准输出
	Level         int
	OPLogPort     uint16
	TrackLogPort  uint16
	WorkIpString  string
	ProcessId     int
	ProcessName   string
	DisableUdpLog bool
}

var (
	stdParams *Params
	stdo      = log.New(os.Stdout, "", log.LstdFlags)
	logOutput = log.New(os.Stdout, "", log.LstdFlags)
)

func Init(params *Params) error {
	level = params.Level
	enablestdout = !params.DisableStdOut
	enableUdpLog = !params.DisableUdpLog
	//	path := fmt.Sprintf("%s_%d_%d.log", c.Log.Path, os.Getpid(), time.Now().Unix())
	//	if c.Log.Path != "stdout" && c.Log.MaxSize != 0 {
	stdParams = params
	if len(params.Path) > 0 {
		logger := &lumberjack.Logger{
			Filename:   params.Path,
			MaxSize:    params.MaxSize,
			MaxBackups: params.MaxBackups,
			MaxAge:     params.MaxAge,
			LocalTime:  true,
		}

		if logger.MaxBackups == 0 {
			logger.MaxBackups = 1
		}
		if logger.MaxAge == 0 {
			logger.MaxAge = 30
		}

		//log.SetOutput(logger)
		logOutput = log.New(logger, "", log.LstdFlags)
	}
	//log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	//log.SetFlags(log.Ldate | log.Ltime)

	//	log.SetFlags(0)
	return nil
}
func SwitchStdout(en bool) {
	enablestdout = en
}
func SwitchUdp(en bool) {
	enableUdpLog = en
}
func logPrintln(prefix string, v ...interface{}) {
	//log.SetPrefix(prefix)
	//log.Output(3, prefix+fmt.Sprint(v...))
	logOutput.Output(3, prefix+fmt.Sprint(v...))
}
func logPrintf(prefix string, format string, v ...interface{}) {
	//log.SetPrefix(prefix)
	//log.Output(3, prefix+fmt.Sprintf(format, v...))
	logOutput.Output(3, prefix+fmt.Sprintf(format, v...))
}

func logStdout(prefix string, v ...interface{}) {
	if enablestdout {
		//fmt.Printf("%s%s\n", prefix, fmt.Sprint(v...))
		//stdo.SetPrefix(prefix)
		stdo.Output(3, prefix+fmt.Sprint(v...))
	}
}

func logStdoutf(prefix string, format string, v ...interface{}) {
	if enablestdout {
		//fmt.Printf("%s%s\n", prefix, fmt.Sprintf(format, v...))
		//stdo.SetPrefix(prefix)
		stdo.Output(3, prefix+fmt.Sprintf(format, v...))
	}
}

func logTrack(level int, prefix string, v ...interface{}) {
	if enableUdpLog {
		logTrackf(level, prefix, fmt.Sprint(v...))
	}
}

func logTrackf(level int, prefix, format string, v ...interface{}) {
	if enableUdpLog {
		TrackLog(getLevelStr(level), prefix+format, v...)
	}
}

func Output(level int, calldepth int, ctx string, format string, v ...interface{}) {
	prefix := getPrefix(level, calldepth)

	txt := prefix + ctx + fmt.Sprintf(format, v...)

	//log.SetPrefix(prefix)
	logOutput.Output(calldepth, txt)

	//stdo.SetPrefix(prefix)
	stdo.Output(calldepth, txt)

	logTrackf(level, txt, "")
}

func Trace(v ...interface{}) {
	if level > TRACE {
		return
	}

	prefix := getPrefix(TRACE, 2)
	logPrintln(prefix, v...)
	logStdout(prefix, v...)
	logTrack(TRACE, prefix, v...)
}
func Tracef(format string, v ...interface{}) {
	if level > TRACE {
		return
	}

	prefix := getPrefix(TRACE, 2)
	logPrintf(prefix, format, v...)
	logStdoutf(prefix, format, v...)
	logTrackf(TRACE, prefix, format, v...)
}
func ForceTracef(format string, v ...interface{}) {
	prefix := getPrefix(TRACE, 2)
	logPrintf(prefix, format, v...)
	logStdoutf(prefix, format, v...)
	logTrackf(TRACE, prefix, format, v...)
}
func Debug(v ...interface{}) {
	if level > DEBUG {
		return
	}

	prefix := getPrefix(DEBUG, 2)
	logPrintln(prefix, v...)
	logStdout(prefix, v...)
}
func Debugf(format string, v ...interface{}) {
	if level > DEBUG {
		return
	}

	prefix := getPrefix(DEBUG, 2)
	logPrintf(prefix, format, v...)
	logStdoutf(prefix, format, v...)
}
func Info(v ...interface{}) {
	if level > INFO {
		return
	}

	prefix := getPrefix(INFO, 2)
	logPrintln(prefix, v...)
	logStdout(prefix, v...)
	//logTrack(INFO, prefix, v...)
}
func Infof(format string, v ...interface{}) {
	if level > INFO {
		return
	}

	prefix := getPrefix(INFO, 2)
	logPrintf(prefix, format, v...)
	logStdoutf(prefix, format, v...)
	//logTrackf(INFO, prefix, format, v...)
}
func Warn(v ...interface{}) {
	if level > WARN {
		return
	}

	prefix := getPrefix(WARN, 2)
	logPrintln(prefix, v...)
	logStdout(prefix, v...)
	logTrack(WARN, prefix, v...)
}
func Warnf(format string, v ...interface{}) {
	if level > WARN {
		return
	}

	prefix := getPrefix(WARN, 2)
	logPrintf(prefix, format, v...)
	logStdoutf(prefix, format, v...)
	logTrackf(WARN, prefix, format, v...)
}
func Error(v ...interface{}) {
	if level > ERROR {
		return
	}

	prefix := getPrefix(ERROR, 2)
	logPrintln(prefix, v...)
	logStdout(prefix, v...)
	logTrack(ERROR, prefix, v...)
}
func Errorf(format string, v ...interface{}) {
	if level > ERROR {
		return
	}

	prefix := getPrefix(ERROR, 2)
	logPrintf(prefix, format, v...)
	logStdoutf(prefix, format, v...)
	logTrackf(ERROR, prefix, format, v...)
}
func Fatal(v ...interface{}) {
	if level > FATAL {
		return
	}
	prefix := getPrefix(FATAL, 2)
	logPrintln(prefix, v...)
	logStdout(prefix, v...)
	logTrack(FATAL, prefix, v...)
	//log.SetPrefix(prefix)
	log.Panic(v...)
}
func Fatalf(format string, v ...interface{}) {
	if level > FATAL {
		return
	}

	prefix := getPrefix(FATAL, 2)
	logPrintf(prefix, format, v...)
	logStdoutf(prefix, format, v...)
	logTrackf(FATAL, prefix, format, v...)
	//log.SetPrefix(prefix)
	log.Panicf(format, v...)
}
func Stack(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	s += "\n"
	buf := make([]byte, 1024*1024)
	n := runtime.Stack(buf, true)
	s += string(buf[:n])
	s += "\n"

	logPrintln(getPrefix(level, 2), s)
	logStdout(getPrefix(level, 2), s)
}

func SetLevel(l int) {
	level = l
}

func GetLevel() int {
	return level
}
