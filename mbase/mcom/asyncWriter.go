/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:18:56
 * @LastEditTime: 2020-12-26 15:53:48
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcom

import (
	"container/list"
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase/mutils"
	"github.com/ChenLong-dev/gobase/mlog"
	"io"
	"sync"
	"sync/atomic"
)

type AsyncWriter struct {
	closed   int32
	mu       sync.Mutex
	waitList *list.List
	writer   io.Writer
	notifyC  chan int
}

func NewAsyncWriter(writer io.Writer) *AsyncWriter {
	awriter := &AsyncWriter{waitList: list.New(), writer: writer, notifyC: make(chan int, 2)}
	go awriter.writeLoop()
	return awriter
}
func (awriter *AsyncWriter) Close() {
	mlog.Tracef("awriter=%v", awriter)

	if !atomic.CompareAndSwapInt32(&awriter.closed, 0, 1) {
		return
	}
	awriter.notify()
}
func (awriter *AsyncWriter) String() string {
	if awriter == nil {
		return "{nil}"
	} else {
		awriter.mu.Lock()
		wcnt := awriter.waitList.Len()
		awriter.mu.Unlock()
		return fmt.Sprintf("{closed:%t,waitListLen:%d}", awriter.IsClosed(), wcnt)
	}
}
func (awriter *AsyncWriter) IsClosed() bool {
	return atomic.LoadInt32(&awriter.closed) != 0
}
func (awriter *AsyncWriter) notify() {

	select {
	case awriter.notifyC <- 1:
	default:
	}
}
func (awriter *AsyncWriter) writeLoop() {
	for !awriter.IsClosed() {
		var p []byte = nil
		awriter.mu.Lock()
		if e := awriter.waitList.Front(); e != nil {
			p = awriter.waitList.Remove(e).([]byte)
		}
		awriter.mu.Unlock()

		if p == nil {
			<-awriter.notifyC
		} else {
			n, err := mutils.WriteN(awriter.writer, p)
			mlog.Tracef("awriter=%v,WriterN(%d) return n=%d,err=%v", awriter, len(p), n, err)
			if err != nil {
				awriter.Close()
				break
			}
		}
	}
}
func (awriter *AsyncWriter) Write(p []byte) (n int, err error) {
	mlog.Tracef("awriter=%v,len(p)=%d", awriter, len(p))
	defer func() { mlog.Tracef("n=%d,err=%v", n, err) }()

	if awriter.IsClosed() {
		return 0, fmt.Errorf("already closed")
	}

	awriter.mu.Lock()
	awriter.waitList.PushBack(p)
	awriter.mu.Unlock()

	awriter.notify()

	return len(p), nil
}
