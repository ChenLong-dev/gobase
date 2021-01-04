/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:45:32
 * @LastEditTime: 2020-12-16 14:45:32
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcontainer

import (
	"errors"
	"github.com/ocdogan/rbt"
	"sync"
	"sync/atomic"
	"unsafe"
)

/*
 协程安全的rbtree，在之前的rbtree上包装了锁而已。
 局部缓存还是用原始的rbt就好了
 另外这个没有办法保证业务节点的一致性
 这个看是不是在业务节点做锁还是怎么处理
*/

type Mrbtree struct {
	mu sync.RWMutex
	t  *rbt.RbTree
}

func MrbtreeNew() *Mrbtree {
	var t Mrbtree
	t.t = rbt.NewRbTree()
	return &t
}

func (t *Mrbtree) Get(key rbt.RbKey) (interface{}, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.t.Get(key)
}

func (t *Mrbtree) Insert(key rbt.RbKey, value interface{}) {
	t.mu.Lock()
	t.t.Insert(key, value)
	t.mu.Unlock()
}

func (t *Mrbtree) Upsert(key rbt.RbKey, value interface{}) (old interface{}, up bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	old, up = t.t.Get(key)
	t.t.Insert(key, value)
	return old, up
}

func (t *Mrbtree) Delete(key rbt.RbKey) (del interface{}, ok bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	del, ok = t.t.Get(key)
	if ok {
		t.t.Delete(key)
	}

	return del, ok
}

func (t *Mrbtree) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.t.Count()
}
func (t *Mrbtree) IsEmpty() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.t.IsEmpty()
}

func (t *Mrbtree) Exists(key rbt.RbKey) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.t.Exists(key)
}

// RbIterator interface used for iterating on a RbTree
type MrbtreeItor struct {
	// All iterates on all items of the RbTree
	t    *Mrbtree
	itor rbt.RbIterator
}

func (t *Mrbtree) NewRbIterator(cb rbt.RbIterationCallback) (*MrbtreeItor, error) {

	var it MrbtreeItor
	var err error
	it.t = t
	it.itor, err = t.t.NewRbIterator(cb)

	if err != nil {
		return nil, err
	}
	return &it, nil
}

func (it *MrbtreeItor) All() (int, error) {
	it.t.mu.RLock()
	defer it.t.mu.RUnlock()
	return it.itor.All()
}
func (it *MrbtreeItor) Between(lkey rbt.RbKey, rkey rbt.RbKey) (int, error) {
	it.t.mu.RLock()
	defer it.t.mu.RUnlock()
	return it.itor.Between(lkey, rkey)
}

func (it *MrbtreeItor) Close() {
	it.t = nil
	it.itor.Close()
}

/*
  这个是模拟sync.map 写的 空间换性能的，释放读锁
*/

type Mrbt struct {
	mu     sync.Mutex
	read   atomic.Value // readOnly
	dirty  *rbt.RbTree
	misses int
}

// readOnly is an immutable struct stored atomically in the Map.read field.
type readOnly struct {
	t       *rbt.RbTree
	amended bool // true的时候表示，dirty里面有数据，在t里面不存在
}

// expunged is an arbitrary pointer that marks entries which have been deleted
// from the dirty map.
var expunged = unsafe.Pointer(new(interface{}))

// An entry is a slot in the map corresponding to a particular key.
type entry struct {
	p unsafe.Pointer // *interface{}
}

func (e *entry) load() (value interface{}, ok bool) {
	p := atomic.LoadPointer(&e.p)
	if p == nil || p == expunged {
		return nil, false
	}
	return *(*interface{})(p), true
}

func (e *entry) tryStore(i *interface{}) bool {
	p := atomic.LoadPointer(&e.p)
	if p == expunged {
		return false
	}
	for {
		if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(i)) {
			return true
		}
		p = atomic.LoadPointer(&e.p)
		if p == expunged {
			return false
		}
	}
}

func (e *entry) unexpungeLocked() (wasExpunged bool) {
	return atomic.CompareAndSwapPointer(&e.p, expunged, nil)
}

func (e *entry) storeLocked(i *interface{}) {
	atomic.StorePointer(&e.p, unsafe.Pointer(i))
}

