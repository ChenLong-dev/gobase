/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 12:36:11
 * @LastEditTime: 2020-12-16 12:36:11
 * @LastEditors: Chen Long
 * @Reference:
 */

package mlog

import (
	"fmt"
	"net"
)

var (
	oplogConn    net.Conn
	prefixFormat = "[%s %s (%d,%d)]: "
)

func strings2Cbytes(strs ...string) []byte {
	if len(strs) == 0 {
		return nil
	}

	size := 0
	for _, str := range strs {
		size += len(str)
	}
	size++ // 0
	b := make([]byte, size)

	size = 0
	for _, str := range strs {
		for i := 0; i < len(str); i++ {
			b[size] = str[i]
			size++
		}
	}
	return b
}

func opLogInit() error {
	addr := fmt.Sprintf("%s:%d", stdParams.WorkIpString, stdParams.OPLogPort)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return err
	}
	oplogConn = conn
	return nil
}

func OPLOG(op string, did, pid int64, format string, a ...interface{}) error {
	if oplogConn == nil {
		err := opLogInit()
		if err != nil {
			return err
		}
	}

	prefix := fmt.Sprintf(prefixFormat, stdParams.ProcessName, op, did, pid)
	msg := fmt.Sprintf(format, a...)
	b := strings2Cbytes(prefix, msg)

	_, err := oplogConn.Write(b)
	return err
}
