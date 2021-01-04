/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:46:45
 * @LastEditTime: 2020-12-16 14:46:45
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcontainer

import (
	"container/list"
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase/msys"
	"sync"

	"github.com/ocdogan/rbt"
)

type MTLruNode struct {
	Key rbt.RbKey
	Val interface{}
	le  *list.Element
	t   uint64
}
type MTLru struct {
	sync.RWMutex
	l       *list.List
	rbt     *Mrbtree
	maxN    int
	timeout uint64
	timer   *msys.MTimer
}

func cleanTTimeout(pridata interface{}) {
	mtlru := pridata.(*MTLru)
	mtlru.Lock()
	defer mtlru.Unlock()

	mtlru.cleanTimeout()

	if mtlru.timer != nil {
		mtlru.timer = msys.StartMTimer(int64(mtlru.timeout), cleanTTimeout, mtlru)
	}
}
func (mtlru *MTLru) init() {
	mtlru.l = list.New()
	mtlru.rbt = MrbtreeNew()
	if mtlru.timeout > 0 {
		mtlru.timer = msys.StartMTimer(int64(mtlru.timeout), cleanTTimeout, mtlru)
	}
}
func (mtlru *MTLru) access(node *MTLruNode) {
	node.t = msys.NowMillisecond()
	mtlru.l.MoveToFront(node.le)
}
func (mtlru *MTLru) remove(node *MTLruNode) {
	if _, ok := mtlru.rbt.Get(node.Key); !ok {
		return
	}
	mtlru.rbt.Delete(node.Key)
	mtlru.l.Remove(node.le)
}
func (mtlru *MTLru) cleanTimeout() {
	nowTime := msys.NowMillisecond()
	for le := mtlru.l.Back(); le != nil; le = mtlru.l.Back() {
		node := le.Value.(*MTLruNode)
		if nowTime-node.t < mtlru.timeout {
			break
		}

		mtlru.remove(node)
	}
}
func (mtlru *MTLru) front() *MTLruNode {
	le := mtlru.l.Front()
	if le == nil {
		return nil
	}
	return le.Value.(*MTLruNode)
}
func (mtlru *MTLru) back() *MTLruNode {
	le := mtlru.l.Back()
	if le == nil {
		return nil
	}
	return le.Value.(*MTLruNode)
}
func (mtlru *MTLru) Insert(key rbt.RbKey, val interface{}) (err error) {
	mtlru.Lock()
	defer mtlru.Unlock()

	if n, ok := mtlru.rbt.Get(key); ok {
		node := n.(*MTLruNode)
		node.Val = val
		mtlru.access(node)
		return nil
	}

	node := &MTLruNode{}
	node.Key = key
	node.Val = val
	mtlru.rbt.Insert(key, node)
	node.le = mtlru.l.PushFront(node)
	mtlru.access(node)

	if mtlru.maxN > 0 && mtlru.l.Len() > mtlru.maxN {
		mtlru.remove(mtlru.back())
	}

	return nil
}
func (mtlru *MTLru) Remove(key rbt.RbKey) (v interface{}, err error) {
	mtlru.Lock()
	defer mtlru.Unlock()

	n, ok := mtlru.rbt.Get(key)
	if !ok {
		return nil, fmt.Errorf("not exist key %v", key)
	}
	node := n.(*MTLruNode)
	mtlru.remove(node)

	return node.Val, nil
}

func (mtlru *MTLru) GetSet(key rbt.RbKey, val interface{}) (nv interface{}, set bool) {
	mtlru.Lock()
	defer mtlru.Unlock()

	if n, ok := mtlru.rbt.Get(key); ok {
		node := n.(*MTLruNode)
		mtlru.access(node)
		return node.Val, false
	}

	node := &MTLruNode{}
	node.Key = key
	node.Val = val
	mtlru.rbt.Insert(key, node)
	node.le = mtlru.l.PushFront(node)
	mtlru.access(node)

	if mtlru.maxN > 0 && mtlru.l.Len() > mtlru.maxN {
		mtlru.remove(mtlru.back())
	}

	return val, true
}

func (mtlru *MTLru) Get(key rbt.RbKey) (v interface{}, exist bool) {
	mtlru.RLock()
	defer mtlru.RUnlock()

	if n, ok := mtlru.rbt.Get(key); ok {
		node := n.(*MTLruNode)
		mtlru.access(node)
		return node.Val, true
	}
	return nil, false
}
func (mtlru *MTLru) Gets(loKey rbt.RbKey, hiKey rbt.RbKey) (nodes []*MTLruNode, err error) {
	mtlru.RLock()
	defer mtlru.RUnlock()

	f := func(iterator rbt.RbIterator, key rbt.RbKey, value interface{}) {
		node := value.(*MTLruNode)
		mtlru.access(node)
		nodes = append(nodes, node)
	}
	it, err := mtlru.rbt.NewRbIterator(f)
	if err != nil {
		return nil, err
	}
	it.Between(loKey, hiKey)

	return nodes, nil
}

func (mtlru *MTLru) StopClean() {
	if mtlru.timer != nil {
		mtlru.timer.Stop()
		mtlru.timer = nil
	}
}

func NewMTLru(maxN int, timeout uint64) *MTLru {
	mtlru := &MTLru{maxN: maxN, timeout: timeout}
	mtlru.init()

	return mtlru
}
