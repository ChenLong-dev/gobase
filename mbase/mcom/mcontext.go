/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:24:58
 * @LastEditTime: 2020-12-16 14:24:58
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcom

import (
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
	"github.com/ChenLong-dev/gobase/mlog"
	"github.com/ChenLong-dev/gobase/sspb"
	"time"
)

type MContext struct {
	mhead    *sspb.PBHead //	路由头
	trace    bool
	compress bool
}

func (mctx *MContext) Deadline() (deadline time.Time, ok bool) {
	return
}
func (mctx *MContext) Done() <-chan struct{} {
	return nil
}
func (mctx *MContext) Err() error {
	return nil
}
func (mctx *MContext) Value(key interface{}) interface{} {
	if skey, ok := key.(string); ok {
		switch skey {
		case "opCode":
			return mctx.OpCode()
		case "errCode":
			return mctx.ErrCode()
		//case "flags","flag":
		//	return mctx.Flags()
		case "errMessage":
			return mctx.ErrMessage()
		}
	}
	return nil
}

func (mctx *MContext) SetTrace(trace bool) {
	if mctx != nil {
		mctx.trace = trace
	}
}
func (mctx *MContext) CanTrace() bool {
	if mctx == nil {
		return false
	}
	return mctx.trace
}
func (mctx *MContext) SetCompress(compress bool) {
	if mctx != nil {
		mctx.compress = compress
	}
}
func (mctx *MContext) Compress() bool {
	if mctx == nil {
		return false
	}
	return mctx.compress
}
func (mctx *MContext) MHead() *sspb.PBHead {
	if mctx == nil {
		return nil
	}
	return mctx.mhead
}
func (mctx *MContext) SetMHead(mhead *sspb.PBHead) {
	if mctx != nil {
		mctx.mhead = mhead
	}
}
func (mctx *MContext) OpCode() int32 {
	mhead := mctx.MHead()
	if mhead != nil {
		return mhead.GetOpCode()
	}
	return 0
}

/*func (mctx *MContext) Flags() uint64 {
	mhead := mctx.MHead()
	if mhead != nil {
		return mhead.GetFlags()
	}
	return 0
}*/
func (mctx *MContext) ErrCode() int32 {
	mhead := mctx.MHead()
	if mhead != nil {
		return mhead.GetErrCode()
	}
	return 0
}
func (mctx *MContext) ErrMessage() string {
	mhead := mctx.MHead()
	if mhead != nil {
		return mhead.GetErrMessage()
	}
	return ""
}
func (mctx *MContext) Tracef(format string, v ...interface{}) {
	if mctx != nil && mctx.trace {
		//mlog.ForceTracef(format, v...)
		mlog.Output(mlog.TRACE, 3, fmt.Sprint("mctx:", mctx), format, v...)
	} else {
		if mlog.GetLevel() <= mlog.TRACE {
			mlog.Output(mlog.TRACE, 3, fmt.Sprint("mctx:", mctx), format, v...)
		}
	}
}

func (mctx *MContext) String() string {
	if mctx == nil {
		return ""
	}

	return fmt.Sprintf("{trace:%v,mhead:%s}", mctx.trace, mutils.JsonPrint(mctx.mhead))
}

func NewMContextWithMHead(mhead *sspb.PBHead, trace bool) *MContext {
	return &MContext{mhead: mhead, trace: trace}
}
func NewMContextWithTrace(trace bool) *MContext {
	return &MContext{trace: trace}
}
func NewMContext() *MContext {
	return &MContext{}
}
