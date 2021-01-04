/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:12:16
 * @LastEditTime: 2020-12-16 14:12:16
 * @LastEditors: Chen Long
 * @Reference:
 */

package ac

import (
	"encoding/binary"
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
)

//type TlvType int32

const BYOD_FIELD_MIN int32 = 0xFA00 /* 最小值 */
const (
	/* 通用字段 */
	BYOD_FIELD_OPT        int32 = BYOD_FIELD_MIN + iota /* 请求操作 */
	BYOD_FIELD_RET        int32 = BYOD_FIELD_MIN + iota /* 返回码 */
	BYOD_FIELD_RET_REASON int32 = BYOD_FIELD_MIN + iota /* 返回原因 */
	BYOD_FIELD_ACCOUNT    int32 = BYOD_FIELD_MIN + iota /* 账号 */
	BYOD_FIELD_PASSWORD   int32 = BYOD_FIELD_MIN + iota /* 密码【加密】 */
	BYOD_FIELD_MAC        int32 = BYOD_FIELD_MIN + iota /* MAC地址 */

	/* 业务相关 */
	BYOD_FIELD_FILE int32 = BYOD_FIELD_MIN + iota /* 文件数据 */

	/* 终端相关 */
	BYOD_FIELD_VIP   int32 = BYOD_FIELD_MIN + iota /* 终端唯一固定虚拟IP */
	BYOD_FIELD_GRP   int32 = BYOD_FIELD_MIN + iota /* 终端用户所属组 */
	BYOD_FIELD_REUSE int32 = BYOD_FIELD_MIN + iota /* 终端用户账号共用标志 */

	/* 引流接入相关 */
	BYOD_FIELD_DP_IP        int32 = BYOD_FIELD_MIN + iota /* DP引流接入IP */
	BYOD_FIELD_DP_PORT      int32 = BYOD_FIELD_MIN + iota /* DP引流接入端口 */
	BYOD_FIELD_BYOD_KEY     int32 = BYOD_FIELD_MIN + iota /* BYOD引流密匙 */
	BYOD_FIELD_VAC_AUTH_KEY int32 = BYOD_FIELD_MIN + iota /* VAC认证共享密钥 */

	/* 准入数据透明代理相关 */
	BYOD_FIELD_INGRESS_DATA        int32 = BYOD_FIELD_MIN + iota /* 准入原始数据 */
	BYOD_FIELD_INGRESS_LISTEN_IP   int32 = BYOD_FIELD_MIN + iota /* 平台代理准入数据服务的监听IP */
	BYOD_FIELD_INGRESS_LISTEN_PORT int32 = BYOD_FIELD_MIN + iota /* 平台代理准入数据服务的监听端口 */

	/* 信息上报通用字段 */
	BYOD_FIELD_VERSION        int32 = BYOD_FIELD_MIN + iota /* 版本号 */
	BYOD_FIELD_INSTALL_TIME   int32 = BYOD_FIELD_MIN + iota /* 安装时间 */
	BYOD_FIELD_OPERATE_SYSTEM int32 = BYOD_FIELD_MIN + iota /* 操作系统 */
	BYOD_FIELD_MEMORY         int32 = BYOD_FIELD_MIN + iota /* 内存信息 */
	BYOD_FIELD_CPU            int32 = BYOD_FIELD_MIN + iota /* CPU */

	/* 信息上报客户端相关字段 */
	BYOD_FIELD_PC_NAME int32 = BYOD_FIELD_MIN + iota /* PC用户名信息 */

	/* 2020.06.11以后新增字段，在此后添加，否则全平台必须整体升级，否则协议不兼容 */
	BYOD_FIELD_TIMESTAMP int32 = BYOD_FIELD_MIN + iota /* 时间戳 */
	BYOD_FIELD_MAX       int32 = BYOD_FIELD_MIN + iota /* 最大值 */
)

