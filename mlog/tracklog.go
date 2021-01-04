/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 12:37:11
 * @LastEditTime: 2020-12-16 14:06:06
 * @LastEditors: Chen Long
 * @Reference:
 */

package mlog

import (
	"fmt"
	"net"
)

var (
	trackLogConn net.Conn
)

func tackLogInit() error {
	//addr := fmt.Sprintf("%s:%d", "tracklog.service", stdParams.TrackLogPort)
	addr := fmt.Sprintf("%s:%d", "127.0.0.1", stdParams.TrackLogPort)
	//fmt.Println(addr)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return err
	}
	trackLogConn = conn
	return nil
}

func TrackLog(level, format string, a ...interface{}) error {
	if trackLogConn == nil {
		err := tackLogInit()
		if err != nil {
			return err
		}
	}

	prefix := fmt.Sprintf("[%s %s:%s:%d]: ", level, stdParams.WorkIpString, stdParams.ProcessName, stdParams.ProcessId)
	msg := fmt.Sprintf(format, a...)

	b := strings2Cbytes(prefix, msg)
	_, err := trackLogConn.Write(b)
	return err
}
