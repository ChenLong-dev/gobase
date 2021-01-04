/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:06:50
 * @LastEditTime: 2020-12-16 14:07:52
 * @LastEditors: Chen Long
 * @Reference:
 */

package ac

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
	"github.com/ChenLong-dev/gobase/mlog"
	"net"
	"os"
	"strings"
)

type Cfgc struct {
	ProgId    string
	Zone      string
	Ns        string
	conn      net.Conn
	autoField bool
}

func (cfgc *Cfgc) String() string {
	if cfgc == nil {
		return "{}"
	}
	return fmt.Sprintf("{ProgId:%s,Zone:%s,Ns:%s,Conn:%v}", cfgc.ProgId, cfgc.Zone, cfgc.Ns, cfgc.conn)
}

const UNIX_PATH_MAX = 108

func (cfgc *Cfgc) connect() (err error) {
	mlog.Tracef("%v connect...", cfgc)
	defer func() { mlog.Tracef("err=%v", err) }()

	cfgAddr := os.Getenv("EPS_CFG_CENTER_SOCK")
	if cfgAddr == "" || cfgAddr[0] != '/' {
		cfgAddr = fmt.Sprintf("@/var/cfgcenter.sock.%d", int(EpsGetLocation()))
	}

	/*rAddr, rerr := net.ResolveUnixAddr("unix", cfgAddr)
	if rerr != nil {
		return rerr
	}*/
	if cfgAddr[0] == '@' {
		cfgAddr += string(bytes.Repeat([]byte{0}, UNIX_PATH_MAX-len(cfgAddr)))
	}

	conn, cerr := net.Dial("unix", cfgAddr)
	if cerr != nil {
		return fmt.Errorf("dial unix addr(%s) error:%v", cfgAddr, cerr)
	}
	if cfgc.conn != nil {
		cfgc.conn.Close()
	}

	cfgc.conn = conn
	return nil
}
func (cfgc *Cfgc) GetError(rsp *XMsg) string {
	err := rsp.Get("@error")
	if err == nil {
		return ""
	}
	return string(err)
}
func (cfgc *Cfgc) GetErrno(rsp *XMsg) int32 {
	errno := rsp.Get("@errno")
	if errno == nil || len(errno) < 4 {
		return -1
	}
	return int32(binary.LittleEndian.Uint32(errno))
}
func (cfgc *Cfgc) send(req *XMsg) (err error) {
	if cfgc.conn == nil {
		if err = cfgc.Open(); err != nil {
			return fmt.Errorf("open error:%v", err)
		}
	}

	reqData := req.Marshal()
	reqDataSize := uint32(len(reqData))
	if err = mutils.WriteUint32(cfgc.conn, reqDataSize); err != nil {
		return fmt.Errorf("write head error:%v", err)
	}
	if _, err = mutils.WriteN(cfgc.conn, reqData); err != nil {
		return fmt.Errorf("write body error:%v", err)
	}
	return nil
}
func (cfgc *Cfgc) recv() (rsp *XMsg, err error) {
	var rspDataSize uint32
	if rspDataSize, err = mutils.ReadUint32(cfgc.conn); err != nil {
		return nil, fmt.Errorf("read head error:%v", err)
	}
	if rspDataSize > 100*1024*1024 {
		return nil, fmt.Errorf("recv body size(%d) too large", rspDataSize)
	}
	rspData := make([]byte, rspDataSize)
	if _, err = mutils.ReadN(cfgc.conn, rspData); err != nil {
		return nil, fmt.Errorf("read body error:%v", err)
	}
	return UnmarshalXMsg(rspData)
}
func (cfgc *Cfgc) Rpc(req *XMsg) (rsp *XMsg, err error) {
	mlog.Tracef("req=%v", req)
	defer func() {
		if err == nil {
			mlog.Tracef("rsp=%v,err=%v", rsp, err)
		} else {
			mlog.Warnf("err=%v", err)
			cfgc.Close()
		}
	}()

	if err = cfgc.send(req); err != nil {
		return nil, err
	}

	return cfgc.recv()
}
func (cfgc *Cfgc) Open() (err error) {
	if cfgc.ProgId == "" || cfgc.Zone == "" || cfgc.Ns == "" {
		return fmt.Errorf("%v has empty param", cfgc)
	}
	if cfgc.conn == nil {
		if err = cfgc.connect(); err != nil {
			return err
		}
	}

	if os.Getenv("CFGC_AUTO_FIELD") == "1" {
		cfgc.autoField = true
	} else {
		cfgc.autoField = false
	}

	openReq := NewXMsgWithCreator(cfgc.ProgId)
	openReq.Add("cmd", []byte("open"))
	openReq.Add("zone", []byte(cfgc.Zone))
	openReq.Add("ns", []byte(cfgc.Ns))

	openRsp, openErr := cfgc.Rpc(openReq)
	if openErr != nil {
		return fmt.Errorf("open error:%v", openErr)
	}
	if errinfo := cfgc.GetError(openRsp); errinfo != "" {
		return fmt.Errorf("open error(%d):%s", cfgc.GetErrno(openRsp), errinfo)
	}
	return nil
}
func (cfgc *Cfgc) Close() {
	conn := cfgc.conn
	cfgc.conn = nil
	if conn != nil {
		conn.Close()
	}
}
func (cfgc *Cfgc) call(req *XMsg) (rsp *XMsg, err error) {
	rsp, err = cfgc.Rpc(req)
	if err != nil {
		return rsp, err
	}
	if einfo := cfgc.GetError(rsp); einfo != "" {
		mlog.Tracef("response error(%d):%s", cfgc.GetErrno(rsp), einfo)
		return rsp, fmt.Errorf("response error(%d):%s", cfgc.GetErrno(rsp), einfo)
	}
	return rsp, nil
}
func (cfgc *Cfgc) ListNs() (*XMsg, error) {
	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("listns"))
	req.Add("zone", []byte(cfgc.Zone))

	return cfgc.call(req)
}
func (cfgc *Cfgc) ListConnection() (*XMsg, error) {
	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("listns"))
	req.Add("zone", []byte(cfgc.Zone))

	return cfgc.call(req)
}
func (cfgc *Cfgc) QueryMeta(req *XMsg) (*XMsg, error) {
	req.Add("cmd", []byte("meta"))
	return cfgc.Rpc(req)
}
func (cfgc *Cfgc) QueryVoidMeta(req *XMsg) (*XMsg, error) {
	req.Add("cmd", []byte("void_meta"))
	return cfgc.Rpc(req)
}
func (cfgc *Cfgc) GetObject(path string, oc uint16) (rsp *XMsg, err error) {
	mlog.Tracef("path=%s,oc=%d", path, oc)
	defer func() { mlog.Tracef("rsp=%v,err=%v", rsp, err) }()

	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("get"))
	req.Add("oc", mutils.BytesLittleEndianUint16(oc))
	req.Add("path", []byte(path))

	return cfgc.call(req)
}
func (cfgc *Cfgc) GetObjectCount(path string, oc uint16, recursive bool) int {
	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("subcnt"))
	req.Add("oc", mutils.BytesLittleEndianUint16(oc))
	req.Add("path", []byte(path))
	if recursive {
		req.Add("recursive", []byte("1"))
	}

	rsp, err := cfgc.call(req)
	if err != nil {
		return -1
	}
	if byteCount := rsp.Get("count"); len(byteCount) != 4 {
		return -1
	} else {
		return int(binary.LittleEndian.Uint32(byteCount))
	}
}
func (cfgc *Cfgc) AddObject(path string, oc uint16, attr *XMsg) (err error) {
	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("add"))
	req.Add("oc", mutils.BytesLittleEndianUint16(oc))
	req.Add("path", []byte(path))
	if attr != nil {
		if cfgc.autoField {
			attr.Del("__auto_add__")
			attr.Add("__auto_add__", []byte("1"))
		}
		req.Add("attr", attr.Marshal())
	}

	_, err = cfgc.call(req)
	return err
}
func (cfgc *Cfgc) ModObject(path string, oc uint16, attr *XMsg) (err error) {
	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("modify"))
	req.Add("oc", mutils.BytesLittleEndianUint16(oc))
	req.Add("path", []byte(path))
	if attr != nil {
		req.Add("attr", attr.Marshal())
	}

	_, err = cfgc.call(req)
	return err
}
func (cfgc *Cfgc) DelObject(path string, oc uint16) (err error) {
	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("del"))
	req.Add("oc", mutils.BytesLittleEndianUint16(oc))
	req.Add("path", []byte(path))

	_, err = cfgc.call(req)
	return err
}
func (cfgc *Cfgc) MovObject(path string, oc uint16, newPath string) (err error) {
	attr := NewXMsgWithCreator(cfgc.ProgId)
	attr.Add("new_path", []byte(newPath))

	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("move"))
	req.Add("oc", mutils.BytesLittleEndianUint16(oc))
	req.Add("path", []byte(path))
	req.Add("attr", attr.Marshal())

	_, err = cfgc.call(req)
	return err
}
func (cfgc *Cfgc) DelAttr(path string, oc uint16, attrs []string) (err error) {
	attr := NewXMsgWithCreator(cfgc.ProgId)
	if len(attrs) == 0 {
		attr.Add("allattr", []byte("1"))
	} else {
		for _, at := range attrs {
			attr.Add("attr", []byte(at))
		}
	}

	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("dattr"))
	req.Add("oc", mutils.BytesLittleEndianUint16(oc))
	req.Add("path", []byte(path))
	req.Add("attr", attr.Marshal())

	_, err = cfgc.call(req)
	return err
}

