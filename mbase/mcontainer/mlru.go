/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:45:03
 * @LastEditTime: 2020-12-16 14:45:03
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcontainer

import (
	"container/list"
	"fmt"
	"github.com/ChenLong-dev/gobase/mbase/msys"
	"sync"
)

type mlruNode struct {
	key interface{}
	val interface{}
	le  *list.Element
	t   uint64
}
type MLru struct {
	sync.Mutex
	l       *list.List
	m       map[interface{}]*mlruNode
	maxN    int
	timeout uint64
	timer   *msys.MTimer
}

func cleanTimeout(pridata interface{}) {
	mlru := pridata.(*MLru)
	mlru.Lock()
	defer mlru.Unlock()

	mlru.cleanTimeout()

	if mlru.timer != nil {
		mlru.timer = msys.StartMTimer(int64(mlru.timeout), cleanTimeout, mlru)
	}
}
func (mlru *MLru) init() {
	mlru.l = list.New()
	mlru.m = make(map[interface{}]*mlruNode)
	if mlru.timeout > 0 {
		mlru.timer = msys.StartMTimer(int64(mlru.timeout), cleanTimeout, mlru)
	}
}
func (mlru *MLru) access(node *mlruNode) {
	node.t = msys.NowMillisecond()
	mlru.l.MoveToFront(node.le)
}
func (mlru *MLru) remove(node *mlruNode) {
	if _, ok := mlru.m[node.key]; !ok {
		return
	}
	delete(mlru.m, node.key)
	mlru.l.Remove(node.le)
}
func (mlru *MLru) cleanTimeout() {
	nowTime := msys.NowMillisecond()
	for le := mlru.l.Back(); le != nil; le = mlru.l.Back() {
		node := le.Value.(*mlruNode)
		if nowTime-node.t < mlru.timeout {
			break
		}

		mlru.remove(node)
	}
}
func (mlru *MLru) front() *mlruNode {
	le := mlru.l.Front()
	if le == nil {
		return nil
	}
	return le.Value.(*mlruNode)
}
func (mlru *MLru) back() *mlruNode {
	le := mlru.l.Back()
	if le == nil {
		return nil
	}
	return le.Value.(*mlruNode)
}
func (mlru *MLru) Insert(key interface{}, val interface{}) (err error) {
	mlru.Lock()
	defer mlru.Unlock()

	if node, ok := mlru.m[key]; ok {
		node.val = val
		mlru.access(node)
		return nil
	}

	node := &mlruNode{}
	node.key = key
	node.val = val
	mlru.m[key] = node
	node.le = mlru.l.PushFront(node)
	mlru.access(node)

	if mlru.maxN > 0 && mlru.l.Len() > mlru.maxN {
		mlru.remove(mlru.back())
	}

	return nil
}
func (mlru *MLru) Remove(key interface{}) (v interface{}, err error) {
	mlru.Lock()
	defer mlru.Unlock()

	node, ok := mlru.m[key]
	if !ok {
		return nil, fmt.Errorf("not exist key %v", key)
	}
	mlru.remove(node)

	return node.val, nil
}
func (mlru *MLru) GetSet(key interface{}, val interface{}) (nv interface{}, set bool) {
	mlru.Lock()
	defer mlru.Unlock()

	if node, ok := mlru.m[key]; ok {
		mlru.access(node)
		return node.val, false
	}

	node := &mlruNode{}
	node.key = key
	node.val = val
	mlru.m[key] = node
	node.le = mlru.l.PushFront(node)
	mlru.access(node)

	if mlru.maxN > 0 && mlru.l.Len() > mlru.maxN {
		mlru.remove(mlru.back())
	}

	return val, true
}

func (mlru *MLru) Get(key interface{}) (v interface{}, exist bool) {
	mlru.Lock()
	defer mlru.Unlock()

	if node, ok := mlru.m[key]; ok {
		mlru.access(node)
		return node.val, true
	}
	return nil, false
}
func (mlru *MLru) StopClean() {
	if mlru.timer != nil {
		mlru.timer.Stop()
		mlru.timer = nil
	}
}

func NewMLru(maxN int, timeout uint64) *MLru {
	mlru := &MLru{maxN: maxN, timeout: timeout}
	mlru.init()

	return mlru
}
