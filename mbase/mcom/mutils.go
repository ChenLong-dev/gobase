/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:41:14
 * @LastEditTime: 2020-12-26 15:49:06
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcom

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
	"github.com/ChenLong-dev/gobase/mlog"
	. "github.com/ChenLong-dev/gobase/sspb"
	"github.com/golang/protobuf/proto"
)

const (
	CODE_MAX     int32 = 1000
	CODE_RPC_MAX int32 = 800
)

func GetOpcodeService(opcode int32) int32 {
	return opcode / CODE_MAX
}
func GetOpcodeNo(opcode int32) int32 {
	return opcode % CODE_MAX
}
func HasMFrameFlag(mframe *MFrame, flg uint16) bool {
	return (mframe.Flag & flg) == flg
}
func SetMFrameFlag(mframe *MFrame, flg uint16) {
	mframe.Flag |= flg
}
func ClrMFrameFlag(mframe *MFrame, flg uint16) {
	mframe.Flag &= (^flg)
}
func IsMHeadPush(mhead *PBHead) bool {
	return (mhead.GetFlags() & (MHEAD_FLAG_REQUEST | MHEAD_FLAG_RESPONSE)) == 0
}
func IsRequest(mhead *PBHead) bool {
	return (mhead.GetFlags() & MHEAD_FLAG_REQUEST) == MHEAD_FLAG_REQUEST
	/*opcode := sh.GetOpCode()
	code := GetOpcodeNo(opcode)
	return code <= CODE_RPC_MAX && ((code & 1) != 0) && !IsMHeadPush(sh)*/
}
func IsResponse(mhead *PBHead) bool {
	return (mhead.GetFlags() & MHEAD_FLAG_RESPONSE) == MHEAD_FLAG_RESPONSE
	/*opcode := sh.GetOpCode()
	code := GetOpcodeNo(opcode)
	return code <= CODE_RPC_MAX && ((code & 1) == 0) && !IsMHeadPush(sh)*/
}

func MarshalMFrameBuff(buff []byte, mframe *MFrame) []byte {
	buffSize := len(buff)
	if cap(buff)-buffSize < MFrameSize {
		data := make([]byte, 0, buffSize+MFrameSize)
		buff = append(data, buff...)
	}
	mb := buff[buffSize : buffSize+MFrameSize]
	binary.LittleEndian.PutUint32(mb, uint32(mframe.Len))
	binary.LittleEndian.PutUint16(mb[4:], uint16(mframe.Flag))
	binary.LittleEndian.PutUint16(mb[6:], uint16(mframe.Mark))

	return buff[:buffSize+MFrameSize]
}
func MarshalMFrame(mframe *MFrame) []byte {
	buff := make([]byte, MFrameSize)

	return MarshalMFrameBuff(buff, mframe)
}
func UnmarshalMFrame(data []byte) (mf *MFrame, res []byte) {
	if len(data) < MFrameSize {
		return nil, data
	}
	mframe := &MFrame{}

	mframe.Len = (binary.LittleEndian.Uint32(data))
	mframe.Flag = (binary.LittleEndian.Uint16(data[4:]))
	mframe.Mark = (binary.LittleEndian.Uint16(data[6:]))

	if mframe.Mark != MFrameMark {
		return nil, res
	}

	return mframe, data[MFrameSize:]
}
func MakeMHead(opcode int32) *PBHead {
	mhead := &PBHead{OpCode: opcode}
	/*code := GetOpcodeNo(opcode)
	if code > CODE_RPC_MAX {
		mhead.Flags |= MHEAD_FLAG_PUSH
	}*/
	return mhead
}
func MakeMHeadResponse(req *PBHead) (rsp *PBHead) {
	rsp = &PBHead{}
	rsp.OpCode = req.OpCode + 1
	rsp.Flags |= MHEAD_FLAG_RESPONSE
	rsp.Ext = req.Ext
	rsp.Rpcid = req.Rpcid

	return rsp
}

func MarshalMHeadBuff(buff []byte, mhead *PBHead) []byte {
	buffSize := len(buff)
	mheadSize := mhead.XXX_Size()
	if cap(buff)-buffSize < MHeadPrefSize+mheadSize {
		data := make([]byte, 0, buffSize+MHeadPrefSize+mheadSize)
		buff = append(data, buff...)
	}

	buff = buff[:buffSize+MHeadPrefSize]
	binary.LittleEndian.PutUint32(buff[buffSize:buffSize+MHeadPrefSize], uint32(mheadSize))
	//mutils.BytesWriteUint32(buff[buffSize:buffSize+MHeadPrefSize], uint32(mheadSize))

	bs, err := mhead.XXX_Marshal(buff[buffSize+MHeadPrefSize:], false)
	if err != nil || len(bs) != mheadSize {
		mlog.Warnf("Marshal error! expect marshal len=%d,actual marshal len=%d,marshal err=%v", mheadSize, len(bs), err)
		return nil
	}
	return buff[:buffSize+MHeadPrefSize+mheadSize]
}
func MarshalMHead(shead *PBHead) []byte {
	return MarshalMHeadBuff(nil, shead)
}
func UnmarshalMHead(data []byte) (mhead *PBHead, res []byte) {
	dataLen := len(data)
	if dataLen < MHeadPrefSize {
		return nil, data
	}
	slen := binary.LittleEndian.Uint32(data)
	if dataLen < MHeadPrefSize+int(slen) {
		return nil, data
	}

	mhead = &PBHead{}
	err := proto.Unmarshal(data[MHeadPrefSize:MHeadPrefSize+int(slen)], mhead)
	if err != nil {
		mlog.Debugf("err=%v", err)
		return nil, data
	}
	return mhead, data[MHeadPrefSize+int(slen):]
}

