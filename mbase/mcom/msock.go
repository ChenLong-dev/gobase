/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:35:14
 * @LastEditTime: 2020-12-26 15:48:34
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcom

import (
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
	"github.com/ChenLong-dev/gobase/mlog"
	"net"
	"sync/atomic"
	"time"
)

type MSockEntryFunc func(mctx *MContext, ms *MSock)
type MSockReconnectFunc func(mctx *MContext, ms *MSock)
type MSockHandleFunc func(mctx *MContext, ms *MSock, data []byte)
type MSockErrorFunc func(mctx *MContext, ms *MSock, err error)

const (
	MSockStatus_CONNECTING   int32 = 0 //	连接中
	MSockStatus_ESTABLISHED  int32 = 1 //	建立状态
	MSockStatus_DISCONNECTED int32 = 2 //	断开状态
	MSockStatus_CLOSED       int32 = 3 //	已经关闭
)

func statusName(status int32) string {
	switch status {
	case MSockStatus_CONNECTING:
		return "CONNECTING"
	case MSockStatus_ESTABLISHED:
		return "ESTABLISHED"
	case MSockStatus_DISCONNECTED:
		return "DISCONNECTED"
	case MSockStatus_CLOSED:
		return "CLOSED"
	default:
		return fmt.Sprintf("unknown(%d)", status)
	}
}

type msockSendData struct {
	conn  net.Conn
	mctx  *MContext
	datas [][]byte
}
type MSock struct {
	//conn         net.Conn
	conn atomic.Value
	//sendC 		 chan *msockSendData
	reconnNotify MSockReconnectFunc
	handler      MSockHandleFunc
	errorNotify  MSockErrorFunc
	isServer     bool
	disableGo    bool
	compress     bool
	//status 		atomic.Value
	status   int32
	cryptKey []byte
	pdata    atomic.Value
}

func (ms *MSock) GetPrivateData() interface{} {
	return ms.pdata.Load()
}

func (ms *MSock) SetPrivateData(pdata interface{}) {
	if pdata == nil {
		ms.pdata = atomic.Value{}
	} else {
		ms.pdata.Store(pdata)
	}
}

func (ms *MSock) String() string {
	if ms == nil {
		return "MSock{nil}"
	}

	var local, peer string
	if ms.Status() == MSockStatus_ESTABLISHED {
		local, peer = ms.LocalAddr().String(), ms.RemoteAddr().String()
	}
	return fmt.Sprintf("MSock{local:%s,peer:%s,isServer:%t,disableGo:%t,compress:%t,status:%s}",
		local, peer, ms.isServer, ms.disableGo, ms.compress, statusName(ms.Status()))
}

type MSockListener struct {
	listener    net.Listener
	handler     MSockHandleFunc
	entryNotify MSockEntryFunc
	errorNotify MSockErrorFunc
	disableGo   bool
	compress    bool
}

type msockOptions struct {
	disableGo bool
	compress  bool
}

type MSockOption struct {
	f func(*msockOptions)
}

func MSockOptDisableGo() MSockOption {
	return MSockOption{f: func(do *msockOptions) {
		do.disableGo = true
	}}
}
func MSockOptCompress() MSockOption {
	return MSockOption{f: func(do *msockOptions) {
		do.compress = true
	}}
}

