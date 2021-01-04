/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:22:31
 * @LastEditTime: 2020-12-16 14:22:32
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcom

import (
	"bufio"
	"bytes"
	"container/list"
	"crypto/tls"
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase"
	"github.com/ChenLong-dev/gobase/mbase/msys"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	reconnTimeProtect   int64 = int64(3 * time.Second) //1000毫秒
	maxWaitResponseTime int64 = int64(10 * time.Second)
	hctxDyingTime       int64 = 180 * 1000 //3分钟
)

type httpAddr struct {
	scheme string
	host   string
	port   string
}

func (haddr *httpAddr) String() string {
	return fmt.Sprintf("%v://%v:%v", haddr.scheme, haddr.host, haddr.port)
}
func NewHttpAddr(u *url.URL) *httpAddr {
	haddr := &httpAddr{}
	haddr.scheme = u.Scheme
	haddr.host = u.Hostname()
	haddr.port = u.Port()
	if haddr.port == "" {
		if strings.ToLower(haddr.scheme) == "https" {
			haddr.port = "443"
		} else {
			haddr.port = "80"
		}
	}
	return haddr
}

type httpCall struct {
	req        *http.Request
	reqMarshal bytes.Buffer
	res        *http.Response
	err        error

	rchan chan int
}
type httpConnectContext struct {
	conn          net.Conn
	connTimestamp int64
	waitSendQ     *list.List
	waitRecvQ     *list.List
	wNotify       chan int
	rNotify       chan int
	maxRequest    int64
	currRequest   int64
	maxIdleAge    int64
	lastUsed      int64 //	最近一次使用时间(按发送)
	lastActive    int64 //	最近一次活跃时间(按接受)
	dying         int64
	dyingMT       *msys.MTimer
	//sendCnt			int64
	//recvCnt			int64
}
type HttpMixConnect struct {
	sync.Mutex
	haddr httpAddr

	MaxIdleAge        int64
	MaxRequestPerConn int64
	currActiveConn    int64

	hctx *httpConnectContext

	statReconnectCnt       int64
	statResetCnt           int64
	statResetRequestCnt    int64
	statResetIdleCnt       int64
	statResetRspTimeoutCnt int64
	statResetReadCnt       int64
	statResetWriteCnt      int64
}

func (hmc *HttpMixConnect) String() string {
	return fmt.Sprintf("{haddr=%v,MaxIdleAge:%v,MaxRequestPerConn:%d,currActiveConn:%d,"+
		"reconnect:%d,reset:%d,resetRequest:%d,resetIdle:%d,resetRspTimeout:%d,resetRead:%d,resetWrite:%d}",
		hmc.haddr, time.Duration(hmc.MaxIdleAge), hmc.MaxRequestPerConn, hmc.CountActiveConn(),
		atomic.LoadInt64(&hmc.statReconnectCnt), atomic.LoadInt64(&hmc.statResetCnt), atomic.LoadInt64(&hmc.statResetRequestCnt), atomic.LoadInt64(&hmc.statResetIdleCnt),
		atomic.LoadInt64(&hmc.statResetRspTimeoutCnt), atomic.LoadInt64(&hmc.statResetReadCnt), atomic.LoadInt64(&hmc.statResetWriteCnt))
}

func (hmc *HttpMixConnect) notify(nc chan int) {
	select {
	case nc <- 1:
		return
	default:
		return
	}
}
func (hmc *HttpMixConnect) needReset(hctx *httpConnectContext, timestamp int64) bool {
	if hctx == nil || hctx.conn == nil {
		return true
	}
	if hctx.maxRequest > 0 && hctx.currRequest >= hctx.maxRequest {
		atomic.AddInt64(&hmc.statResetRequestCnt, 1)
		return true
	}

	var flyCnt int
	if hctx.waitSendQ != nil {
		flyCnt += hctx.waitSendQ.Len()
	}
	if hctx.waitRecvQ != nil {
		flyCnt += hctx.waitRecvQ.Len()
	}

	if flyCnt == 0 {
		if hctx.maxIdleAge > 0 && timestamp-atomic.LoadInt64(&hctx.lastUsed) > hctx.maxIdleAge {
			atomic.AddInt64(&hmc.statResetIdleCnt, 1)
			return true
		}
	} else {
		if timestamp-atomic.LoadInt64(&hctx.lastActive) >= maxWaitResponseTime {
			atomic.AddInt64(&hmc.statResetRspTimeoutCnt, 1)
			return true
		}
	}
	return false
}

