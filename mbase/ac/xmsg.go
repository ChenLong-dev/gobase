/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:13:24
 * @LastEditTime: 2020-12-16 14:13:24
 * @LastEditors: Chen Long
 * @Reference:
 */

package ac

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
)

const headSlot = 16

func xmsgHashKey(key []byte) uint32 {
	h := uint32(5318)
	for _, c := range key {
		h = h*33 + uint32(c)
	}

	return h
}
func xmsgIndex(key []byte) uint32 {
	return xmsgHashKey(key) & (headSlot - 1)
}

type xmsgElem struct {
	next  uint32
	dsize uint32
	ksize uint16
	flag  uint8
	align uint8
}

const xmsgElemSize = 12

func (xe *xmsgElem) marshal() []byte {
	bs := make([]byte, xmsgElemSize)

	binary.LittleEndian.PutUint32(bs, xe.next)
	binary.LittleEndian.PutUint32(bs[4:], xe.dsize)
	binary.LittleEndian.PutUint16(bs[8:], xe.ksize)
	bs[10] = xe.flag
	bs[11] = xe.align

	return bs
}
func unmarshalXmsgElem(bs []byte) (xe *xmsgElem) {
	if len(bs) < xmsgElemSize {
		return nil
	}
	xe = &xmsgElem{}
	xe.next = binary.LittleEndian.Uint32(bs)
	xe.dsize = binary.LittleEndian.Uint32(bs[4:])
	xe.ksize = binary.LittleEndian.Uint16(bs[8:])
	xe.flag = bs[10]
	xe.align = bs[11]

	return xe
}

type xmsgHead struct {
	magic uint32
	hash  uint32
	nelem uint32
	ndata uint32
}

const xmsgHeadSize = 16
const XMSG_MAGIC = uint32(0x47534d58)

func (xh *xmsgHead) marshal() []byte {
	bs := make([]byte, xmsgHeadSize)

	binary.LittleEndian.PutUint32(bs, xh.magic)
	binary.LittleEndian.PutUint32(bs[4:], xh.hash)
	binary.LittleEndian.PutUint32(bs[8:], xh.nelem)
	binary.LittleEndian.PutUint32(bs[12:], xh.ndata)

	return bs
}
func unmarshalXmsgHead(bs []byte) (xh *xmsgHead) {
	if len(bs) < xmsgHeadSize {
		return nil
	}

	xh = &xmsgHead{}
	xh.magic = binary.LittleEndian.Uint32(bs)
	xh.hash = binary.LittleEndian.Uint32(bs[4:])
	xh.nelem = binary.LittleEndian.Uint32(bs[8:])
	xh.ndata = binary.LittleEndian.Uint32(bs[12:])

	return xh
}

/*
const headSlot = 16
const elemFlagDeleted = 1
type XMsg struct {
	heads 	[headSlot]uint32
	Data 	[]byte
	Nelem	uint32
}
func (xmsg *XMsg) Add2(key []byte, val []byte) {
	realSize := len(key) + len(val) + 1 + xmsgElemSize
	space := mutils.Upbound(realSize, 4)

	elem := xmsgElem{}
	elem.flag  = 0
	elem.ksize = uint16(len(key))
	elem.dsize = uint32(len(val))
	elem.align = uint8(space - realSize)
	idx := xmsgHashKey(key)
	elem.next = xmsg.heads[idx]
	xmsg.heads[idx] = uint32(len(xmsg.Data))
	elemBs := elem.marshal()
	xmsg.Data = append(xmsg.Data, elemBs...)
	xmsg.Data = append(xmsg.Data, key...)
	xmsg.Data = append(xmsg.Data, val...)
	xmsg.Data = append(xmsg.Data, bytes.Repeat([]byte{0}, space-realSize+1)...)

	xmsg.Nelem++
}
func (xmsg *XMsg) Add(key string, val []byte) {
	xmsg.Add2([]byte(key), val)
}
func (xmsg *XMsg) Del2(key []byte) int {
	idx := xmsgHashKey(key)
	for next := xmsg.heads[idx]; next != 0; {

	}
}
func (xmsg *XMsg) Del(key string) int {
	return xmsg.Del2([]byte(key))
}*/

type XMsgEntry struct {
	key string
	val []byte
}
type XMsg struct {
	entrys []*XMsgEntry
	m      map[string][]*XMsgEntry
}