func msockTcpOpt(ms *MSock) error {
	tcpConn, ok := ms.Conn().(*net.TCPConn)
	if ok {
		/*tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(time.Second * 20)*/
		SetTcpKeepAlive(tcpConn, 60, 30, 6)
		tcpConn.SetNoDelay(true)
	}
	return nil
}
func (ms *MSock) SetKeepAlive(en bool, idle, intvl, probs int) {
	if ms == nil {
		return
	}
	tcpConn, ok := ms.Conn().(*net.TCPConn)
	if ok {
		if en {
			SetTcpKeepAlive(tcpConn, idle, intvl, probs)
		} else {
			tcpConn.SetKeepAlive(false)
		}
	}
}
func (ms *MSock) msockHandleError(mctx *MContext, err error) {
	addr := ms.Conn().RemoteAddr()

	mctx.Tracef("ms=%v,network=%s,addr=%s,err=%v", ms, addr.Network(), addr.String(), err)

	if ms.setStatus(MSockStatus_DISCONNECTED) == MSockStatus_DISCONNECTED {
		return
	}

	if ms.errorNotify != nil {
		if ms.disableGo {
			ms.errorNotify(mctx, ms, err)
		} else {
			go ms.errorNotify(mctx, ms, err)
		}
	}

	if ms.isServer {
		ms.Close(mctx)
	} else {
		for ms.Status() == MSockStatus_DISCONNECTED { //	客户端自动重连
			time.Sleep(time.Second) //重连sleep 1s
			conn, nerr := net.Dial(addr.Network(), addr.String())
			mctx.Tracef("ms=%v,Dial(%s,%s) error=%v", ms, addr.Network(), addr.String(), nerr)
			if nerr == nil {
				oldConn := ms.Conn()
				ms.setConn(conn)
				if oldConn != nil {
					oldConn.Close()
				}
				oldStatus := ms.setStatus(MSockStatus_ESTABLISHED)
				if oldStatus == MSockStatus_CLOSED {
					ms.setStatus(MSockStatus_CLOSED)
					ms.Conn().Close()
					//ms.conn = atomic.Value{}
					//ms.setConn(nil)
				} else if ms.reconnNotify != nil {
					if ms.disableGo {
						ms.reconnNotify(mctx, ms)
					} else {
						go ms.reconnNotify(mctx, ms)
					}
				}
			}
		}
	}
}
func msockRecvN(conn net.Conn, data []byte) error {
	//mbase.Tracef("ms=%v,len(data)=%d", ms, len(data))

	needRecv := len(data)

	for needRecv > 0 {
		rbuf := data[len(data)-needRecv:]
		n, err := conn.Read(rbuf)
		mbase.Tracef("n=%d,err=%v", n, err)
		if err != nil {
			return err
		}

		needRecv -= n
	}

	return nil
}
func checkMFrame(mframe *MFrame) bool {
	if mframe == nil {
		return false
	}
	if int(mframe.Len) <= MFrameSize || mframe.Len > 1000000000 {
		return false
	}
	if mframe.Mark != MFrameMark {
		return false
	}
	return true
}

/*
func (ms *MSock) msockHandleSend() {
	mbase.Tracef("")

	var n int
	var err error
	c := ms.sendC
	for sd := range c {
		for _, d := range sd.datas {
			n, err = mutils.WriteN(sd.conn, d)
			sd.mctx.Tracef("WriteN(%d) bytes return n=%d,err=%v", len(d), n, err)
		}
	}
}*/
func (ms *MSock) msockHandleRecv() {
	mbase.Tracef("")

	for ms.Status() != MSockStatus_CLOSED {
		conn := ms.Conn()
		//	1.MFrame
		data := make([]byte, 8, 512)
		err := msockRecvN(conn, data)
		if err != nil {
			ms.msockHandleError(NewMContext(), err)
			continue
		}

		mframe, data := UnmarshalMFrame(data)
		if !checkMFrame(mframe) {
			ms.msockHandleError(NewMContext(), mbase.NewMError(mbase.MERR_MBASE_ILLEGAL))
			continue
		}
		mctx := NewMContextWithTrace(HasMFrameFlag(mframe, MFRAME_FLAG_TRACELOG))

		//	2.whole
		data = make([]byte, mframe.Len-uint32(MFrameSize))
		err = msockRecvN(conn, data)
		if err != nil {
			ms.msockHandleError(mctx, err)
			continue
		}
		mctx.Tracef("%v recv a packet(%d)", ms, mframe.Len)

		//	3.umcompress
		if HasMFrameFlag(mframe, MFRAME_FLAG_COMPRESS) {
			if ud, uderr := mutils.GZipUnCompress(data); uderr == nil {
				data = ud
			}
			mctx.SetCompress(true)
		}

		//	4.handle
		if ms.handler != nil {
			if ms.disableGo {
				ms.handler(mctx, ms, data)
			} else {
				go ms.handler(mctx, ms, data)
			}
		}
	}
}
func (ms *MSock) msockHandle() {
	mbase.Trace(ms)

	msockTcpOpt(ms)

	//ms.sendC = make(chan *msockSendData, mbase.GetMBaseConf().SendChanMax)
	//go ms.msockHandleSend()
	ms.msockHandleRecv()
}

func (ms *MSock) msockHandleConn(mctx *MContext, entryNotify MSockEntryFunc) {
	mctx.Tracef("%v", ms)

	if entryNotify != nil {
		if ms.disableGo {
			entryNotify(mctx, ms)
		} else {
			go entryNotify(mctx, ms)
		}
	}

	ms.msockHandle()
}
func (ml *MSockListener) msockServer(mctx *MContext) {
	mbase.Trace(ml)

	for {
		c, err := ml.listener.Accept()
		if err != nil {
			mlog.Warnf("accept error:", err)
			break
		}
		// start a new goroutine to handle
		// the new connection.
		ms := &MSock{handler: ml.handler, errorNotify: ml.errorNotify, isServer: true, disableGo: ml.disableGo, compress: ml.compress /*, status:MSockStatus_ESTABLISHED*/}
		ms.setConn(c)
		ms.setStatus(MSockStatus_ESTABLISHED)

		go ms.msockHandleConn(mctx, ml.entryNotify)
	}
}