func (hmc *HttpMixConnect) RoundTrip(req *http.Request) (res *http.Response, err error) {
	mbase.Tracef("hmc=%v,req=%v", hmc, req)
	defer func() { mbase.Tracef("err=%v,res=%v", err, res) }()

	now := time.Now()
	timestamp := now.UnixNano()

	hcall := &httpCall{req: req, rchan: make(chan int, 1)}
	err = req.Write(&hcall.reqMarshal)
	if err != nil {
		return nil, err
	}

	hmc.Lock()

	hctx := hmc.hctx
	//if hctx == nil || hctx.conn == nil || (hmc.MaxIdleAge > 0 && now.Sub(hmc.lastUsed) > hmc.MaxIdleAge) || (hctx.maxRequest > 0 && hctx.currRequest >= hctx.maxRequest) {
	if hmc.needReset(hctx, timestamp) {
		newHctx, newErr := hmc.build()
		if newErr == nil {
			if hctx != nil {
				atomic.StoreInt64(&hctx.dying, 1)
				if hctx.waitSendQ != nil && hctx.waitRecvQ != nil && hctx.waitSendQ.Len()+hctx.waitRecvQ.Len() == 0 {
					hmc.reset(hctx, fmt.Errorf("timeout MaxIdleAge"))
				} else if hctx.dyingMT == nil {
					f := func(pridata interface{}) {
						h := pridata.(*httpConnectContext)
						hctx.dyingMT = nil
						hmc.ResetHctx(h, fmt.Errorf("timeout wait response"))
					}
					hctx.dyingMT = msys.StartMTimer(hctxDyingTime, f, hctx)
				}
			}
			hmc.hctx = newHctx
			hctx = hmc.hctx
		} else if hctx == nil || hctx.conn == nil {
			hmc.Unlock()
			return nil, fmt.Errorf("connect %v error:%v", req.URL.Host, err)
		}
	}

	atomic.StoreInt64(&hctx.lastUsed, timestamp)
	if hctx.waitSendQ.Len()+hctx.waitRecvQ.Len() == 0 {
		atomic.StoreInt64(&hctx.lastActive, timestamp)
	}

	hctx.waitSendQ.PushBack(hcall)
	hctx.currRequest++

	hmc.notify(hctx.wNotify)

	hmc.Unlock()

	select {
	case r := <-hcall.rchan:
		if r == 0 && hcall.err == nil {
			return hcall.res, fmt.Errorf("network disconnect")
		}
		return hcall.res, hcall.err
	case <-req.Cancel:
		return nil, fmt.Errorf("user cancel")
	}
}

func (hmc *HttpMixConnect) connect() (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 15 * time.Second,
	}
	var conn net.Conn
	var err error
	var host string = hmc.haddr.host
	if ip, rerr := mutils.ResolveDns(hmc.haddr.host); rerr == nil {
		host = ip
	}
	if strings.ToLower(hmc.haddr.scheme) == "https" {
		config := tls.Config{ServerName: hmc.haddr.host}
		conn, err = tls.DialWithDialer(dialer, "tcp", fmt.Sprintf("%s:%s", host, hmc.haddr.port), &config)
	} else {
		conn, err = dialer.Dial("tcp", fmt.Sprintf("%s:%s", host, hmc.haddr.port))
	}
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (hmc *HttpMixConnect) build() (*httpConnectContext, error) {
	//hctx := hmc.hctx
	//hmc.reset(hctx, fmt.Errorf("need rebuild"))

	conn, err := hmc.connect()
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().UnixNano()
	atomic.AddInt64(&hmc.currActiveConn, 1)
	hctx := &httpConnectContext{}
	hctx.conn = conn
	hctx.connTimestamp = timestamp
	hctx.lastUsed = timestamp
	hctx.lastActive = timestamp
	hctx.maxRequest = hmc.MaxRequestPerConn
	hctx.maxIdleAge = hmc.MaxIdleAge
	hctx.waitSendQ = list.New()
	hctx.waitRecvQ = list.New()
	hctx.wNotify = make(chan int, 1)
	hctx.rNotify = make(chan int, 1)
	//hmc.hctx = hctx

	go hmc.writeLoop(hctx)
	go hmc.readLoop(hctx)
	//go hmc.watchLoop(hmc.rNotify)

	//hmc.lastUsed = time.Now().UnixNano()

	return hctx, nil
}