type newMarshaler interface {
	XXX_Size() int
	XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
}

func MarshalPacketPb(mctx *MContext, mhead *PBHead, pb proto.Message) ([]byte, error) {
	mheadSize := mhead.XXX_Size()
	payloadSize := 0
	var marsh newMarshaler
	if m, ok := pb.(newMarshaler); !ok && pb != nil {
		return nil, mbase.NewMError(mbase.MERR_MBASE_PARAM)
	} else if pb != nil {
		payloadSize = m.XXX_Size()
		marsh = m
	}

	pkSize := MFrameSize + MHeadPrefSize + mheadSize + payloadSize
	pk := make([]byte, 0, pkSize)

	//	MFrame
	mf := MFrame{}
	mf.Len = uint32(pkSize)
	mf.Mark = MFrameMark
	if mctx.CanTrace() {
		mf.Flag |= MFRAME_FLAG_TRACELOG
	}
	pk = MarshalMFrameBuff(pk, &mf)

	//	MHead
	pk = MarshalMHeadBuff(pk, mhead)

	//	payload
	if marsh != nil {
		mdata, merr := marsh.XXX_Marshal(pk[len(pk):], false)
		if merr != nil || len(mdata) != payloadSize {
			return nil, fmt.Errorf("payload marshal error! except len=%d,actual len=%d,merr=%v", payloadSize, len(mdata), merr)
		}
	}

	pk = pk[:pkSize]

	if mctx.Compress() {
		if b, berr := mutils.GZipCompress(pk[MFrameSize:]); berr == nil {
			cPk := make([]byte, 0, MFrameSize+len(b))
			mf.Flag |= MFRAME_FLAG_COMPRESS
			mf.Len = uint32(MFrameSize + len(b))
			cPk = MarshalMFrameBuff(cPk, &mf)
			cPk = append(cPk, b...)
			pk = cPk
		}
	}

	return pk, nil
}

func MarshalPacket(mctx *MContext, mhead *PBHead, payload []byte) []byte {
	/*	var compress bool = false
		if mctx.Compress() && len(payload) > 0 {
			if b, berr := mutils.GZipCompress(payload); berr == nil {
				payload = b
				compress = true
			}
		}*/

	mheadSize := mhead.XXX_Size()
	payloadSize := len(payload)
	pkSize := MFrameSize + MHeadPrefSize + mheadSize + payloadSize
	pk := make([]byte, 0, pkSize)

	//	MFrame
	mf := MFrame{}
	mf.Len = uint32(pkSize)
	mf.Mark = MFrameMark
	if mctx.CanTrace() {
		mf.Flag |= MFRAME_FLAG_TRACELOG
	}

	pk = MarshalMFrameBuff(pk, &mf)

	//	MHead
	pk = MarshalMHeadBuff(pk, mhead)

	//	payload
	pk = append(pk, payload...)

	if mctx.Compress() {
		if b, berr := mutils.GZipCompress(pk[MFrameSize:]); berr == nil {
			cPk := make([]byte, 0, MFrameSize+len(b))
			mf.Flag |= MFRAME_FLAG_COMPRESS
			mf.Len = uint32(MFrameSize + len(b))
			cPk = MarshalMFrameBuff(cPk, &mf)
			cPk = append(cPk, b...)
			pk = cPk
		}
	}

	return pk

	/*
		sheadData, _ := proto.Marshal(shead)
		slen := len(sheadData)
		plen := len(payload)
		pklen := MFrameSize + MHeadPrefSize + slen + plen

		data := make([]byte, MFrameSize + MHeadPrefSize, pklen)

		mframe := &MFrame{Len:uint32(pklen), Flag:0, Mark:MFrameMark}
		MarshalMFrameBuff(data, mframe)
		binary.LittleEndian.PutUint32(data[MFrameSize:], uint32(slen))
		data = append(data, sheadData...)
		data = append(data, payload...)

		return data*/
}

func PbNew(msg proto.Message) proto.Message {
	return proto.Clone(msg)
}

func PbPrint(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	b, _ := json.MarshalIndent(v, "", "\t")
	return string(b)
}

func PbPrintUnformatted(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	b, _ := json.Marshal(v)
	return string(b)
}

/*
func MFrameWrite( w io.Writer, mframe *MFrame) {

	//buf := &bytes.Buffer{}
	binary.Write(w, binary.LittleEndian, mframe.Len)
	binary.Write(w, binary.LittleEndian, mframe.Flag)
	binary.Write(w, binary.LittleEndian, mframe.Mark)

	//return buf.Bytes()
}
func MFrameRead(data []byte) *MFrame {
	mhead := &MFrame{}

	buf := bytes.NewReader(data)
	binary.Read(buf, binary.LittleEndian, &mhead.Len)
	binary.Read(buf, binary.LittleEndian, &mhead.Flag)
	binary.Read(buf, binary.LittleEndian, &mhead.Mark)

	return mhead
}*/
