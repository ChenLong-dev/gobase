/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:31:45
 * @LastEditTime: 2020-12-26 15:47:49
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcom

import (
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase"
	"github.com/ChenLong-dev/gobase/mlog"
	"github.com/ChenLong-dev/gobase/sspb"
	"github.com/golang/protobuf/proto"
	"net"
	"sync"
	"sync/atomic"
)

type MSessionEntryCb func(mctx *MContext, newMSession *MSession)
type MSessionDisconnectCb func(mctx *MContext, msession *MSession, err error)
type MSessionReconnectCb func(mctx *MContext, msession *MSession)
type MStreamEntryCb func(mctx *MContext, newMStream *MStream)

const MaxQuickStreamNum uint32 = 500
const MaxQuickStreamId uint64 = 1000

type MSessionCbs struct {
	SessionEntryCb      MSessionEntryCb
	SessionDisconnectCb MSessionDisconnectCb
	SessionReconnectCb  MSessionReconnectCb
	StreamEntryCb       MStreamEntryCb
}

type MSession struct {
	msrv      *MServer
	idAlloter uint64
	Cbs       MSessionCbs
	streams   sync.Map
	num       uint32
	Pridata   interface{}
}
type msessionOptions struct {
	src      string
	dst      string
	ext      string
	compress bool
}
type MSessionOption struct {
	f func(*msessionOptions)
}

func MSessionOptSrc(src string) MSessionOption {
	return MSessionOption{f: func(do *msessionOptions) {
		do.src = src
	}}
}
func MSessionOptDst(dst string) MSessionOption {
	return MSessionOption{f: func(do *msessionOptions) {
		do.dst = dst
	}}
}
func MSessionOptExt(ext string) MSessionOption {
	return MSessionOption{f: func(do *msessionOptions) {
		do.ext = ext
	}}
}
func MSessionOptCompress() MSessionOption {
	return MSessionOption{f: func(do *msessionOptions) {
		do.compress = true
	}}
}