func MSockListen(mctx *MContext, proto, addr string, entryNotify MSockEntryFunc, handler MSockHandleFunc, errorNotify MSockErrorFunc, opts ...MSockOption) (*MSockListener, error) {
	mlog.Infof("proto=%s,addr=%s", proto, addr)

	l, err := net.Listen(proto, addr)
	if err != nil {
		mlog.Warnf("Listen(%s,%s) error=%v", proto, addr, err)
		return nil, err
	}

	msockOpts := &msockOptions{}
	for _, opt := range opts {
		opt.f(msockOpts)
	}

	ml := &MSockListener{
		listener: l,
		handler:  handler, entryNotify: entryNotify, errorNotify: errorNotify,
		disableGo: msockOpts.disableGo,
		compress:  msockOpts.compress,
	}

	go ml.msockServer(mctx)

	return ml, nil
}

func (ml *MSockListener) Close() {
	mlog.Info(ml)

	ml.listener.Close()
}

func (ms *MSock) msockHandleDial(mctx *MContext, proto, addr string) {
	for ms.Status() == MSockStatus_CONNECTING {
		time.Sleep(time.Second)
		if ms.Status() == MSockStatus_CONNECTING {
			c, err := net.Dial(proto, addr)
			if err == nil {
				ms.setConn(c)
				ms.setStatus(MSockStatus_ESTABLISHED)

				if ms.reconnNotify != nil {
					if ms.disableGo {
						ms.reconnNotify(mctx, ms)
					} else {
						go ms.reconnNotify(mctx, ms)
					}
				}

				break
			}
		}
	}

	if ms.Status() == MSockStatus_ESTABLISHED {
		ms.msockHandle()
	}
}

func MSockDial(mctx *MContext, proto, addr string, reconnNotify MSockReconnectFunc, handler MSockHandleFunc, errNotify MSockErrorFunc, opts ...MSockOption) (*MSock, error) {
	mctx.Tracef("proto=%s,addr=%s", proto, addr)

	msockOpts := &msockOptions{}
	for _, opt := range opts {
		opt.f(msockOpts)
	}

	ms := &MSock{
		reconnNotify: reconnNotify, handler: handler, errorNotify: errNotify,
		isServer: false, disableGo: msockOpts.disableGo, compress: msockOpts.compress, /*, status:MSockStatus_ESTABLISHED*/
	}

	ms.setStatus(MSockStatus_CONNECTING)

	c, err := net.Dial(proto, addr)
	if err == nil {
		//ms.conn = c
		ms.setConn(c)
		ms.setStatus(MSockStatus_ESTABLISHED)
	}

	go ms.msockHandleDial(mctx, proto, addr)

	return ms, nil
}
func (ms *MSock) Close(mctx *MContext) {
	mctx.Tracef("%v", ms)

	if ms.setStatus(MSockStatus_CLOSED) != MSockStatus_CLOSED {
		conn := ms.Conn()
		//ms.conn = atomic.Value{}
		if conn != nil {
			conn.Close()
		}
		//close(ms.sendC)
		//ms.sendC = nil
	}
	/*
		ms.mutex.Lock()
		defer ms.mutex.Unlock()

		if ms.Status() != MSockStatus_CLOSED {
			ms.status.Store(MSockStatus_CLOSED)
			conn := ms.conn
			ms.conn = nil
			if conn != nil {
				conn.Close()
			}
			close(ms.sendC)
			ms.sendC = nil
		}*/
}

// LocalAddr returns the local network address.
func (ms *MSock) LocalAddr() net.Addr {
	conn := ms.Conn()
	if conn != nil {
		return conn.LocalAddr()
	}
	return nil
}

// RemoteAddr returns the remote network address.
func (ms *MSock) RemoteAddr() net.Addr {
	conn := ms.Conn()
	if conn != nil {
		return conn.RemoteAddr()
	}
	return nil
}

