/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:06:32
 * @LastEditTime: 2020-12-17 09:06:32
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
type Group struct {
	mu sync.Mutex // protects m
	m  sync.Map
	//m  map[string]*call // lazily initialized
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	if vCall, ok := g.m.Load(key); ok {
		c := vCall.(*call)
		c.wg.Wait()
		return c.val, c.err
	}

	g.mu.Lock()
	if vCall, ok := g.m.Load(key); ok {
		g.mu.Unlock()
		c := vCall.(*call)
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m.Store(key, c)
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.m.Delete(key)

	return c.val, c.err
}

var stdGroup = &Group{}

func SingleFlight(key string, fn func() (interface{}, error)) (interface{}, error) {
	return stdGroup.Do(key, fn)
}