//var slotMstream MStream
func (msession *MSession) allocId() uint64 {
	/*	if msession.idAlloter & 1 == 1 {	//	只对客户端发起的流做快速处理，服务端发起的不处理（基本上没有服务端发起流的应用场景）
			if atomic.LoadUint32(&msession.quickNum) < MaxQuickStreamNum {
				msession.quickMu.Lock()
				defer msession.quickMu.Unlock()
				for i := uint32(0); i < MaxQuickStreamNum; i++ {
					pos := msession.quickIdStart % MaxQuickStreamNum
					msession.quickIdStart ++
					if msession.quickStreams[pos] == nil {
						msession.quickStreams[pos] = &slotMstream	//	先占位，避免并发分配冲突
						return uint64(pos) * 2 + 1
					}
				}

				id := atomic.AddUint64(&msession.idAlloter, 2)
				if id < MaxQuickStreamId {
					id = atomic.AddUint64(&msession.idAlloter, MaxQuickStreamId)
				}
				return id
			}
		}
	*/
	return atomic.AddUint64(&msession.idAlloter, 2)
}
func (msession *MSession) insertStream(mstream *MStream) (err error) {
	/*	if mstream.id < MaxQuickStreamId && (mstream.id & 1 == 1) {
			atomic.AddUint32(&msession.quickNum, 1)
			msession.quickStreams[mstream.id/2] = mstream
			return nil
		}
	*/

	if _, loaded := msession.streams.LoadOrStore(mstream.id, mstream); loaded {
		return fmt.Errorf("exist stream[%d]!", mstream.id)
	}
	atomic.AddUint32(&msession.num, 1)
	return nil
}
func (msession *MSession) removeStream(id uint64) *MStream {
	mstream := msession.getStream(id)
	if mstream != nil {
		msession.streams.Delete(id)
	}
	return mstream
}
func (msession *MSession) getStream(id uint64) *MStream {
	//if id < MaxQuickStreamId && (id & 1 == 1) {
	//	return msession.quickStreams[id/2]
	//}
	if v, ok := msession.streams.Load(id); ok {
		return v.(*MStream)
	}
	return nil
}
func (msession *MSession) Num() int {
	return int(atomic.LoadUint32(&msession.num))
}
func (msession *MSession) OpenMStream(mctx *MContext, params ...MSessionOption) (mstream *MStream, err error) {
	opts := &msessionOptions{}
	for _, opt := range params {
		opt.f(opts)
	}

	mctx.Tracef("opts=%v", opts)
	defer func() { mctx.Tracef("mstream=%v,err=%v", mstream, err) }()

	id := msession.allocId()
	mstream = newMStream(msession, id, opts.src, opts.dst, mbase.GetMBaseConf().StreamBuff, mbase.GetMBaseConf().StreamBuff, opts.ext)
	err = msession.insertStream(mstream)
	if err != nil {
		return nil, err
	}

	pbOpen := &sspb.PBStreamOpen{Id: id, Src: opts.src, Dst: opts.dst, LocalWin: mbase.GetMBaseConf().StreamBuff, PeerWin: mbase.GetMBaseConf().StreamBuff, Ext: opts.ext}
	mhead := &sspb.PBHead{OpCode: int32(sspb.PBOpCode_OPC_MBASE_STREAM_OPEN), Rpcid: id} //
	err = msession.msrv.SendPb(mctx, mhead, pbOpen)
	if err != nil {
		msession.removeStream(id)
		return nil, err
	}

	return mstream, err
}
func (msession *MSession) onMStreamOpen(mctx *MContext, data []byte) {
	pbOpen := &sspb.PBStreamOpen{}
	err := proto.Unmarshal(data, pbOpen)
	if err != nil {
		mctx.Tracef("unmarshal error:%v", err)
		return
	}
	mctx.Tracef("pbOpen=%v", pbOpen)

	localWin := mbase.GetMBaseConf().StreamBuff

	mstream := newMStream(msession, pbOpen.Id, pbOpen.Src, pbOpen.Dst, localWin, pbOpen.LocalWin, pbOpen.Ext)
	err = msession.insertStream(mstream)
	if err != nil {
		mlog.Errorf("insert stream(%v) error:%v", mstream, err)
		oldS := msession.getStream(pbOpen.Id)
		msession.closeMStream(mctx, pbOpen.Id, err.Error())
		msession.insertStream(oldS)
	} else {
		msession.updateWindow(mctx, pbOpen.Id, localWin-pbOpen.PeerWin)
		if msession.Cbs.StreamEntryCb != nil {
			go msession.Cbs.StreamEntryCb(mctx, mstream)
		}
	}
}
func (msession *MSession) closeMStream(mctx *MContext, id uint64, reason string) (err error) {
	mctx.Tracef("id=%d,reason=%s", id, reason)
	defer func() { mctx.Tracef("id=%d,err=%v", id, err) }()

	mstream := msession.removeStream(id)
	if mstream == nil {
		return fmt.Errorf("no stream(%d) exist!", id)
	}

	pbClose := &sspb.PBStreamClose{Id: id, Reason: reason}
	mhead := &sspb.PBHead{OpCode: int32(sspb.PBOpCode_OPC_MBASE_STREAM_CLOSE), Rpcid: id} //
	err = msession.msrv.SendPb(mctx, mhead, pbClose)
	if err != nil {
		return fmt.Errorf("SendPb error:%v", err)
	}
	return err
}
func (msession *MSession) onMStreamClose(mctx *MContext, data []byte) {
	pbClose := &sspb.PBStreamClose{}
	err := proto.Unmarshal(data, pbClose)
	if err != nil {
		mctx.Tracef("unmarshal error:%v", err)
		return
	}
	mctx.Tracef("pbClose=%v", pbClose)

	mstream := msession.removeStream(pbClose.Id)
	if mstream == nil {
		mctx.Tracef("no stream(%d) exist", pbClose.Id)
		return
	}

	var closeErr error
	if pbClose.Reason != "" {
		closeErr = fmt.Errorf("%s", pbClose.Reason)
	}

	mstream.onClose(mctx, closeErr)
}
func (msession *MSession) updateWindow(mctx *MContext, id uint64, win int32) (err error) {
	mctx.Tracef("id=%d,win=%d", id, win)
	defer func() { mctx.Tracef("id=%d,win=%d,err=%v", id, win, err) }()

	pbUpdateWindow := &sspb.PBStreamUpdateWindow{Id: id, Win: win}
	mhead := &sspb.PBHead{OpCode: int32(sspb.PBOpCode_OPC_MBASE_STREAM_UPDATE_WINDOW), Rpcid: id} //
	err = msession.msrv.SendPb(mctx, mhead, pbUpdateWindow)
	if err != nil {
		return fmt.Errorf("SendPb error:%v", err)
	}
	return err
}
func (msession *MSession) onMStreamUpdateWindow(mctx *MContext, data []byte) {
	pbUpdateWindow := &sspb.PBStreamUpdateWindow{}
	err := proto.Unmarshal(data, pbUpdateWindow)
	if err != nil {
		mctx.Tracef("unmarshal error:%v", err)
		return
	}
	mctx.Tracef("pbUpdateWindow=%v", pbUpdateWindow)

	mstream := msession.getStream(pbUpdateWindow.Id)
	if mstream == nil {
		mctx.Tracef("no stream(%d) exist", pbUpdateWindow.Id)
		return
	}

	mstream.onUpdateWindow(mctx, pbUpdateWindow.Win)
}

