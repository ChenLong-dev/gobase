/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:38:22
 * @LastEditTime: 2020-12-16 14:38:22
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcom

import (
	"github.com/ChenLong-dev/gobase/mlog"
	"net"
	"time"
)

func SetTcpKeepAlive(conn *net.TCPConn, idle, intvl, probs int) (err error) {
	defer func() { mlog.Tracef("conn=%v,idle=%d,intvl=%d,probs=%d,err=%v", conn, idle, intvl, probs, err) }()
	if err = conn.SetKeepAlive(true); err != nil {
		return err
	}
	if err = conn.SetKeepAlivePeriod(time.Second * time.Duration(intvl)); err != nil {
		return err
	}

	return err
}
