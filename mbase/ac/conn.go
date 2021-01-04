/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:09:12
 * @LastEditTime: 2020-12-16 14:09:12
 * @LastEditors: Chen Long
 * @Reference:
 */

package ac

import (
	"encoding/binary"
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase/mcom"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
	"github.com/ChenLong-dev/gobase/mlog"
	"net"
	"sync/atomic"
)

const (
	MaxTlvPacketSize uint32 = 20 * 1024 * 1024
)

type Conn struct {
	net.Conn
	aWriter    atomic.Value
	closed     int32
	IsClient   bool
	localAddr  net.Addr
	remoteAddr net.Addr
}

func (conn *Conn) String() string {
	if conn == nil || conn.Conn == nil {
		return fmt.Sprintf("{Conn:nil}")
	} else {
		return fmt.Sprintf("{Local:%v,Remote:%v,Avail:%t,IsClient:%t}",
			conn.localAddr, conn.remoteAddr, conn.Avail(), conn.IsClient)
	}
}

// LocalAddr returns the local network address.
func (conn *Conn) LocalAddr() net.Addr {
	return conn.localAddr
}

// RemoteAddr returns the remote network address.
func (conn *Conn) RemoteAddr() net.Addr {
	return conn.remoteAddr
}
func (conn *Conn) Close() error {
	if !atomic.CompareAndSwapInt32(&conn.closed, 0, 1) {
		return nil
	}
	/*c := conn.Conn
	conn.Conn = nil
	if c != nil {
		c.Close()
	}*/
	if conn.Conn != nil {
		conn.Conn.Close()
	}

	if asyncWriter := conn.asyncWriter(); asyncWriter != nil {
		asyncWriter.Close()
	}

	return nil
}
func (conn *Conn) asyncWriter() *mcom.AsyncWriter {
	if v := conn.aWriter.Load(); v != nil {
		return v.(*mcom.AsyncWriter)
	}
	return nil
}
func (conn *Conn) EnableAsyncWriter() {
	if conn.asyncWriter() == nil {
		conn.aWriter.Store(mcom.NewAsyncWriter(conn.Conn))
	}
}
func (conn *Conn) Avail() bool {
	return conn != nil && atomic.LoadInt32(&conn.closed) == 0 && conn.Conn != nil
}
func (conn *Conn) Send(pk []byte) (err error) {
	mlog.Tracef("conn=%v,len(pk)=%d", conn, len(pk))
	defer func() { mlog.Tracef("conn=%v,err=%v", conn, err) }()

	if len(pk) == 0 || !conn.Avail() {
		return nil
	}

	if asyncWriter := conn.asyncWriter(); asyncWriter != nil {
		_, err = asyncWriter.Write(pk)
	} else {
		_, err = mutils.WriteN(conn.Conn, pk)
	}
	return err
}
func (conn *Conn) Recv(pk []byte) (n int, err error) {
	mlog.Tracef("conn=%v,len(pk)=%d", conn, len(pk))
	defer func() { mlog.Tracef("n=%d,err=%v", n, err) }()

	if !conn.Avail() {
		return 0, fmt.Errorf("conn is invalid")
	}
	return conn.Read(pk)
}
func (conn *Conn) SendTlvPacket(tp *TlvPacket) (err error) {
	pk := tp.Marshal()
	mlog.Tracef("conn=%v,len(pk)=%d", conn, len(pk))
	defer func() { mlog.Tracef("conn=%v,err=%v", conn, err) }()

	if len(pk) == 0 || !conn.Avail() {
		return nil
	}
	//_, err = mutils.WriteN(conn.Conn, pk)
	//return err
	return conn.Send(pk)
}
func (conn *Conn) RecvTlvPacket() (tp *TlvPacket, err error) {
	mlog.Tracef("conn=%v", conn)
	defer func() { mlog.Tracef("conn=%v,err=%v", conn, err) }()

	if !conn.Avail() {
		return nil, fmt.Errorf("conn invalid")
	}

	dlen, herr := mutils.ReadUint32(conn.Conn)
	if herr != nil {
		return nil, fmt.Errorf("read head error:%v", herr)
	}
	if dlen > MaxTlvPacketSize {
		return nil, fmt.Errorf("read data len(%d) over max packet size(%d)", dlen, MaxTlvPacketSize)
	}

	pk := make([]byte, 4+dlen)
	binary.LittleEndian.PutUint32(pk, dlen)
	_, err = mutils.ReadN(conn.Conn, pk[4:])
	if err != nil {
		return nil, err
	}

	return UnmarshalTlvPacket(pk)
}
func (ms *Conn) SetKeepAlive(en bool, idle, intvl, probs int) {
	if !ms.Avail() {
		return
	}
	tcpConn, ok := ms.Conn.(*net.TCPConn)
	if ok {
		if en {
			mcom.SetTcpKeepAlive(tcpConn, idle, intvl, probs)
		} else {
			tcpConn.SetKeepAlive(false)
		}
	}
}

func NewConn(conn net.Conn, isClient bool) *Conn {
	c := &Conn{Conn: conn, IsClient: isClient}
	if conn != nil {
		c.localAddr = conn.LocalAddr()
		c.remoteAddr = conn.RemoteAddr()
	}
	return c
}