func (e *entry) delete() (hadValue bool) {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == nil || p == expunged {
			return false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, nil) {
			return true
		}
	}
}

func (e *entry) tryExpungeLocked() (isExpunged bool) {
	p := atomic.LoadPointer(&e.p)
	for p == nil {
		if atomic.CompareAndSwapPointer(&e.p, nil, expunged) {
			return true
		}
		p = atomic.LoadPointer(&e.p)
	}
	return p == expunged
}

func newEntry(i interface{}) *entry {
	return &entry{p: unsafe.Pointer(&i)}
}

func MrbtNew() *Mrbt {

	var t Mrbt
	/*俩个空指针*/
	t.dirty = rbt.NewRbTree()
	t.read.Store(readOnly{t: rbt.NewRbTree(), amended: false})

	return &t
}

/*

func localGet(t *rbt.RbTree, key rbt.RbKey) (interface{}, bool) {
	if t == nil {
		return  nil, false
	}
	return  t.Get(key)
}

func (t *Mrbt) localInsert(key rbt.RbKey, value interface{}) {
	if t.dirty = nil {
		t.dirty = rbt.NewRbTree()
	}
	t.dirty.Insert(key, value)
}
*/

func (t *Mrbt) missLocked() {
	t.misses++
	if t.misses < t.dirty.Count() {
		return
	}

	t.read.Store(readOnly{t: t.dirty, amended: false})
	t.dirty = rbt.NewRbTree()
	t.misses = 0
	return
}

func (t *Mrbt) dirtyLocked() {
	if t.dirty.Count() != 0 {
		return
	}
	read, _ := t.read.Load().(readOnly)
	//t.dirty = rbt.NewRbTree()

	iter_func := func(iterator rbt.RbIterator, key rbt.RbKey, value interface{}) {
		if !value.(*entry).tryExpungeLocked() { /* 非删除的节点*/
			t.dirty.Insert(key, value)
		}
	}

	iter, _ := read.t.NewRbIterator(iter_func)
	iter.All()
	iter.Close()
}

func (t *Mrbt) Get(key rbt.RbKey) (interface{}, bool) {
	read, _ := t.read.Load().(readOnly)
	e, ok := read.t.Get(key)
	if !ok && read.amended {
		t.mu.Lock()
		read, _ = t.read.Load().(readOnly)
		e, ok = read.t.Get(key)
		if !ok && read.amended {
			e, ok = t.dirty.Get(key)
			t.missLocked() /* 只是查询的话不上升，后续如果查询多，写入多，可以考虑每次查到不一样的都枷锁上升,节约read的时候lock*/
		}
		t.mu.Unlock()
	}
	if !ok {
		return nil, false
	}
	return e.(*entry).load()
}

func (t *Mrbt) Insert(key rbt.RbKey, value interface{}) {
	read, _ := t.read.Load().(readOnly)
	if e, ok := read.t.Get(key); ok && e.(*entry).tryStore(&value) {
		return
	}

	t.mu.Lock()
	read, _ = t.read.Load().(readOnly)
	if e, ok := read.t.Get(key); ok {
		if e.(*entry).unexpungeLocked() {
			t.dirty.Insert(key, e)
		}
		e.(*entry).storeLocked(&value)
	} else if e, ok := t.dirty.Get(key); ok {
		e.(*entry).storeLocked(&value)
	} else {
		if !read.amended {
			t.dirtyLocked()
			t.read.Store(readOnly{t: read.t, amended: true}) //如果后期是count或者itor这种需要重提升的比较多，可以考虑直接先inert，在false
		}
		t.dirty.Insert(key, newEntry(value))
	}
	t.mu.Unlock()
}

func (t *Mrbt) Delete(key rbt.RbKey) {
	read, _ := t.read.Load().(readOnly)
	e, ok := read.t.Get(key)
	if !ok && read.amended {
		t.mu.Lock()
		read, _ = t.read.Load().(readOnly)
		e, ok = read.t.Get(key)
		if !ok && read.amended {
			t.dirty.Delete(key)
		}
		t.mu.Unlock()
	}
	if ok {
		e.(*entry).delete()
	}
}