func (msession *MSession) sendData(mctx *MContext, id uint64, data []byte) (err error) {
	mctx.Tracef("id=%d,len(data)=%d", id, len(data))
	defer func() { mctx.Tracef("id=%d,len(data)=%d,err=%v", id, len(data), err) }()

	pbData := &sspb.PBStreamData{Id: id, Data: data}
	mhead := &sspb.PBHead{OpCode: int32(sspb.PBOpCode_OPC_MBASE_STREAM_DATA), Rpcid: id} //
	err = msession.msrv.SendPb(mctx, mhead, pbData)
	if err != nil {
		return fmt.Errorf("SendPb error:%v", err)
	}
	return err
}
func (msession *MSession) onMStreamData(mctx *MContext, data []byte) {
	pbData := &sspb.PBStreamData{}
	err := proto.Unmarshal(data, pbData)
	if err != nil {
		mctx.Tracef("unmarshal error:%v", err)
		return
	}
	mctx.Tracef("id=%d,len(data)=%d", pbData.Id, len(pbData.Data))

	mstream := msession.getStream(pbData.Id)
	if mstream == nil {
		mctx.Tracef("no stream(%d) exist", pbData.Id)
		return
	}

	mstream.onData(mctx, pbData.Data)
}
func (msession *MSession) sendDataRaw(mctx *MContext, id uint64, data []byte) (err error) {
	mctx.Tracef("id=%d,len(data)=%d", id, len(data))
	defer func() { mctx.Tracef("id=%d,len(data)=%d,err=%v", id, len(data), err) }()

	mhead := &sspb.PBHead{OpCode: int32(sspb.PBOpCode_OPC_MBASE_STREAM_DATA_RAW), Rpcid: id} //
	err = msession.msrv.Send(mctx, mhead, data)
	if err != nil {
		return fmt.Errorf("Send error:%v", err)
	}
	return err
}
func (msession *MSession) onMStreamDataRaw(mctx *MContext, data []byte) {
	id := mctx.MHead().Rpcid
	mctx.Tracef("id=%d,len(data)=%d", id, len(data))

	mstream := msession.getStream(id)
	if mstream == nil {
		mctx.Tracef("no stream(%d) exist", id)
		return
	}

	mstream.onData(mctx, data)
}