func MarshalTlv(t int32, v []byte) []byte {
	tData := mutils.BytesLittleEndianUint32(uint32(t))
	lData := mutils.BytesLittleEndianUint32(uint32(len(v)))
	data := make([]byte, 0, 8+len(v))
	data = append(data, tData...)
	data = append(data, lData...)
	return append(data, v...)
}
func UnmarshalTlv(data []byte) (t int32, v []byte, err error) {
	if len(data) < 8 {
		return 0, nil, fmt.Errorf("dataLen(%d) less", len(data))
	}

	ut := binary.LittleEndian.Uint32(data)
	ul := binary.LittleEndian.Uint32(data[4:])

	if int(ul) > len(data)-8 {
		return 0, nil, fmt.Errorf("len params(%d) greater than %d", ul, len(data)-8)
	}

	return int32(ut), data[8 : 8+ul], nil
}

func MarshalTlvUint8(t int32, v uint8) []byte {
	return MarshalTlv(t, []byte{v})
}
func MarshalTlvUint16(t int32, v uint16) []byte {
	return MarshalTlv(t, mutils.BytesLittleEndianUint16(v))
}
func MarshalTlvUint32(t int32, v uint32) []byte {
	return MarshalTlv(t, mutils.BytesLittleEndianUint32(v))
}
func MarshalTlvUint64(t int32, v uint64) []byte {
	return MarshalTlv(t, mutils.BytesLittleEndianUint64(v))
}
func MarshalTlvString(t int32, v string) []byte {
	return MarshalTlv(t, []byte(v))
}
func UnmarshalTlvUint8(data []byte) (t int32, v uint8, err error) {
	ut, uv, uerr := UnmarshalTlv(data)
	if uerr != nil {
		return 0, 0, uerr
	}
	if len(uv) != 1 {
		return ut, 0, fmt.Errorf("not uint8 data")
	}

	return ut, uv[0], nil
}
func UnmarshalTlvUint16(data []byte) (t int32, v uint16, err error) {
	ut, uv, uerr := UnmarshalTlv(data)
	if uerr != nil {
		return 0, 0, uerr
	}
	if len(uv) != 2 {
		return ut, 0, fmt.Errorf("not uint16 data")
	}

	return ut, binary.LittleEndian.Uint16(uv), nil
}
func UnmarshalTlvUint32(data []byte) (t int32, v uint32, err error) {
	ut, uv, uerr := UnmarshalTlv(data)
	if uerr != nil {
		return 0, 0, uerr
	}
	if len(uv) != 4 {
		return ut, 0, fmt.Errorf("not uint32 data")
	}

	return ut, binary.LittleEndian.Uint32(uv), nil
}
func UnmarshalTlvUint64(data []byte) (t int32, v uint64, err error) {
	ut, uv, uerr := UnmarshalTlv(data)
	if uerr != nil {
		return 0, 0, uerr
	}
	if len(uv) != 8 {
		return ut, 0, fmt.Errorf("not uint64 data")
	}

	return ut, binary.LittleEndian.Uint64(uv), nil
}
func UnmarshalTlvString(data []byte) (t int32, v string, err error) {
	ut, uv, uerr := UnmarshalTlv(data)
	if uerr != nil {
		return 0, "", uerr
	}

	return ut, string(uv), nil
}
func GetTlvType(data []byte) (t int32) {
	if len(data) < 4 {
		return 0
	}
	return int32(binary.LittleEndian.Uint32(data))
}