type CfgGroup struct {
	GrpName  string
	Names    []string
	Ocs      []uint16
	Attrs    []*XMsg
	maxNames int
	curr     int
}

func NewCfgGroup(gpath string, maxNames int) *CfgGroup {
	cgroup := new(CfgGroup)
	cgroup.GrpName = gpath
	cgroup.maxNames = maxNames

	return cgroup
}
func (cgroup *CfgGroup) Add(name string, oc uint16) {
	if len(cgroup.Names) >= cgroup.maxNames {
		return
	}
	cgroup.Ocs = append(cgroup.Ocs, oc)
	cgroup.Names = append(cgroup.Names, name)
	cgroup.Attrs = append(cgroup.Attrs, nil)
}
func (cgroup *CfgGroup) GetPath(idx int) string {
	if idx >= len(cgroup.Names) || idx < 0 {
		return ""
	}
	if cgroup.GrpName != "" {
		if cgroup.GrpName == "/" {
			return cgroup.GrpName + cgroup.Names[idx]
		} else {
			return cgroup.GrpName + "/" + cgroup.Names[idx]
		}
	} else {
		return cgroup.Names[idx]
	}
}

func (cfgc *Cfgc) commitSubs(xmsg *XMsg, gpath string) (cgroup *CfgGroup, err error) {
	mlog.Tracef("xmsg=%v,gpath=%s", xmsg, gpath)
	defer func() { mlog.Tracef("cgroup=%v,err=%v", cgroup, err) }()

	rsp, rerr := cfgc.call(xmsg)
	if rerr != nil {
		return nil, rerr
	}

	cgroup = NewCfgGroup(gpath, rsp.NElem())
	if subs := rsp.Gets("sub"); subs != nil {
		for _, sub := range subs {
			oc := binary.LittleEndian.Uint16(sub.val)
			name := string(sub.val[2:])
			cgroup.Add(name, oc)
		}
	}
	return cgroup, nil
}
func (cfgc *Cfgc) OpenGroupWithOc(path string, oc uint16) (cgroup *CfgGroup, err error) {
	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("subgrp"))
	req.Add("path", []byte(path))
	req.Add("oc", mutils.BytesLittleEndianUint16(oc))

	return cfgc.commitSubs(req, path)
}
func (cfgc *Cfgc) ListObject(path string, oc uint16, recursive, includeBase uint32) (cgroup *CfgGroup, err error) {
	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("listobj"))
	req.Add("path", []byte(path))
	req.Add("oc", mutils.BytesLittleEndianUint16(oc))
	req.Add("recursive", mutils.BytesLittleEndianUint32(recursive))
	req.Add("include_base", mutils.BytesLittleEndianUint32(includeBase))

	return cfgc.commitSubs(req, path)
}
func (cfgc *Cfgc) OpenGroup(path string) (cgroup *CfgGroup, err error) {
	return cfgc.OpenGroupWithOc(path, 0)
}
func (cfgc *Cfgc) QueryAttr(oc uint16, attr string, val []byte) (cgroup *CfgGroup, err error) {
	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("query"))
	req.Add("attr", []byte(attr))
	req.Add("value", val)
	req.Add("oc", mutils.BytesLittleEndianUint16(oc))

	return cfgc.commitSubs(req, "")
}
func (cfgc *Cfgc) QueryAttrHas(oc uint16, attr string, val []byte) (cgroup *CfgGroup, err error) {
	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("query_has"))
	req.Add("attr", []byte(attr))
	req.Add("value", val)
	req.Add("oc", mutils.BytesLittleEndianUint16(oc))

	return cfgc.commitSubs(req, "")
}
func ExpandPath(path string, oc uint16) (cgroup *CfgGroup, err error) {
	if len(path) == 0 {
		return nil, nil
	}
	count := strings.Count(path, "/") + 1
	cgroup = NewCfgGroup("", count)
	cgroup.Add(path, oc)

	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i > 0 && path[i-1] != '/' {
				cgroup.Add(path[:i], uint16('O'))
			}
		}
	}
	cgroup.Add("/", uint16('O'))
	return cgroup, nil
}

