/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:03:56
 * @LastEditTime: 2020-12-17 09:03:56
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import (
	"container/list"
	"sync"
	"sync/atomic"
)

/*
author: yhx
提供有序调用服务。同一时刻相同key函数只允许一个在运行，按调用进入次序有序调用。
*/

/*
type PipeFunc func() (interface{}, error)

type _op struct {
	ac chan bool
}

type PipeOp struct {
	qc chan *_op
	rc chan bool
}
func (pop *PipeOp) Running() {
	for op := range pop.qc {
		op.ac <- true
		<- pop.rc
	}
}
func (pop *PipeOp) Do(fn PipeFunc) (interface{}, error) {
	op := &_op{ac:make(chan bool)}
	pop.qc <- op
	<- op.ac
	r, err := fn()
	pop.rc <- true
	return r, err
}

var stdPipeOps = &sync.Map{}
func PipeOperate(key string, fn PipeFunc) (interface{}, error) {
	var pop *PipeOp
	if val, ok := stdPipeOps.Load(key); ok {
		pop = val.(*PipeOp)
	} else {
		pop = &PipeOp{qc:make(chan *_op, 10), rc:make(chan bool)}
		val, ok = stdPipeOps.LoadOrStore(key, pop)
		if !ok {
			go pop.Running()
		}
		pop = val.(*PipeOp)
	}

	return pop.Do(fn)
}
*/

type PipeFunc func() (interface{}, error)
type PipeOp struct {
	waitN int32
	mu    sync.Mutex
}
type PGroup struct {
	mu sync.Mutex
	m  map[string]interface{}
}

func (g *PGroup) Do(key string, f PipeFunc) (interface{}, error) {
	var pop *PipeOp
	g.mu.Lock()
	if v, ok := g.m[key]; ok {
		pop = v.(*PipeOp)
		pop.waitN++
		//atomic.AddInt32(&pop.waitN, 1)
	} else {
		pop = &PipeOp{waitN: 1}
		g.m[key] = pop
	}
	g.mu.Unlock()

	pop.mu.Lock()
	r, err := f()
	pop.mu.Unlock()
	g.mu.Lock()
	if pop.waitN--; pop.waitN == 0 {
		delete(g.m, key)
	}
	g.mu.Unlock()

	return r, err
}

func NewPGroup() *PGroup {
	return &PGroup{m: make(map[string]interface{})}
}

var stdPGroup = NewPGroup()

func PipeOperate(key string, fn PipeFunc) (interface{}, error) {
	return stdPGroup.Do(key, fn)
}

//////////////////
type PipelineFunc func() error
type PipelineOp struct {
	mu     sync.Mutex
	g      *PipelineGroup
	key    string
	lst    *list.List
	flying int64
}

func (pop *PipelineOp) run() {
	for {
		var f PipelineFunc = nil
		pop.mu.Lock()
		if e := pop.lst.Front(); e != nil {
			f = pop.lst.Remove(e).(PipelineFunc)
		}
		pop.mu.Unlock()

		if f != nil {
			f()
			atomic.AddInt64(&pop.flying, -1)
		} else {
			if pop.flyingCount() == 0 {
				if pop.g.delete(pop.key) {
					return //	退出
				}
			}
		}
	}
}
func (pop *PipelineOp) put(f PipelineFunc) {
	atomic.AddInt64(&pop.flying, 1)
	pop.mu.Lock()
	pop.lst.PushBack(f)
	pop.mu.Unlock()
}
func (pop *PipelineOp) flyingCount() int64 {
	return atomic.LoadInt64(&pop.flying)
}
func NewPipelineOp(g *PipelineGroup, key string, f PipelineFunc) *PipelineOp {
	pop := &PipelineOp{g: g, key: key}
	pop.lst = list.New()
	if f != nil {
		atomic.AddInt64(&pop.flying, 1)
		pop.lst.PushBack(f)
	}

	go pop.run()
	return pop
}

type PipelineGroup struct {
	mu sync.Mutex
	m  map[string]*PipelineOp
}

func (g *PipelineGroup) delete(key string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if pop, ok := g.m[key]; ok {
		if pop.flyingCount() > 0 {
			return false
		}
	}
	delete(g.m, key)
	return true
}
func (g *PipelineGroup) Put(key string, f PipelineFunc) error {
	var pop *PipelineOp
	g.mu.Lock()
	if v, ok := g.m[key]; ok {
		pop = v
		pop.put(f)
		//atomic.AddInt32(&pop.waitN, 1)
	} else {
		pop = NewPipelineOp(g, key, f)
		g.m[key] = pop
	}
	g.mu.Unlock()

	return nil
}

func NewPipelineGroup() *PipelineGroup {
	g := &PipelineGroup{m: make(map[string]*PipelineOp)}
	return g
}

var stdPipelineGroup = NewPipelineGroup()

func Pipeline(key string, f PipelineFunc) error {
	return stdPipelineGroup.Put(key, f)
}