/*
type MarshalTlvFunc func(TlvType, v interface{}) []byte
type TVM struct {
	T 		TlvType
	V 		interface{}
	MFunc	MarshalTlvFunc
}
func (tvm *TVM) Marshal() []byte {
	if tvm.MFunc != nil {
		return tvm.MFunc(tvm.T, tvm.V)
	}
	switch tvm.V.(type) {
	case int32:
		return MarshalTlvUint32(tvm.T, uint32(tvm.V.(int32)))
	case uint32:
		return MarshalTlvUint32(tvm.T, tvm.V.(uint32))
	case int64:
		return MarshalTlvUint64(tvm.T, uint64(tvm.V.(int64)))
	case uint64:
		return MarshalTlvUint64(tvm.T, tvm.V.(uint64))
	case int8:
		return MarshalTlvUint8(tvm.T, uint8(tvm.V.(int8)))
	case uint8:
		return MarshalTlvUint8(tvm.T, uint8(tvm.V.(uint8)))
	case int16:
		return MarshalTlvUint16(tvm.T, uint16(tvm.V.(int16)))
	case uint16:
		return MarshalTlvUint16(tvm.T, tvm.V.(uint16))
	case string:
		return MarshalTlvString(tvm.T, tvm.V.(string))
	case []byte:
		return MarshalTlv(tvm.T, tvm.V.([]byte))
	default:
		panic(fmt.Sprintf("Unsupport type %v,%v", tvm.T, tvm.V))
	}
}
func MarshalPacket(tvms ...*TVM) []byte {
	vs := make([]byte, 0, 1024)
	for _, tvm := range tvms {
		vs = append(vs, tvm.Marshal()...)
	}
	return vs
}

type TVU struct {
	T 	TlvType
	V 	[]byte
}

func UnmarshalPacket(data []byte) (tvus []TVU, err error) {
	for len(data) > 0 {
		t, v, uerr := UnmarshalTlv(data)
		if uerr != nil {
			return tvus, uerr
		}
		tvus = append(tvus, TVU{T: t, V: v})

		data = data[8 + len(v) :]
	}
	return tvus, nil
}
*/

type TV struct {
	T int32
	V []byte
}
type TlvPacket struct {
	packet []byte
	tvs    []TV
}

func (tp *TlvPacket) Count() int {
	return len(tp.tvs)
}
func (tp *TlvPacket) Get(i int) (int32, []byte) {
	if i < 0 || i >= len(tp.tvs) {
		return 0, nil
	}
	return tp.tvs[i].T, tp.tvs[i].V
}
func (tp *TlvPacket) GetAsUint32(i int) (t int32, v uint32) {
	if t, vs := tp.Get(i); len(vs) < 4 {
		return t, 0
	} else {
		return t, binary.LittleEndian.Uint32(vs)
	}
}
func (tp *TlvPacket) Search(t int32) []byte {
	for _, tv := range tp.tvs {
		if tv.T == t {
			return tv.V
		}
	}
	return nil
}
func (tp *TlvPacket) Add(t int32, v []byte) {
	tp.tvs = append(tp.tvs, TV{T: t, V: v})
	tp.packet = nil
}
func (tp *TlvPacket) Set(t int32, v []byte) {
	for idx, tv := range tp.tvs {
		if tv.T == t {
			tp.tvs[idx].V = v
			tp.packet = nil
		}
	}
}
func (tp *TlvPacket) Del(t int32) {
	for i, tv := range tp.tvs {
		if tv.T == t {
			tp.tvs = append(tp.tvs[:i], tp.tvs[i+1:]...)
			tp.packet = nil
			break
		}
	}
}
func (tp *TlvPacket) Marshal() []byte {
	if tp.packet == nil {
		vs := make([]byte, 0, 1024)
		vs = append(vs, mutils.BytesLittleEndianUint32(0)...)
		for _, tv := range tp.tvs {
			vs = append(vs, MarshalTlv(tv.T, tv.V)...)
		}
		binary.LittleEndian.PutUint32(vs, uint32(len(vs)-4))
		tp.packet = vs
	}

	return tp.packet
}
func UnmarshalTlvPacket(packet []byte) (tp *TlvPacket, err error) {
	if len(packet) < 4 {
		return nil, fmt.Errorf("len(%d) invalid", len(packet))
	}
	if n := binary.LittleEndian.Uint32(packet); int(n) != len(packet)-4 {
		return nil, fmt.Errorf("packet len(%d) not match %d", n, len(packet)-4)
	}

	tp = new(TlvPacket)
	tp.packet = packet
	data := packet[4:]
	for len(data) > 0 {
		t, v, uerr := UnmarshalTlv(data)
		if uerr != nil {
			return nil, uerr
		}
		tp.tvs = append(tp.tvs, TV{T: t, V: v})

		data = data[8+len(v):]
	}
	return tp, nil
}