// LocalAddr returns the local network address.
func (msession *MSession) LocalAddr() net.Addr {
	return msession.msrv.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (msession *MSession) RemoteAddr() net.Addr {
	return msession.msrv.RemoteAddr()
}

func (msession *MSession) IsServer() bool {
	return msession.msrv.IsServer()
}
func (msession *MSession) Close(mctx *MContext, err error) {
	mctx.Tracef("msession=%v,err=%v", msession, err)

	msession.msrv.Close(mctx)

	var mstreams []*MStream
	msession.streams.Range(func(key, value interface{}) bool {
		mstreams = append(mstreams, value.(*MStream))
		return true
	})

	for _, mstream := range mstreams {
		mstream.onClose(mctx, err)
	}
}

func mssEntryFunc(cbs MSessionCbs) MServerEntryFunc {
	return func(mctx *MContext, msrv *MServer) {
		mctx.Tracef("msrv=%v", msrv)

		msession := &MSession{msrv: msrv, idAlloter: 2, Cbs: cbs}
		msrv.Pridata = msession

		msrv.SetWriteBuffer(mbase.GetMBaseConf().SessionBuff)
		msrv.SetReadBuffer(mbase.GetMBaseConf().SessionBuff)

		if cbs.SessionEntryCb != nil {
			go cbs.SessionEntryCb(mctx, msession)
		}
	}
}
func mssReconnectFunc(mctx *MContext, msrv *MServer) {
	mctx.Tracef("msrv=%v", msrv)

	msession := msrv.Pridata.(*MSession)
	if msession != nil && msession.Cbs.SessionReconnectCb != nil {
		msession.Cbs.SessionReconnectCb(mctx, msession)
	}
}
func mssHandleFunc(mctx *MContext, msrv *MServer, data []byte) {
	mctx.Tracef("len(data)=%d", len(data))
	msession := msrv.Pridata.(*MSession)
	if msession == nil {
		mlog.Debugf("msession is nil")
		return
	}

	switch mctx.OpCode() {
	case int32(sspb.PBOpCode_OPC_MBASE_STREAM_OPEN):
		msession.onMStreamOpen(mctx, data)
	case int32(sspb.PBOpCode_OPC_MBASE_STREAM_CLOSE):
		msession.onMStreamClose(mctx, data)
	case int32(sspb.PBOpCode_OPC_MBASE_STREAM_DATA):
		msession.onMStreamData(mctx, data)
	case int32(sspb.PBOpCode_OPC_MBASE_STREAM_DATA_RAW):
		msession.onMStreamDataRaw(mctx, data)
	case int32(sspb.PBOpCode_OPC_MBASE_STREAM_UPDATE_WINDOW):
		msession.onMStreamUpdateWindow(mctx, data)
	default:
		mlog.Debugf("unknown opcode(%d)", mctx.OpCode())
	}
}
func mssErrorFunc(mctx *MContext, msrv *MServer, err error) {
	msession := msrv.Pridata.(*MSession)
	mctx.Tracef("msession=%v, err=%v", msession, err)

	if msession != nil {
		if msession.Cbs.SessionDisconnectCb != nil {
			msession.Cbs.SessionDisconnectCb(mctx, msession, err)
		}
		msession.Close(mctx, err)
	}
}

func MSessionDial(mctx *MContext, proto string, addr string, cbs MSessionCbs, params ...MSessionOption) (msession *MSession, err error) {
	mctx.Tracef("proto=%s,addr=%s", proto, addr)
	defer func() { mctx.Tracef("msession=%v,err=%v", msession, err) }()

	opts := &msessionOptions{}
	for _, opt := range params {
		opt.f(opts)
	}

	msrvOpts := []MServerOption{MServerOptDisableGo()}
	if opts.compress {
		msrvOpts = append(msrvOpts, MServerOptCompress())
	}

	msrv, merr := MServerDial(mctx, proto, addr, mssReconnectFunc, mssHandleFunc, mssErrorFunc, msrvOpts... /*MServerOptDisableGo()*/)
	if merr != nil {
		return nil, merr
	}

	msession = &MSession{msrv: msrv, idAlloter: 1, Cbs: cbs}
	msrv.Pridata = msession

	msrv.SetWriteBuffer(mbase.GetMBaseConf().SessionBuff)
	msrv.SetReadBuffer(mbase.GetMBaseConf().SessionBuff)

	return msession, nil

}
func MSessionListen(mctx *MContext, proto string, addr string, cbs MSessionCbs, params ...MSessionOption) (err error) {
	mctx.Tracef("proto=%s,addr=%s", proto, addr)
	defer func() { mctx.Tracef("err=%v", err) }()

	opts := &msessionOptions{}
	for _, opt := range params {
		opt.f(opts)
	}

	msrvOpts := []MServerOption{MServerOptDisableGo()}
	if opts.compress {
		msrvOpts = append(msrvOpts, MServerOptCompress())
	}

	_, merr := MServerListen(mctx, proto, addr, mssEntryFunc(cbs), mssHandleFunc, mssErrorFunc, msrvOpts... /*MServerOptDisableGo()*/)
	if merr != nil {
		return merr
	}

	return merr
}

func MSessionServer(msrv *MServer, cbs MSessionCbs) (msession *MSession) {

	idAlloter := uint64(1)
	if msrv.IsServer() {
		idAlloter = 2
	}

	msession = &MSession{msrv: msrv, idAlloter: idAlloter, Cbs: cbs}
	msrv.Pridata = msession

	msrv.SetWriteBuffer(mbase.GetMBaseConf().SessionBuff)
	msrv.SetReadBuffer(mbase.GetMBaseConf().SessionBuff)
	msrv.EnableGoCallback(false)
	msrv.SetCallbacks(MServerOptHandle(mssHandleFunc), MServerOptErrorNotify(mssErrorFunc))
	if !msrv.IsServer() {
		msrv.SetCallbacks(MServerOptReconnNotify(mssReconnectFunc))
	}

	return msession
}
