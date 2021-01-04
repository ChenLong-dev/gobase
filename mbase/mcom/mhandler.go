/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:29:55
 * @LastEditTime: 2020-12-26 15:47:08
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcom

import (
	"github.com/ChenLong-dev/gobase/mlog"
	"sync"
)

//type MHandleFunc func() MoaError
type MHandleFunc func(mctx *MContext, msrv *MServer, data []byte)
type MHandler struct {
	handlers    sync.Map //map[uint32]	 interface{}
	defaultFunc MHandleFunc
}

func mhandleDefault(mctx *MContext, msrv *MServer, data []byte) {
	mlog.Debugf("no handler(%v)", mctx)
}
func NewMHandler(df MHandleFunc) *MHandler {
	if df == nil {
		df = mhandleDefault
	}
	mhan := &MHandler{defaultFunc: df}
	//mhan.requestMap = make(map[uint32]interface{})
	//mhan.pushMap   = make(map[uint32]interface{})

	return mhan
}

/*
func (mhan *MHandler) MServerHandler() MServerHandleFunc {
	return func(mctx *MContext, msrv *MServer, data []byte) {
		return mhan.Handle(msrv, mctx, data)
	}
}*/

//	handler define:  func Handle(mctx *MContext, msrv *MServer, body []byte)
func (mhan *MHandler) Register(opcode int32, handler MHandleFunc) {
	mhan.handlers.Store(opcode, handler)
}

func (mhan *MHandler) Handle(mctx *MContext, msrv *MServer, data []byte) {
	if h, ok := mhan.handlers.Load(mctx.OpCode()); ok {
		h.(MHandleFunc)(mctx, msrv, data)
	} else {
		mhan.defaultFunc(mctx, msrv, data)
	}
}
func (mhan *MHandler) MServerHandleFunc() MServerHandleFunc {
	return func(mctx *MContext, msrv *MServer, data []byte) {
		mhan.Handle(mctx, msrv, data)
	}
}

var defaultMHandler = NewMHandler(mhandleDefault)

func MServerHandler() MServerHandleFunc {
	return defaultMHandler.MServerHandleFunc()
}

//
func RegisterHandler(opcode int32, handler MHandleFunc) {
	defaultMHandler.Register(opcode, handler)
}

func RegisterHandlerDefault(df MHandleFunc) {
	defaultMHandler.defaultFunc = df
}

func Handle(mctx *MContext, msrv *MServer, data []byte) {
	defaultMHandler.Handle(mctx, msrv, data)
}
