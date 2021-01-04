/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:30:23
 * @LastEditTime: 2020-12-16 14:30:23
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcom

import (
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase"
	"github.com/ChenLong-dev/gobase/mbase/msys"
	"github.com/ChenLong-dev/gobase/mlog"
	"github.com/ChenLong-dev/gobase/sspb"
	"github.com/golang/protobuf/proto"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type MServerEntryFunc func(mctx *MContext, msrv *MServer)
type MServerReconnectFunc func(mctx *MContext, msrv *MServer)
type MServerHandleFunc func(mctx *MContext, msrv *MServer, data []byte)
type MServerErrorFunc func(mctx *MContext, msrv *MServer, err error)

type MServer struct {
	ms           *MSock
	reconnNotify MServerReconnectFunc
	handler      MServerHandleFunc
	errorNotify  MServerErrorFunc
	isServer     bool
	idAllocator  uint64
	rpcs         sync.Map
	Pridata      interface{} //	用户数据
}

type MServerListener struct {
	listener    *MSockListener
	handler     MServerHandleFunc
	entryNotify MServerEntryFunc
	errorNotify MServerErrorFunc
	disableGo   bool
}

func (msl *MServerListener) Close() {
	if msl != nil && msl.listener != nil {
		msl.listener.Close()
	}
}

func (msrv *MServer) String() string {
	if msrv == nil || msrv.ms == nil {
		return "MServer{nil}"
	}
	return fmt.Sprintf("MServer{ms:%v,status:%d,avail:%t,isServer:%t,idAllocator:%d}",
		msrv.ms, msrv.Status(), msrv.IsAvail(), msrv.isServer, msrv.idAllocator)
}

type mserverOptions struct {
	ignoreStatus bool
	disableGo    bool
	compress     bool
	entryNotify  MServerEntryFunc
	reconnNotify MServerReconnectFunc
	handler      MServerHandleFunc
	errorNotify  MServerErrorFunc
}
type MServerOption struct {
	f func(*mserverOptions)
}

func MServerOptIgnoreStatus() MServerOption {
	return MServerOption{f: func(do *mserverOptions) {
		do.ignoreStatus = true
	}}
}
func MServerOptDisableGo() MServerOption {
	return MServerOption{f: func(do *mserverOptions) {
		do.disableGo = true
	}}
}
func MServerOptCompress() MServerOption {
	return MServerOption{f: func(do *mserverOptions) {
		do.compress = true
	}}
}
func MServerOptEntryNotify(entryNotify MServerEntryFunc) MServerOption {
	return MServerOption{f: func(do *mserverOptions) {
		do.entryNotify = entryNotify
	}}
}
func MServerOptReconnNotify(reconnNotify MServerReconnectFunc) MServerOption {
	return MServerOption{f: func(do *mserverOptions) {
		do.reconnNotify = reconnNotify
	}}
}
func MServerOptHandle(handler MServerHandleFunc) MServerOption {
	return MServerOption{f: func(do *mserverOptions) {
		do.handler = handler
	}}
}
func MServerOptErrorNotify(errorNotify MServerErrorFunc) MServerOption {
	return MServerOption{f: func(do *mserverOptions) {
		do.errorNotify = errorNotify
	}}
}

func (msrv *MServer) IsServer() bool {
	return msrv.isServer
}

func (msrv *MServer) IsAvail() bool {
	return msrv != nil && msrv.Status() == MSockStatus_ESTABLISHED
}
func (msrv *MServer) Status() int32 {
	return msrv.ms.Status()
}

func (msrv *MServer) responseKeepalive(ctx *MContext, data []byte) error {
	/*req := &comproto.PB_MServerKeepaliveReq{}
	 err := proto.Unmarshal(data, req)
	 if err != nil {
		 return err
	 }
	 rsp := &comproto.PB_MServerKeepaliveRsp{Seq: req.Seq}
	 _, err = msrv.Reply(ctx, 0, rsp, MServerOptIgnoreStatus())*/
	return nil
}

func newMSockListenerEntry(msl *MServerListener) MSockEntryFunc {
	return func(mctx *MContext, ms *MSock) {
		mctx.Tracef("ms=%v", ms)

		msrv := &MServer{ms: ms, handler: msl.handler, errorNotify: msl.errorNotify, isServer: true, idAllocator: 2}
		ms.SetCallbacks(newMSockHandler(msrv), newMSockError(msrv))

		msrv.ms.EnableGoCallback(!msl.disableGo)

		if msl.entryNotify != nil {
			/*go */ msl.entryNotify(mctx, msrv)
		}
	}
}
func newMSockReconnect(msrv *MServer) MSockReconnectFunc {
	return func(mctx *MContext, ms *MSock) {
		//		go msrv.reconnect() //	这里必须要用goroutine，否则如果disableGo，此时处在同接受goroutine里，会导致调用joinReq陷入逻辑陷阱不返回
		if msrv.reconnNotify != nil {
			msrv.reconnNotify(mctx, msrv)
		}
	}
}

func newMSockHandler(msrv *MServer) MSockHandleFunc {
	return func(mctx *MContext, ms *MSock, data []byte) {
		mctx.Tracef("msrv=%v,ctx=%v,len(data)=%d", msrv, mctx, len(data))

		mhead, payload := UnmarshalMHead(data)
		if mhead == nil {
			//panic("UnmarshalMHead failed!")
			mctx.Tracef("UnmarshalMHead failed! ms=%v", ms)
			return
		}
		mctx.SetMHead(mhead)

		if IsResponse(mhead) {
			rpcResponse(mctx, msrv, payload)
		} else if msrv.handler != nil {
			msrv.handler(mctx, msrv, payload)
		}
	}
}

/*func newMSockListenerHandler(msl *MServerListener) MSockHandleFunc {
	return func(mctx *MContext, ms *MSock, data []byte) {
		mctx.Tracef("ms=%v,len(data)=%d", ms, len(data))

		msrv := &MServer{ms: ms, handler: msl.handler, errorNotify: msl.errorNotify, isServer: true, idAllocator: 2}
		ms.SetCallbacks(newMSockHandler(msrv), newMSockError(msrv))

		msrv.ms.EnableGoCallback(!msl.disableGo)

		if msl.entryNotify != nil {
			 msl.entryNotify(mctx, msrv)
		}
	}
}*/
func newMSockError(msrv *MServer) MSockErrorFunc {
	return func(mctx *MContext, ms *MSock, err error) {
		mctx.Tracef("msrv=%v,err=%v", msrv, err)

		//	全部rpc出错处理
		msrv.rpcAllClean(mctx)

		//	回调用户
		if msrv.errorNotify != nil {
			msrv.errorNotify(mctx, msrv, err)
		}
	}
}

func MServerListen(mctx *MContext, proto, addr string, entryNotify MServerEntryFunc, handler MServerHandleFunc, errorNotify MServerErrorFunc, opts ...MServerOption) (rmsl *MServerListener, rerr error) {
	mlog.Infof("proto=%s,addr=%s", proto, addr)
	defer func() { mlog.Infof("rmsl=%v,rerr=%v", rmsl, rerr) }()

	msl := &MServerListener{handler: handler, entryNotify: entryNotify, errorNotify: errorNotify}

	mopts := &mserverOptions{}
	for _, opt := range opts {
		opt.f(mopts)
	}
	if mopts.disableGo {
		msl.disableGo = mopts.disableGo
		mlog.Infof("set disableGo flag")
	}

	msockOpts := []MSockOption{MSockOptDisableGo()}
	//if msl.disableGo {
	//	msockOpts = append(msockOpts, )
	//}
	if mopts.compress {
		msockOpts = append(msockOpts, MSockOptCompress())
	}

	listener, err := MSockListen(mctx, proto, addr, newMSockListenerEntry(msl) /*newMSockListenerHandler(msl)*/, nil, nil, msockOpts...)
	if err != nil {
		mlog.Warnf("listen(%s,%s) error:%v", proto, addr, err)
		return nil, err
	}
	msl.listener = listener

	return msl, nil
}

func MServerDial(mctx *MContext, proto, addr string, reconnNotify MServerReconnectFunc, handler MServerHandleFunc, errorNotify MServerErrorFunc, opts ...MServerOption) (rmsrv *MServer, rerr error) {
	mctx.Tracef("proto=%s,addr=%s", proto, addr)
	defer func() { mctx.Tracef("rmsrv=%v,rerr=%v", rmsrv, rerr) }()

	msrv := &MServer{reconnNotify: reconnNotify, handler: handler, errorNotify: errorNotify, isServer: false, idAllocator: 1}

	mopts := &mserverOptions{}
	for _, opt := range opts {
		opt.f(mopts)
	}
	msockOpts := []MSockOption{}
	if mopts.disableGo {
		msockOpts = append(msockOpts, MSockOptDisableGo())
	}
	if mopts.compress {
		msockOpts = append(msockOpts, MSockOptCompress())
	}

	ms, err := MSockDial(mctx, proto, addr, newMSockReconnect(msrv), newMSockHandler(msrv), newMSockError(msrv), msockOpts...)
	if err != nil {
		return nil, err
	}
	msrv.ms = ms

	return msrv, nil
}

func (msrv *MServer) EnableGoCallback(en bool) {
	if msrv != nil && msrv.ms != nil {
		msrv.ms.EnableGoCallback(en)
	}
}
func (msrv *MServer) EnableCompress(en bool) {
	if msrv != nil && msrv.ms != nil {
		msrv.ms.EnableCompress(en)
	}
}

func (msrv *MServer) SetReadBuffer(bytes int) error {
	return msrv.ms.SetReadBuffer(bytes)
}
func (msrv *MServer) SetWriteBuffer(bytes int) error {
	return msrv.ms.SetWriteBuffer(bytes)
}

func (msrv *MServer) Close(mctx *MContext) {
	mctx.Tracef("msrv=%v", msrv)

	if msrv != nil && msrv.ms != nil {
		msrv.ms.Close(mctx)
	}

	msrv.rpcAllClean(mctx)
}

// LocalAddr returns the local network address.
func (msrv *MServer) LocalAddr() net.Addr {
	if msrv != nil && msrv.ms != nil {
		return msrv.ms.LocalAddr()
	}
	return nil
}

// RemoteAddr returns the remote network address.
func (msrv *MServer) RemoteAddr() net.Addr {
	if msrv != nil && msrv.ms != nil {
		return msrv.ms.RemoteAddr()
	}
	return nil
}

func (msrv *MServer) SetCallbacks(cbs ...MServerOption) {
	mbase.Tracef("msrv=%v", msrv)

	opts := &mserverOptions{}
	for _, opt := range cbs {
		opt.f(opts)
	}
	if opts.reconnNotify != nil {
		msrv.reconnNotify = opts.reconnNotify
	}
	if opts.handler != nil {
		msrv.handler = opts.handler
	}
	if opts.errorNotify != nil {
		msrv.errorNotify = opts.errorNotify
	}
}

func (msrv *MServer) SendPacket(mctx *MContext, pk []byte) (err error) {
	mctx.Tracef("msrv=%v,len(pk)=%d", msrv, len(pk))

	return msrv.ms.SendPacket(mctx, pk)
	/*
		if msrv.ms == nil {
			return 0, MError(MERR_MOA_NETWORK)
		}
		opts := &mserverOptions{}
		for _, opt := range options {
			opt.f(opts)
		}
		if !opts.ignoreStatus && msrv.Status() != MSockStatus_ESTABLISHED {
			return 0, MError(MERR_MOA_NETWORK)
		}

		return msrv.ms.SendPacket(pk)*/
}

func (msrv *MServer) Send(mctx *MContext, mhead *sspb.PBHead, payload []byte) error {
	mctx.Tracef("msrv=%v,mhead=%v,len(payload)=%v", msrv, mhead, len(payload))

	if msrv != nil && msrv.ms.IsCompress() {
		mctx.SetCompress(true)
	}

	pk := MarshalPacket(mctx, mhead, payload)

	return msrv.SendPacket(mctx, pk)
}
func (msrv *MServer) SendPb(mctx *MContext, mhead *sspb.PBHead, pb proto.Message) (e error) {
	mctx.Tracef("msrv=%v,mhead=%v,pb=%v", msrv, mhead, pb)

	if msrv != nil && msrv.ms.IsCompress() {
		mctx.SetCompress(true)
	}
	pk, err := MarshalPacketPb(mctx, mhead, pb)
	if err != nil {
		mctx.Tracef("err=%v", err)
		return err
	}

	return msrv.SendPacket(mctx, pk)
}

var mserverRpcFlag uint64 = (rand.New(rand.NewSource(time.Now().UnixNano())).Uint64() & 0xffffffff00000000)

func (msrv *MServer) allocRpcId() uint64 {
	id := atomic.AddUint64(&msrv.idAllocator, 2)
	//msrv.idAllocator += 2
	if id == 0 {
		id = atomic.AddUint64(&msrv.idAllocator, 2)
		//msrv.idAllocator += 2
	}

	//return msrv.idAllocator
	return (id & 0xffffffff) | mserverRpcFlag
}

type mserverRpcResponse struct {
	mctx *MContext
	data []byte
	err  error
}
type mserverRpcEnv struct {
	result    chan *mserverRpcResponse
	msrv      *MServer
	mt        *msys.MTimer
	rpcid     uint64
	reqOpcode int32
}

func rpcResponse(mctx *MContext, msrv *MServer, data []byte) (err error) {
	mctx.Tracef("msrv=%v,len(data)=%d", msrv, len(data))
	defer func() { mctx.Tracef("msrv=%v,err=%v", msrv, err) }()

	sh := mctx.MHead()

	rpcid := sh.Rpcid
	if rpcid == 0 || (rpcid&1) != (msrv.idAllocator&1) { //	没有设置rpcid或者不是本端分配的rpcid
		return mbase.NewMError(mbase.MERR_MBASE_NO_HANDLER)
	}

	v, ok := msrv.rpcs.Load(rpcid)
	if !ok {
		//	超时被删除了
		mctx.Tracef("rpc(%d) response timeout, so discard it!", rpcid)
		return mbase.NewMError(mbase.MERR_MBASE_TIMEOUT)
	}
	env := v.(*mserverRpcEnv)
	if env.reqOpcode+1 != sh.OpCode { //	opcode对不上
		return mbase.NewMError(mbase.MERR_MBASE_PARAM)
	}
	if !msrv.rpcRemove(env) {
		//	超时被删除了
		mctx.Tracef("rpc(%d) response timeout, so discard it!", rpcid)
		//return MError(MERR_MOA_TIMEOUT)
		return mbase.NewMError(mbase.MERR_MBASE_TIMEOUT)
	}

	errNo := mctx.ErrCode()
	var errRemote error = nil
	if errNo != 0 {
		errRemote = fmt.Errorf("(%d):%s", errNo, mctx.ErrMessage())
	}
	env.result <- &mserverRpcResponse{mctx: mctx, data: data, err: errRemote /*mbase.NewMError(errNo)*/}
	return nil
}
func rpcTimeoutFunc(mctx *MContext) msys.MTimerOverFunc {
	return func(pridata interface{}) {
		env := pridata.(*mserverRpcEnv)
		mctx.Tracef("env=%v", env)

		env.mt = nil
		if env.msrv.rpcRemove(env) {
			env.result <- &mserverRpcResponse{mctx: nil, data: nil, err: mbase.NewMError(mbase.MERR_MBASE_TIMEOUT)}
		}
	}
}

func (msrv *MServer) rpcRemove(env *mserverRpcEnv) bool {
	rpcid := atomic.SwapUint64(&env.rpcid, 0)
	if rpcid == 0 {
		return false
	}
	msrv.rpcs.Delete(rpcid)
	if env.mt != nil {
		env.mt.Stop()
		env.mt = nil
	}

	return true
}
func (msrv *MServer) Rpc(mctx *MContext, shead *sspb.PBHead, payload interface{}, timeout int64) (rsp []byte, err error) {
	mctx.Tracef("msrv=%v,shead=%v,timeout=%d", msrv, shead, timeout)
	defer func() { mctx.Tracef("len(rsp)=%d,err=%v", len(rsp), err) }()

	if payload == nil {
		payload = make([]byte, 0)
	}

	rpcid := msrv.allocRpcId()

	env := &mserverRpcEnv{msrv: msrv, rpcid: rpcid, reqOpcode: shead.GetOpCode()}
	env.result = make(chan *mserverRpcResponse, 1)
	msrv.rpcs.Store(rpcid, env)
	if timeout > 0 {
		env.mt = msys.StartMTimer(timeout, rpcTimeoutFunc(mctx), env)
	}

	shead.Rpcid = rpcid
	shead.Flags |= MHEAD_FLAG_REQUEST

	if body, ok := payload.([]byte); ok {
		err = msrv.Send(mctx, shead, body)
	} else {
		err = msrv.SendPb(mctx, shead, payload.(proto.Message))
	}

	if err != nil {
		msrv.rpcRemove(env)
		return nil, mbase.NewMError(mbase.MERR_MBASE_NETWORK)
	}

	response := <-env.result

	msrv.rpcRemove(env)

	return response.data, response.err
}

func (msrv *MServer) Reply(mctx *MContext, rerr error, payload interface{}) (err error) {
	mctx.Tracef("msrv=%v,rerr=%v", msrv, rerr)
	defer func() { mctx.Tracef("err=%v", err) }()

	if payload == nil {
		payload = make([]byte, 0)
	}

	reqMHead := mctx.MHead()
	rspMHead := MakeMHeadResponse(reqMHead)
	if rerr != nil {
		if me, ok := rerr.(*mbase.MError); ok {
			rspMHead.ErrCode = me.Errno()
			rspMHead.ErrMessage = me.Error()
		} else {
			rspMHead.ErrCode = -1
			rspMHead.ErrMessage = rerr.Error()
		}
	}

	if body, ok := payload.([]byte); ok {
		err = msrv.Send(mctx, rspMHead, body)
	} else {
		err = msrv.SendPb(mctx, rspMHead, payload.(proto.Message))
	}

	return err
}

func rpcClean(msrv *MServer) func(interface{}, interface{}) bool {
	return func(key, value interface{}) bool {
		env := value.(*mserverRpcEnv)
		if msrv.rpcRemove(env) {
			env.result <- &mserverRpcResponse{mctx: nil, data: nil, err: mbase.NewMError(mbase.MERR_MBASE_NETWORK)}
		}
		return true
	}
}
func (msrv *MServer) rpcAllClean(mctx *MContext) {
	mctx.Tracef("msrv=%v", msrv)

	msrv.rpcs.Range(rpcClean(msrv))
}