const BAT_READ_OBJ_COUNT = 500

func (cfgc *Cfgc) GetSubObjects(sub *CfgGroup, idx int) (err error) {
	mlog.Tracef("sub=%v,idx=%d", sub, idx)
	defer func() { mlog.Tracef("err=%v", err) }()

	if idx >= len(sub.Names) || idx < 0 {
		return fmt.Errorf("idx(%d) over", idx)
	}

	if sub.Attrs[idx] != nil {
		return nil
	}

	count := 0
	xmsg := NewXMsgWithCreator(cfgc.ProgId)
	xmsg.Add("cmd", []byte("bget"))
	for i := idx; i < len(sub.Names); i++ {
		if sub.Attrs[i] != nil {
			break
		}
		path := sub.GetPath(idx)
		sub.Attrs[i], err = cfgc.GetObject(path, sub.Ocs[i])
		if sub.Attrs[i] != nil {
			break
		}
		key := fmt.Sprintf("oc%d", count)
		xmsg.Add(key, mutils.BytesLittleEndianUint16(sub.Ocs[i]))
		key2 := fmt.Sprintf("path%d", count)
		xmsg.Add(key2, []byte(path))
		count++
		if count > BAT_READ_OBJ_COUNT {
			break
		}
	}

	if count == 0 {
		return nil
	}

	xmsg.Add("count", []byte(fmt.Sprintf("%d", count)))

	if err = cfgc.send(xmsg); err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		rsp, rerr := cfgc.recv()
		if rerr != nil {
			return rerr
		}
		sub.Attrs[idx+i] = rsp
	}
	return nil
}
func (cfgc *Cfgc) ReadObject(sub *CfgGroup) (path string, oc uint16, obj *XMsg, err error) {
	if sub.curr >= sub.maxNames {
		return "", 0, nil, fmt.Errorf("no object")
	}
	oc = sub.Ocs[sub.curr]
	name := sub.Names[sub.curr]

	if sub.Attrs[sub.curr] == nil {
		cfgc.GetSubObjects(sub, sub.curr)
	}
	if sub.Attrs[sub.curr] != nil {
		obj = sub.Attrs[sub.curr]
	}
	sub.curr++
	return name, oc, obj, nil
}

func (cfgc *Cfgc) Save(path string) (err error) {
	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("save"))
	req.Add("path", []byte(path))

	_, err = cfgc.call(req)
	return err
}
func (cfgc *Cfgc) replace(path string, slow uint32) (err error) {
	req := NewXMsgWithCreator(cfgc.ProgId)
	req.Add("cmd", []byte("replace"))
	req.Add("path", []byte(path))
	req.Add("slow", mutils.BytesLittleEndianUint32(slow))

	_, err = cfgc.call(req)

	return err
}
func (cfgc *Cfgc) Replace(path string) (err error) {
	return cfgc.replace(path, 0)
}
func (cfgc *Cfgc) ReplaceSlow(path string) (err error) {
	return cfgc.replace(path, 1)
}

func NewCfgc(progId, zone, ns string) (*Cfgc, error) {
	cfgc := new(Cfgc)
	cfgc.ProgId = progId
	cfgc.Zone = zone
	cfgc.Ns = ns

	if err := cfgc.Open(); err != nil {
		return nil, err
	}

	return cfgc, nil
}