/*
func (hmc *HttpMixConnect) watchLoop(rNotify chan int) {
	for {
		select {
		case
		}
	}
}*/

func (hmc *HttpMixConnect) readLoop(hctx *httpConnectContext) (err error) {
	mbase.Tracef("entry...")

	rNotify := hctx.rNotify
	conn := hctx.conn
	waitRecvQ := hctx.waitRecvQ
	waitSendQ := hctx.waitSendQ
	if rNotify == nil || conn == nil || waitRecvQ == nil {
		return
	}
	reader := bufio.NewReader(conn)
	defer func() { hmc.ResetHctx(hctx, err) }()

	reconn := func() error {
		hmc.Lock()
		defer hmc.Unlock()
		nowTime := time.Now().UnixNano()
		if atomic.LoadInt64(&hctx.dying) != 0 || hctx.conn == nil || waitSendQ.Len()+waitRecvQ.Len() == 0 || (nowTime-hctx.connTimestamp < reconnTimeProtect && hctx.conn == conn) {
			//	已经被重置或者没有数据待发送
			if hctx.conn == conn {
				atomic.AddInt64(&hmc.statResetReadCnt, 1)
			}
			return fmt.Errorf("reset")
		}

		if hctx.conn == conn {
			if reerr := hmc.reconnect(hctx); reerr != nil {
				if hctx.conn == conn {
					atomic.AddInt64(&hmc.statResetReadCnt, 1)
				}
				return reerr
			}
		}

		conn = hctx.conn
		reader = bufio.NewReader(conn)
		return nil
	}

	for {
		if _, err = reader.Peek(1); err != nil {
			mbase.Tracef("err=%v", err)
			if rerr := reconn(); rerr != nil {
				return err
			}
			continue
		}

		timestamp := time.Now().UnixNano()
		var hcall *httpCall = nil
		hmc.Lock()
		atomic.StoreInt64(&hctx.lastActive, timestamp)
		if waitRecvQ.Len() > 0 {
			hcall = waitRecvQ.Remove(waitRecvQ.Front()).(*httpCall)
		}
		hmc.Unlock()

		if hcall == nil {
			//	有数据抵达，但是响应队列里没有等待的请求，通常是通道被关闭了或者重连了
			//return fmt.Errorf("has data recv but waitRecvQ is empty")
			if reerr := reconn(); reerr != nil {
				return reerr
			}
			continue
		}

		hcall.res, hcall.err = http.ReadResponse(reader, hcall.req)
		//atomic.AddInt64(&hctx.recvCnt, 1)

		if hcall.err == nil && hcall.res.ContentLength != 0 {
			//bufio.NewReader(hcall.res.Body)
			data, rerr := ioutil.ReadAll(hcall.res.Body)
			hcall.res.Body.Close()
			hcall.res.Body = nil
			if rerr != nil {
				hcall.err = rerr
			} else {
				hcall.res.Body = mutils.NewByteReadCloser(data)
			}
		}

		if hcall.err == nil {
			hcall.rchan <- 1
		} else {
			//close(hcall.rchan)
			//return hcall.err
			close(hcall.rchan) //	这个不处理了，直接当出错关闭！因为头部正常接收，内容出错很少见
			if rerr := reconn(); rerr != nil {
				return err
			}
		}

		hmc.Lock()
		if waitSendQ.Len()+waitRecvQ.Len() == 0 && atomic.LoadInt64(&hctx.dying) != 0 {
			hmc.Unlock()
			return fmt.Errorf("maxRequest limit")
		}
		hmc.Unlock()
	}
}
func isRequestCanceled(req *http.Request) bool {
	if req.Cancel == nil {
		return false
	}
	select {
	case <-req.Cancel:
		return true
	default:
		return false
	}
}
func (hmc *HttpMixConnect) writeLoop(hctx *httpConnectContext) {
	mbase.Tracef("entry...")

	wNotify := hctx.wNotify
	conn := hctx.conn
	waitSendQ := hctx.waitSendQ
	waitRecvQ := hctx.waitRecvQ
	if wNotify == nil || conn == nil || waitSendQ == nil || waitRecvQ == nil {
		return
	}

	reconn := func() error {
		hmc.Lock()
		defer hmc.Unlock()
		nowTime := time.Now().UnixNano()
		if atomic.LoadInt64(&hctx.dying) != 0 || hctx.conn == nil || waitSendQ.Len()+waitRecvQ.Len() == 0 || (nowTime-hctx.connTimestamp < reconnTimeProtect && hctx.conn == conn) {
			//	已经被重置或者没有数据待发送
			if hctx.conn == conn {
				atomic.AddInt64(&hmc.statResetWriteCnt, 1)
			}
			return fmt.Errorf("reset")
		}

		if hctx.conn == conn {
			if reerr := hmc.reconnect(hctx); reerr != nil {
				if hctx.conn == conn {
					atomic.AddInt64(&hmc.statResetWriteCnt, 1)
				}
				return reerr
			}
		}

		conn = hctx.conn
		return nil
	}
	for {
		var sendHcall *httpCall = nil
		hmc.Lock()
		for waitSendQ.Len() > 0 {
			hcall := waitSendQ.Remove(waitSendQ.Front()).(*httpCall)
			if isRequestCanceled(hcall.req) {
				hcall.err = fmt.Errorf("user cancel")
				close(hcall.rchan)
			} else {
				sendHcall = hcall
				waitRecvQ.PushBack(hcall)
				break
			}
		}
		hmc.Unlock()

		if sendHcall == nil {
			if w := <-wNotify; w == 0 { // 等待有数据发送。 w为0表示被关闭
				return
			}
		} else {
			//err := req.Write(conn)
			_, err := mutils.WriteN(conn, sendHcall.reqMarshal.Bytes())
			//atomic.AddInt64(&hctx.sendCnt, 1)
			if err != nil {
				mbase.Tracef("hctx=%v,err=%v", hctx, err)
				//hmc.ResetHctx(hctx, err)
				//return
				if rerr := reconn(); rerr != nil {
					hmc.ResetHctx(hctx, err)
					return
				}
			}
		}
	}
}