func (t *Mrbt) Count() int {
	/*count不能直接count了，蛋疼,不过好在count用的也不会太多*/

	read, _ := t.read.Load().(readOnly)
	if read.amended {
		/*直接加载到read里面*/
		t.mu.Lock()
		read, _ = t.read.Load().(readOnly)
		if read.amended {
			read = readOnly{t: t.dirty, amended: false}
			t.read.Store(read)
			t.dirty = rbt.NewRbTree()
			t.misses = 0
		}
		t.mu.Unlock()
	}
	count := 0
	iter_func := func(iterator rbt.RbIterator, key rbt.RbKey, value interface{}) {
		_, ok := value.(*entry).load()
		if ok {
			count++
		}
	}
	/* 这个时候遍历没有问题，因为就算有其他地方read转换，也是另外一个read和hash，只是节点还是同一个节点而已*/
	iter, _ := read.t.NewRbIterator(iter_func)
	iter.All()
	iter.Close()
	return count
}

func (t *Mrbt) IsEmpty() bool {
	if t.Count() == 0 {
		return true
	}
	return false
}

func (t *Mrbt) Exists(key rbt.RbKey) bool {
	_, ok := t.Get(key)
	return ok
}

// RbIterator interface used for iterating on a RbTree
type MrbtItor struct {
	// All iterates on all items of the RbTree
	t  *Mrbt
	cb rbt.RbIterationCallback
	// 遍历的时候再去生成？
	// itor rbt.RbIterator
}

func (t *Mrbt) NewRbIterator(cb rbt.RbIterationCallback) (*MrbtItor, error) {

	if t == nil || cb == nil {
		return nil, errors.New("Param error!")
	}

	var it MrbtItor
	//var err error
	it.t = t
	it.cb = cb
	//it.itor = nil

	return &it, nil

}

func (it *MrbtItor) All() (int, error) {
	t := it.t
	read, _ := t.read.Load().(readOnly)
	if read.amended {
		t.mu.Lock()
		read, _ := t.read.Load().(readOnly)
		if read.amended {
			read = readOnly{t: t.dirty, amended: false}
			t.read.Store(read)
			t.dirty = rbt.NewRbTree()
			t.misses = 0
		}
		t.mu.Unlock()
	}
	iter_func := func(iterator rbt.RbIterator, key rbt.RbKey, value interface{}) {
		v, ok := value.(*entry).load()
		if ok {
			it.cb(iterator, key, v)
		}
	}
	/* 这个时候遍历没有问题，因为就算有其他地方read转换，也是另外一个read和hash，只是节点还是同一个节点而已*/
	iter, _ := read.t.NewRbIterator(iter_func)
	iter.All()
	iter.Close()
	return 0, nil
}

func (it *MrbtItor) Between(lkey rbt.RbKey, rkey rbt.RbKey) (int, error) {
	t := it.t
	read, _ := t.read.Load().(readOnly)
	if read.amended {
		t.mu.Lock()
		read, _ := t.read.Load().(readOnly)
		if read.amended {
			read = readOnly{t: t.dirty, amended: false}
			t.read.Store(read)
			t.dirty = rbt.NewRbTree()
			t.misses = 0
		}
		t.mu.Unlock()
	}
	iter_func := func(iterator rbt.RbIterator, key rbt.RbKey, value interface{}) {
		v, ok := value.(*entry).load()
		if ok {
			it.cb(iterator, key, v)
		}
	}
	/*每次新生成一个iter去遍历*/
	iter, _ := read.t.NewRbIterator(iter_func)
	iter.Between(lkey, rkey)
	iter.Close()
	return 0, nil
}

func (it *MrbtItor) Close() {
	it.t = nil
	it.cb = nil
}

type RbtStringKey struct {
	k string
}

func (key RbtStringKey) ComparedTo(rk rbt.RbKey) rbt.KeyComparison {
	rkey := rk.(RbtStringKey)
	if key.k < rkey.k {
		return rbt.KeyIsLess
	} else if key.k > rkey.k {
		return rbt.KeyIsGreater
	} else {
		return rbt.KeysAreEqual
	}
}
func NewRbtSstringKey(k string) RbtStringKey {
	return RbtStringKey{k: k}
}
func NewRbtStringKeyHi(k string) RbtStringKey {
	bs := []byte(k)
	bs = append(bs, byte(0xff))
	return RbtStringKey{k: string(bs)}
}
