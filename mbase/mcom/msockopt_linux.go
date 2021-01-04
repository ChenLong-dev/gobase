/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:37:39
 * @LastEditTime: 2020-12-26 15:48:41
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcom

import (
	"net"
	"syscall"
	"time"

	"github.com/ChenLong-dev/gobase/mlog"
)

/*
func SetTcpKeepAlive(fd *netFD, d time.Duration) error {
	// The kernel expects seconds so round to next highest second.
	d += (time.Second - time.Nanosecond)
	secs := int(d.Seconds())
	if err := fd.pfd.SetsockoptInt(syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, secs); err != nil {
		return wrapSyscallError("setsockopt", err)
	}
	err := fd.pfd.SetsockoptInt(syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, secs)
	runtime.KeepAlive(fd)
	return wrapSyscallError("setsockopt", err)
}*/

func SetTcpKeepAlive(conn *net.TCPConn, idle, intvl, probs int) (err error) {
	defer func() { mlog.Tracef("conn=%v,idle=%d,intvl=%d,probs=%d,err=%v", conn, idle, intvl, probs, err) }()
	if err = conn.SetKeepAlive(true); err != nil {
		return err
	}
	if err = conn.SetKeepAlivePeriod(time.Second * time.Duration(intvl)); err != nil {
		return err
	}

	rawconn, rerr := conn.SyscallConn()
	if rerr != nil {
		return rerr
	}
	err = rawconn.Control(func(ufd uintptr) {
		fd := int(ufd)

		syscall.SetsockoptInt(fd, syscall.SOL_TCP, syscall.TCP_KEEPIDLE, idle)
		syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, probs)
	})

	return err
}