func clearQ(q *list.List, err error) {
	if q == nil {
		return
	}

	for q.Len() > 0 {
		hcall := q.Remove(q.Front()).(*httpCall)
		hcall.err = err
		close(hcall.rchan)
	}
}
func (hmc *HttpMixConnect) reset(hctx *httpConnectContext, err error) {
	if hctx == nil {
		return
	}

	mt := hctx.dyingMT
	hctx.dyingMT = nil
	if mt != nil {
		mt.Stop()
	}

	conn := hctx.conn
	hctx.conn = nil
	if conn != nil {
		conn.Close()
		atomic.AddInt64(&hmc.currActiveConn, -1)
		atomic.AddInt64(&hmc.statResetCnt, 1)
	}

	waitSendQ := hctx.waitSendQ
	hctx.waitSendQ = nil
	if waitSendQ != nil {
		clearQ(waitSendQ, err)
	}

	waitRecvQ := hctx.waitRecvQ
	hctx.waitRecvQ = nil
	if waitRecvQ != nil {
		clearQ(waitRecvQ, err)
	}

	rNotify := hctx.rNotify
	hctx.rNotify = nil
	if rNotify != nil {
		close(rNotify)
	}

	wNotify := hctx.wNotify
	hctx.wNotify = nil
	if wNotify != nil {
		close(wNotify)
	}
}
func (hmc *HttpMixConnect) ResetHctx(hctx *httpConnectContext, err error) {
	mbase.Tracef("hmc=%v,hctx=%v,err=%v", hmc, hctx, err)

	hmc.Lock()
	defer hmc.Unlock()

	hmc.reset(hctx, err)
}