func (xmsg *XMsg) String() string {
	str := fmt.Sprintf("{nelem:%d,elems[", len(xmsg.entrys))
	for _, e := range xmsg.entrys {
		str += fmt.Sprintf("%s:%d,", e.key, len(e.val))
	}
	str += "]}"
	return str
}
func NewXMsg() *XMsg {
	xmsg := new(XMsg)
	xmsg.entrys = make([]*XMsgEntry, 0, 10)
	xmsg.m = make(map[string][]*XMsgEntry)

	return xmsg
}
func NewXMsgWithCreator(creator string) *XMsg {
	xmsg := NewXMsg()
	if creator == "" {
		creator = "(null)"
	}
	xmsg.Add("creator", append([]byte(creator), 0))
	return xmsg
}
func (xmsg *XMsg) Add(key string, val []byte) {
	entry := new(XMsgEntry)
	entry.key = key
	entry.val = val

	xmsg.entrys = append(xmsg.entrys, entry)
	es := xmsg.m[key]
	es = append(es, entry)
	xmsg.m[key] = es
}
func (xmsg *XMsg) Del(key string) int {
	if len(xmsg.entrys) == 0 {
		return 0
	}

	delNum := 0
	for i := 0; i < len(xmsg.entrys); i++ {
		if xmsg.entrys[i].key == key {
			xmsg.entrys = append(xmsg.entrys[:i], xmsg.entrys[i+1:]...)
			delNum++
		}
	}
	delete(xmsg.m, key)

	return delNum
}
func (xmsg *XMsg) Get(key string) []byte {
	if xes, ok := xmsg.m[key]; !ok || len(xes) == 0 {
		return nil
	} else {
		return xes[0].val
	}
}
func (xmsg *XMsg) Gets(key string) []*XMsgEntry {
	if xes, ok := xmsg.m[key]; ok {
		return xes
	} else {
		return nil
	}
}
func (xmsg *XMsg) NElem() int {
	return len(xmsg.entrys) - 1 //	去掉creator
}

type XmsgIterator struct {
	xmsg *XMsg
	i    int
}

func (xi *XmsgIterator) Done() bool {
	return xi.i < 0 || xi.i >= len(xi.xmsg.entrys)
}
func (xi *XmsgIterator) Get() *XMsgEntry {
	if xi.Done() {
		return nil
	}
	return xi.xmsg.entrys[xi.i]
}
func (xi *XmsgIterator) Next() {
	if !xi.Done() {
		xi.i++
	}
}
func (xi *XmsgIterator) Prev() {
	if !xi.Done() {
		xi.i--
	}
}
func (xmsg *XMsg) First() *XmsgIterator {
	if len(xmsg.entrys) == 0 {
		return nil
	}

	return &XmsgIterator{xmsg: xmsg, i: 0}
}
func (xmsg *XMsg) Last() *XmsgIterator {
	if len(xmsg.entrys) == 0 {
		return nil
	}

	return &XmsgIterator{xmsg: xmsg, i: len(xmsg.entrys) - 1}
}

func (xmsg *XMsg) Marshal() []byte {

	var heads [headSlot]uint32
	body := make([]byte, 0, 3084)
	for _, e := range xmsg.entrys {
		realSize := len(e.key) + len(e.val) + 1 + xmsgElemSize
		space := mutils.Upbound(realSize, 4)
		bKey := []byte(e.key)

		elem := xmsgElem{}
		elem.flag = 0
		elem.ksize = uint16(len(e.key))
		elem.dsize = uint32(len(e.val))
		elem.align = uint8(space - realSize)
		idx := xmsgIndex(bKey)
		elem.next = heads[idx]
		heads[idx] = uint32(len(body))
		elemBs := elem.marshal()
		body = append(body, elemBs...)
		body = append(body, bKey...)
		body = append(body, e.val...)
		body = append(body, bytes.Repeat([]byte{0}, space-realSize+1)...)
	}

	xh := xmsgHead{}
	xh.magic = XMSG_MAGIC
	xh.hash = xmsgHashKey(body)
	xh.ndata = uint32(len(body))
	xh.nelem = uint32(len(xmsg.entrys))

	return append(xh.marshal(), body...)
}

func UnmarshalXMsg(data []byte) (xmsg *XMsg, err error) {
	xh := unmarshalXmsgHead(data)
	if xh == nil {
		return nil, fmt.Errorf("unmarshalXmsgHead failed")
	}
	data = data[xmsgHeadSize:]
	if xh.hash != xmsgHashKey(data) {
		return nil, fmt.Errorf("hash verify failed")
	}

	xmsg = NewXMsg()
	pb := data
	for len(pb) > 0 {
		xe := unmarshalXmsgElem(pb)
		if xe == nil {
			return nil, fmt.Errorf("unmarshalXmsgElem failed")
		}
		pb = pb[xmsgElemSize:]
		if int(xe.ksize)+int(xe.dsize) >= len(pb) {
			return nil, fmt.Errorf("xmsgElem len error")
		}

		bKey := pb[:xe.ksize]
		val := pb[xe.ksize : uint32(xe.ksize)+xe.dsize]
		xmsg.Add(string(bKey), val)

		pb = pb[uint32(xe.ksize)+xe.dsize+1+uint32(xe.align):]
	}

	return xmsg, nil
}
