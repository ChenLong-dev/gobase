/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:44:13
 * @LastEditTime: 2020-12-16 14:44:13
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcontainer

import "sync"

//	做个加锁版本，sync.Map无法统计数量

type Map struct {
	rwmu sync.RWMutex
	mds  map[interface{}]interface{}
}

func NewMap() *Map {
	m := new(Map)
	m.mds = make(map[interface{}]interface{})
	return m
}

// Store sets the value for a key.
func (m *Map) Store(key, value interface{}) {
	m.rwmu.Lock()
	m.mds[key] = value
	m.rwmu.Unlock()
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
	m.rwmu.RLock()
	v, ok := m.mds[key]
	m.rwmu.RUnlock()
	return v, ok
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *Map) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()
	actual, loaded = m.mds[key]
	if loaded {
		return
	}
	m.mds[key] = value
	return value, false
}

// Delete deletes the value for a key.
func (m *Map) Delete(key interface{}) (del interface{}, ok bool) {
	m.rwmu.Lock()
	del, ok = m.mds[key]
	delete(m.mds, key)
	m.rwmu.Unlock()
	return del, ok
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (m *Map) Range(f func(key, value interface{}) bool) {
	type kv struct {
		k interface{}
		v interface{}
	}

	skv := make([]*kv, 0, m.Size())
	m.rwmu.RLock()
	for k, v := range m.mds {
		skv = append(skv, &kv{k: k, v: v})
	}
	m.rwmu.RUnlock()
	for _, pkv := range skv {
		if !f(pkv.k, pkv.v) {
			break
		}
	}
}

func (m *Map) Size() int {
	m.rwmu.RLock()
	size := len(m.mds)
	m.rwmu.RUnlock()
	return size
}
