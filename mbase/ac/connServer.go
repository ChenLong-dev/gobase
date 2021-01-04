/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:10:36
 * @LastEditTime: 2020-12-16 14:10:36
 * @LastEditors: Chen Long
 * @Reference:
 */

package ac

import (
	"net"
	"sync/atomic"

	"github.com/ChenLong-dev/gobase/mbase/mcom"
	"github.com/ChenLong-dev/gobase/mlog"
)

type NewConnCb func(conn *Conn)

type ConnServer struct {
	l    net.Listener
	cb   NewConnCb
	addr string
	en   int32
}

func (cs *ConnServer) run() {
	l := cs.l
	for atomic.LoadInt32(&cs.en) != 0 {
		c, cerr := l.Accept()
		if cerr != nil {
			mlog.Warnf("accept error:%v", cerr)
			continue
		}

		tcpConn, ok := c.(*net.TCPConn)
		if ok {
			/*tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(time.Second * 20)*/
			_ = mcom.SetTcpKeepAlive(tcpConn, 60, 30, 6)
			_ = tcpConn.SetNoDelay(true)
		}

		conn := NewConn(c, false)
		if cs.cb != nil {
			go cs.cb(conn)
		} else {
			_ = conn.Close()
		}
	}
}
func (cs *ConnServer) Close() {
	if cs != nil && atomic.CompareAndSwapInt32(&cs.en, 0, 1) {
		l := cs.l
		cs.l = nil
		if l != nil {
			_ = l.Close()
		}
	}
}

func RunConnServer(addr string, cb NewConnCb) (cs *ConnServer, err error) {
	mlog.Tracef("addr=%s", addr)
	defer func() { mlog.Tracef("err=%v", err) }()

	l, lerr := net.Listen("tcp", addr)
	if lerr != nil {
		return nil, lerr
	}
	cs = &ConnServer{l: l, cb: cb, addr: addr, en: 1}

	go cs.run()
	return cs, nil
}