func (hmc *HttpMixConnect) reconnect(hctx *httpConnectContext) error {
	if hctx == nil {
		return fmt.Errorf("hctx is nil")
	}

	conn, err := hmc.connect()
	if err != nil {
		return err
	}

	if hctx.conn != nil {
		hctx.conn.Close()
		atomic.AddInt64(&hmc.currActiveConn, -1)
		atomic.AddInt64(&hmc.statReconnectCnt, 1)
	}
	hctx.conn = conn
	timestamp := time.Now().UnixNano()
	hctx.connTimestamp = timestamp
	atomic.StoreInt64(&hctx.lastActive, timestamp)
	atomic.StoreInt64(&hctx.lastUsed, timestamp)
	atomic.AddInt64(&hmc.currActiveConn, 1)

	wq := hctx.waitSendQ
	e := wq.Front()
	for e != nil {
		hcall := e.Value.(*httpCall)
		ce := e
		e = e.Next()
		if isRequestCanceled(hcall.req) {
			hcall.err = fmt.Errorf("user cancel")
			close(hcall.rchan)
			wq.Remove(ce)
		}
	}

	rq := hctx.waitRecvQ
	for rq.Len() > 0 {
		hcall := rq.Remove(rq.Back()).(*httpCall)
		if !isRequestCanceled(hcall.req) {
			wq.PushFront(hcall)
		} else {
			hcall.err = fmt.Errorf("user cancel")
			close(hcall.rchan)
		}
	}

	//	atomic.StoreInt64(&hctx.sendCnt, 0)
	//	atomic.StoreInt64(&hctx.recvCnt, 0)

	if hctx.waitSendQ.Len() > 0 {
		hmc.notify(hctx.wNotify)
	}

	return nil
}
func (hmc *HttpMixConnect) Close(err error) {
	mbase.Tracef("hmc=%v", hmc)

	hmc.Lock()
	defer hmc.Unlock()

	hmc.reset(hmc.hctx, err)
	hmc.hctx = nil
}
func (hmc *HttpMixConnect) CountActiveConn() int64 {
	if hmc == nil {
		return 0
	}
	return atomic.LoadInt64(&hmc.currActiveConn)
}

func NewHttpMixConnect(haddr *httpAddr, maxIdleAge time.Duration, maxRequest int64) (*HttpMixConnect, error) {
	mbase.Tracef("haddr=%v", haddr)

	hmc := &HttpMixConnect{haddr: *haddr, MaxIdleAge: int64(maxIdleAge), MaxRequestPerConn: maxRequest}
	/*
		err := hmc.connect()
		if err != nil {
			return nil, err
		}*/
	return hmc, nil
}

type HttpMixTransport struct {
	conns      sync.Map
	maxIdleAge time.Duration
	maxRequest int64
}

func (hmt *HttpMixTransport) String() string {
	var cnt int64
	var conns string = "{\n"

	hmt.conns.Range(func(key, val interface{}) bool {
		hmc := val.(*HttpMixConnect)
		cnt += hmc.CountActiveConn()
		conns += "\t" + hmc.String() + "\n"
		return true
	})
	conns += "}\n"

	return fmt.Sprintf("{AllActiveConn:%d,conns:\n%s}", cnt, conns)
}
func (hmt *HttpMixTransport) CountAllActiveConn() int64 {
	var cnt int64
	hmt.conns.Range(func(key, val interface{}) bool {
		hmc := val.(*HttpMixConnect)
		cnt += hmc.CountActiveConn()
		return true
	})

	return cnt
}
func (hmt *HttpMixTransport) CountActiveConn(serverAddr string) int64 {
	if c, ok := hmt.conns.Load(serverAddr); ok {
		return c.(*HttpMixConnect).CountActiveConn()
	}
	return 0
}
func (hmt *HttpMixTransport) getConn(req *http.Request) (hmc *HttpMixConnect, err error) {
	haddr := NewHttpAddr(req.URL)

	if c, ok := hmt.conns.Load(haddr.String()); ok {
		return c.(*HttpMixConnect), nil
	}

	f := func() (interface{}, error) {
		if v, ok := hmt.conns.Load(haddr.String()); ok {
			return v, nil
		}
		hmc, err = NewHttpMixConnect(haddr, hmt.maxIdleAge, hmt.maxRequest)
		if err == nil {
			hmt.conns.Store(haddr.String(), hmc)
		}
		return hmc, err
	}
	v, verr := mutils.SingleFlight(fmt.Sprintf("mbase.HttpMixConnect:%v", haddr.String()), f)
	return v.(*HttpMixConnect), verr
}
func (hmt *HttpMixTransport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	mbase.Tracef("req=%v", req)
	defer func() { mbase.Tracef("err=%v,res=%v", err, res) }()

	hmc, gerr := hmt.getConn(req)
	if gerr != nil {
		return nil, gerr
	}

	return hmc.RoundTrip(req)
}

func NewHttpMixTransport(maxIdleAge time.Duration, maxRequest int64) *HttpMixTransport {
	return &HttpMixTransport{maxIdleAge: maxIdleAge, maxRequest: maxRequest}
}