func (ms *MSock) SetCallbacks(handler MSockHandleFunc, errNotify MSockErrorFunc) {
	if ms != nil {
		ms.handler = handler
		ms.errorNotify = errNotify
	}
}
func (ms *MSock) EnableGoCallback(en bool) {
	ms.disableGo = !en
}
func (ms *MSock) EnableCompress(en bool) {
	ms.compress = en
}
func (ms *MSock) IsCompress() bool {
	if ms == nil {
		return false
	}
	return ms.compress
}
func (ms *MSock) Status() int32 {
	return atomic.LoadInt32(&ms.status)
	/*v := ms.status.Load()
	if v == nil {
		return 0
	}
	return v.(int)*/
}
func (ms *MSock) setStatus(status int32) int32 {
	oldStatus := atomic.SwapInt32(&ms.status, status)
	mbase.Tracef("ms=%v,oldStatus=%d,newStatus=%d", ms, oldStatus, status)

	return oldStatus
	/*ms.mutex.Lock()
	defer ms.mutex.Unlock()

	oldStatus := ms.Status()

	mbase.Tracef("ms=%v,oldStatus=%d,newStatus=%d", ms, oldStatus, status)

	if oldStatus != MSockStatus_CLOSED {
		ms.status.Store(status)
	}

	return oldStatus*/
}
func (ms *MSock) SetReadBuffer(bytes int) error {
	conn := ms.Conn()
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		return tcpConn.SetReadBuffer(bytes)
	}
	return fmt.Errorf("Unsupported")
}

// SetWriteBuffer sets the size of the operating system's
// transmit buffer associated with the connection.
func (ms *MSock) SetWriteBuffer(bytes int) error {
	conn := ms.Conn()
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		return tcpConn.SetWriteBuffer(bytes)
	}
	return fmt.Errorf("Unsupported")
}

func (ms *MSock) Conn() net.Conn {
	if v := ms.conn.Load(); v != nil {
		return v.(net.Conn)
	}
	return nil
}
func (ms *MSock) setConn(c net.Conn) {
	ms.conn.Store(c)
}
func (ms *MSock) IsServer() bool {
	return ms.isServer
}

/*
func (ms *MSock) SendPacket(pk []byte) (n int, err error) {
	mbase.Tracef("ms=%v,len(pk)=%d", ms, len(pk))
	defer func() { mbase.Tracef("n=%d,err=%v", n, err) }()

	if ms == nil || ms.conn == nil {
		return 0, mbase.NewMError(mbase.MERR_MBASE_PARAM)
	}

	return mutils.WriteN(ms.conn, pk)
}
func (ms *MSock) sendPackets(mctx *MContext, datas ...[]byte) (err error) {
	mctx.Tracef("ms=%v,len(pk)=%d", ms, len(datas))
	defer func() { mctx.Tracef("err=%v", err) }()

	sd := &msockSendData{conn:ms.Conn(), mctx:mctx, datas:datas}

	if ms.Status() != MSockStatus_ESTABLISHED {
		return mbase.NewMError(mbase.MERR_MBASE_NETWORK)
	}

	ms.sendC <- sd

	return nil
}*/
func (ms *MSock) Send(mctx *MContext, body []byte) (err error) {
	mctx.Tracef("ms=%v", ms)
	defer func() { mctx.Tracef("ms=%v,err=%v", ms, err) }()

	mf := MFrame{}
	compress := ms.compress
	if compress {
		if len(body) > 0 {
			if b, berr := mutils.GZipCompress(body); berr == nil {
				mf.Flag |= MFRAME_FLAG_COMPRESS
				body = b
			}
		}
	}
	mf.Len = uint32(MFrameSize) + uint32(len(body))
	mf.Mark = MFrameMark
	if mctx.CanTrace() {
		mf.Flag |= MFRAME_FLAG_TRACELOG
	}

	data := make([]byte, 0, MFrameSize+len(body))
	data = MarshalMFrameBuff(data, &mf)

	if !compress {
		data = append(data, body...)
	}

	//return ms.sendPackets(mctx, data)
	if conn := ms.Conn(); conn == nil {
		return mbase.NewMError(mbase.MERR_MBASE_NETWORK)
	} else {
		_, err = mutils.WriteN(ms.Conn(), data)
		return err
	}
}
func (ms *MSock) SendPacket(mctx *MContext, pk []byte) (err error) {
	mctx.Tracef("ms=%v", ms)
	defer func() { mctx.Tracef("ms=%v,err=%v", ms, err) }()

	if conn := ms.Conn(); conn == nil {
		return mbase.NewMError(mbase.MERR_MBASE_NETWORK)
	} else {
		_, err = mutils.WriteN(ms.Conn(), pk)
		return err
	}
}

func (ms *MSock) SetCryptKey(cryptKey []byte) {
	ms.cryptKey = cryptKey
}
