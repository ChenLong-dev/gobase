package mcom

import (
	"container/list"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrTimeout is used when we reach an IO deadline
	ErrTimeout = fmt.Errorf("i/o deadline reached")
	ErrStreamClosed = fmt.Errorf("stream closed")
)

type MStream struct {
	msession	*MSession
	id 			uint64
	src 		string
	dst 		string
	Ext 		string
	mctx 		*MContext
	writer 		mstreamWriter
	reader 		mstreamReader
}
func (mstream *MStream) SetMContext(mctx *MContext) {
	mstream.mctx = mctx
}
func (mstream *MStream) Read(b []byte) (n int, err error) {
	mstream.mctx.Tracef("mstream=%v,len(b)=%d", mstream, len(b))
	defer func() {mstream.mctx.Tracef("mstream=%v,err=%v", mstream, err)} ()

	return mstream.reader.Read(mstream.mctx, b)
}

func (mstream *MStream) Write(b []byte) (n int, err error) {
	mstream.mctx.Tracef("mstream=%v,len(b)=%d", mstream, len(b))
	defer func() {mstream.mctx.Tracef("mstream=%v,err=%v", mstream, err)} ()

	return mstream.writer.Write(mstream.mctx, b)
}

// LocalAddr returns the local network address.
func (mstream *MStream) LocalAddr() net.Addr {
	return mstream.msession.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (mstream *MStream) RemoteAddr() net.Addr {
	return mstream.msession.RemoteAddr()
}
func (mstream *MStream) Src() string {
	return mstream.src
}
func (mstream *MStream) Dst() string {
	return mstream.dst
}
// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (mstream *MStream) SetDeadline(t time.Time) error {
	mstream.reader.SetDeadline(t)
	mstream.writer.SetDeadline(t)
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (mstream *MStream) SetReadDeadline(t time.Time) error {
	mstream.reader.SetDeadline(t)
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (mstream *MStream) SetWriteDeadline(t time.Time) error {
	mstream.writer.SetDeadline(t)
	return nil
}

func (mstream *MStream) shutdown(reason string) {
	mstream.reader.Close(mstream.mctx)
	mstream.writer.Close(mstream.mctx)
	mstream.msession.closeMStream(mstream.mctx, mstream.id, reason)
}
func (mstream *MStream) Close() error {
	mstream.mctx.Tracef("mstream=%v", mstream)

	mstream.shutdown("Close")
	return nil
}
func (mstream *MStream) onClose(mctx *MContext, err error) {
	mctx.Tracef("mstream=%v,err=%v", mstream, err)

	mstream.shutdown(err.Error())
}
func (mstream *MStream) onUpdateWindow(mctx *MContext, win int32) {
	mctx.Tracef("mstream=%v,win=%d", mstream, win)

	mstream.writer.UpdateWin(mctx, win)
}
func (mstream *MStream) onData(mctx *MContext, data []byte) {
	mctx.Tracef("mstream=%v,len(data)=%d", mstream, len(data))

	mstream.mctx = mctx

	mstream.reader.onData(mctx, data)
}

func newMStream(msession *MSession, id uint64, src, dst string, localWin, peerWin int32, ext string) *MStream {
	mstream := &MStream{
		msession:msession,
		id:id,
		src:src,
		dst:dst,
		Ext:ext,
		writer:mstreamWriter{win:peerWin},
		//reader:mstreamReader{},
	}

	mstream.writer.waitList.Init()
	mstream.writer.w = func(mctx *MContext, b []byte) (err error) {
		return mstream.msession.sendDataRaw(mctx, id, b)
	}

	mstream.reader.caches.Init()
	mstream.reader.waitList.Init()
	mstream.reader.updateWindow = func(mctx *MContext, addWin int32) error {
		return mstream.msession.updateWindow(mctx, id, addWin)
	}

	return mstream
}

type mstreamWriter struct {
	win 		int32
	closed 		int32
	mu 			sync.Mutex
	waitList 	list.List
	w			func(mctx *MContext, b []byte) (err error)

	deadline 	atomic.Value
}
func (writer *mstreamWriter) handleWrite(mctx *MContext, b []byte, waitElemnt *list.Element) (n int, err error) {

	blen := int32(len(b))
	//waiter := waitElemnt.Value.(*mstreamWriteWaiter)
	mctx.Tracef("writer=%v,len(data)=%d", writer, len(b))
	defer func() {mctx.Tracef("writer=%v,n=%d,err=%v", writer, n, err)} ()

	writer.mu.Lock()
	canSend := blen
	if canSend > writer.win {
		canSend = writer.win
	}
	writer.win -= canSend
	writer.mu.Unlock()

	err = writer.w(mctx, b[:canSend])	//	能发的先发走
	if err != nil {
		return 0, err
	}
	if canSend == blen {
		writer.mu.Lock()
		writer.waitList.Remove(waitElemnt)
		writer.notify(mctx)
		writer.mu.Unlock()
	}
	return int(canSend), nil
}
func (writer *mstreamWriter) Write(mctx *MContext, b []byte) (n int, err error) {
	blen := int32(len(b))
	mctx.Tracef("writer=%v,len(b)=%d", writer, blen)
	defer func() { mctx.Tracef("writer=%v,n=%d,err=%v", writer, n, err) } ()

	if blen == 0 {
		return 0, nil
	}
	if atomic.LoadInt32(&writer.closed) != 0 {
		return 0, ErrStreamClosed
	}

	c := make(chan int, 1)
	writer.mu.Lock()
	waitElement := writer.waitList.PushBack(c)
	isFirst := writer.waitList.Len() == 1
	writer.mu.Unlock()

	if isFirst {	//	抢到了队头，有优先发送权限
		if n, err = writer.handleWrite(mctx, b, waitElement); n == int(blen) || err != nil {
			return n, err
		}
	}

	//	等待发送
	for {
		var timeout <-chan time.Time
		var timer *time.Timer
		deadline := writer.deadline.Load()
		if deadline != nil && !deadline.(time.Time).IsZero() {
			delay := deadline.(time.Time).Sub(time.Now())
			timer = time.NewTimer(delay)
			timeout = timer.C
		}
		select {
		case <-c:
			if timer != nil {
				timer.Stop()
			}
			if atomic.LoadInt32(&writer.closed) != 0 {
				return n, ErrStreamClosed
			}
			wn, werr := writer.handleWrite(mctx, b[n:], waitElement)
			n += wn
			if n == int(blen) || werr != nil {
				return n, werr
			}
		case <-timeout:
			writer.mu.Lock()
			writer.waitList.Remove(waitElement)
			writer.notify(mctx)
			writer.mu.Unlock()
			return n, ErrTimeout
		}
	}
}
func (writer *mstreamWriter) notify(mctx *MContext) {
	mctx.Tracef("writer=%v", writer)

	if writer.win == 0 {
		return
	}
	front := writer.waitList.Front()
	if front != nil {
		//front.Value.(*mstreamWriteWaiter).Notify()
		select {
		case front.Value.(chan int) <- 1:
		default:
		}
	}
}
func (writer *mstreamWriter) UpdateWin(mctx *MContext, addWin int32) {
	mctx.Tracef("writer=%v,addWin=%d", writer, addWin)

	writer.mu.Lock()
	writer.win += addWin
	writer.notify(mctx)
	writer.mu.Unlock()
}
func (writer *mstreamWriter) SetDeadline(deadline time.Time) {
	writer.deadline.Store(deadline)
}
func (writer *mstreamWriter) Close(mctx *MContext) {
	mctx.Tracef("writer=%v", writer)

	if atomic.CompareAndSwapInt32(&writer.closed, 0, 1) {
		writer.mu.Lock()
		defer writer.mu.Unlock()

		for front := writer.waitList.Front(); front != nil; front = writer.waitList.Front() {
			ec := front.Value.(chan int)
			writer.waitList.Remove(front)
			close(ec)
		}
	}
}

type mstreamReader struct {
	closed 			int32
	caches 			list.List
	waitList 		list.List
	mu 				sync.Mutex
	updateWindow	func(mctx *MContext, addWin int32) error

	deadline 	atomic.Value
}
func (reader *mstreamReader) handleRead(mctx *MContext, b []byte, waitElement *list.Element) (n int) {
	mctx.Tracef("reader=%v,len(b)=%d", reader, len(b))
	defer func() {mctx.Tracef("reader=%v,n=%d", reader, n)} ()

	needRead := len(b)
	buffs := make([][]byte, 0, 10)
	reader.mu.Lock()
	for e := reader.caches.Front(); e != nil; e = reader.caches.Front() {
		cache := e.Value.([]byte)
		cacheLen := len(cache)
		if cacheLen == 0 {
			atomic.StoreInt32(&reader.closed, 1)
			reader.caches.Remove(e)
		} else if cacheLen <= needRead {
			needRead -= cacheLen
			buffs = append(buffs, cache)
			reader.caches.Remove(e)
		} else {
			//needRead = 0
			if needRead > 0 {
				buffs = append(buffs, cache[:needRead])
				e.Value = cache[needRead:]
				needRead = 0
			}

			break
		}
	}

	if needRead < len(b) || atomic.LoadInt32(&reader.closed) != 0 {
		reader.waitList.Remove(waitElement)
		if reader.caches.Len() > 0 && reader.waitList.Len() > 0 {
			reader.notify(mctx)
		}
	}
	reader.mu.Unlock()

	for _, buff := range buffs {
		copy(b[n:], buff)
		n += len(buff)
	}

	if atomic.LoadInt32(&reader.closed) != 0 {
		reader.Close(NewMContext())
	} else if n > 0 {
		reader.updateWindow(mctx, int32(n))
	}

	return n
}
func (reader *mstreamReader) Read(mctx *MContext, b []byte) (n int, err error) {
	mctx.Tracef("reader=%v,len(b)=%d", reader, len(b))
	defer func() {	mctx.Tracef("reader=%v,n=%d,err=%v", reader, n, err) } ()

	if len(b) == 0 {
		return 0, nil
	}
	if atomic.LoadInt32(&reader.closed) != 0 {
		return 0, io.EOF //ErrStreamClosed
	}

	c := make(chan int, 1)
	reader.mu.Lock()
	waitElement := reader.waitList.PushBack(c)
	isFirst := reader.waitList.Len() == 1
	reader.mu.Unlock()

	if isFirst {
		n = reader.handleRead(mctx, b, waitElement)
		if n > 0 {
			return n, nil
		}
	}

	//	等待接受
	for {
		var timeout <-chan time.Time
		var timer *time.Timer
		deadline := reader.deadline.Load()
		if deadline != nil && !deadline.(time.Time).IsZero() {
			delay := deadline.(time.Time).Sub(time.Now())
			timer = time.NewTimer(delay)
			timeout = timer.C
		}
		select {
		case <- c:
			if timer != nil {
				timer.Stop()
			}

			n += reader.handleRead(mctx, b[n:], waitElement)
			if n > 0 {
				return n, nil
			}

			if atomic.LoadInt32(&reader.closed) != 0 {
				return n, io.EOF//ErrStreamClosed
			}
		case <-timeout:
			reader.mu.Lock()
			reader.waitList.Remove(waitElement)
			reader.notify(mctx)
			reader.mu.Unlock()
			return n, ErrTimeout
		}
	}
}
func (reader *mstreamReader) notify(mctx *MContext) {
	mctx.Tracef("reader=%v,waiters=%d,caches=%d", reader, reader.waitList.Len(), reader.caches.Len())

	front := reader.waitList.Front()

	if front != nil && reader.caches.Len() > 0 {
		select {
		case front.Value.(chan int) <- 1:
		default:
		}
	}

}
func (reader *mstreamReader) onData(mctx *MContext, data []byte) {
	mctx.Tracef("reader=%v,len(data)=%d", reader, len(data))

	if len(data) == 0 {
		return
	}

	reader.mu.Lock()
	reader.caches.PushBack(data)
	reader.notify(mctx)
	reader.mu.Unlock()
}
func (reader *mstreamReader) SetDeadline(deadline time.Time) {
	reader.deadline.Store(deadline)
}
func (reader *mstreamReader) Close(mctx *MContext) {
	mctx.Tracef("reader=%v", reader)
/*
	if atomic.CompareAndSwapInt32(&reader.closed, 0, 1) {
		reader.mu.Lock()
		defer reader.mu.Unlock()

		for front := reader.waitList.Front(); front != nil; front = reader.waitList.Front() {
			ec := front.Value.(chan int)
			reader.waitList.Remove(front)
			close(ec)
		}
	}
 */
	reader.mu.Lock()
	defer reader.mu.Unlock()

	if reader.caches.Len() == 0 {
		atomic.StoreInt32(&reader.closed, 1)
		for front := reader.waitList.Front(); front != nil; front = reader.waitList.Front() {
			ec := front.Value.(chan int)
			reader.waitList.Remove(front)
			close(ec)
		}
	} else if atomic.LoadInt32(&reader.closed) == 0 {
		reader.caches.PushBack(make([]byte, 0))
		reader.notify(mctx)
	}


}