/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-22 20:55:28
 * @LastEditTime: 2020-12-22 20:57:19
 * @LastEditors: Chen Long
 * @Reference:
 */

//一致性哈希

package chash

import (
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
)

type ANode struct {
	Key      string
	Data     interface{}
	Weight   uint
	HitCount uint
}

func (an *ANode) String() string {
	return fmt.Sprintf(`{"Key":%s,"Weight":%d,"HitCount":%d}`, an.Key, an.Weight, an.HitCount)
}

type vNode struct {
	actNode  *ANode
	k        uint32
	hitCount uint32
}

func (vn *vNode) String() string {
	return fmt.Sprintf(`{"k":%d,"hitCount":%d}`, vn.k, vn.hitCount)
}

type CHash struct {
	sync.RWMutex

	vNodesFactor uint
	aNodes       map[string]*ANode
	vNodes       []*vNode
}

func (c *CHash) String() string {
	c.RLock()
	defer c.RUnlock()

	ans := "\t{\n"
	for _, an := range c.aNodes {
		ans += "\t\t" + an.String() + ",\n"
	}
	ans += "\t}"

	vns := "\t{\n"
	for _, vn := range c.vNodes {
		vns += "\t\t" + vn.String() + ",\n"
	}
	vns += "\t}"

	return fmt.Sprintf("{\n\tvNodesFactor:%d,\n\taNodes:\n%s,\nvNodes:\n%s\n}", c.vNodesFactor, ans, vns)
}

func NewCHash(vNodesFactor uint) *CHash {
	if vNodesFactor == 0 {
		vNodesFactor = 63
	}
	return &CHash{
		vNodesFactor: vNodesFactor,
		aNodes:       make(map[string]*ANode),
		vNodes:       []*vNode{},
	}
}

func (c *CHash) Add(nk string, nd interface{}, nw uint) (oldData interface{}) {
	c.Lock()
	defer c.Unlock()

	if nw == 0 {
		nw = 1
	}

	n := &ANode{
		Key:    nk,
		Data:   nd,
		Weight: nw,
	}

	if an, ok := c.aNodes[nk]; ok {
		oldData = an.Data
		if an.Weight == nw {
			an.Data = nd
			return oldData
		}

		c.remove(nk)
	}
	c.aNodes[nk] = n

	count := int(c.vNodesFactor * nw)
	for i := 0; i < count; i++ {
		vn := &vNode{}
		vn.actNode = n
		vn.k = c.hashStr(fmt.Sprintf("%s#%d", nk, i))
		c.vNodes = append(c.vNodes, vn)
	}
	c.sortHashRing()
	return
}
func (c *CHash) remove(key string) (del interface{}, exist bool) {
	an, ok := c.aNodes[key]
	if !ok {
		return nil, false
	}
	del = an.Data
	exist = true

	vNodes := make([]*vNode, 0, len(c.vNodes))
	for _, vn := range c.vNodes {
		if vn.actNode.Key != key {
			vNodes = append(vNodes, vn)
		}
	}
	c.vNodes = vNodes
	delete(c.aNodes, key)
	c.sortHashRing()

	return del, exist
}
func (c *CHash) Remove(key string) (del interface{}, exist bool) {
	c.Lock()
	defer c.Unlock()

	return c.remove(key)
}

func (c *CHash) sortHashRing() {
	sort.Slice(c.vNodes, func(i, j int) bool {
		return c.vNodes[i].k < c.vNodes[j].k
	})
}

func (c *CHash) Hit(key string) interface{} {
	hash := c.hashStr(key)

	c.RLock()
	defer c.RUnlock()

	if len(c.vNodes) == 0 {
		return nil
	}
	n := len(c.vNodes)
	i := sort.Search(n, func(i int) bool { return c.vNodes[i].k >= hash })
	if i == n {
		i = 0
	}
	an := c.vNodes[i].actNode
	an.HitCount++
	return an.Data
}
func (c *CHash) HitFunc(key string, f func(nd interface{}) bool) interface{} {
	hash := c.hashStr(key)

	c.RLock()
	defer c.RUnlock()

	if len(c.vNodes) == 0 {
		return nil
	}
	vLen := len(c.vNodes)
	pos := sort.Search(vLen, func(i int) bool { return c.vNodes[i].k >= hash })
	if pos == vLen {
		pos = 0
	}

	for i := 0; i < vLen; i++ {
		an := c.vNodes[pos].actNode
		if f(an.Data) {
			an.HitCount++
			return an.Data
		}
		pos = (pos + 1) % vLen
	}

	return nil
}
func (c *CHash) GetData(key string) (data interface{}, ok bool) {
	c.RLock()
	defer c.RUnlock()

	if an, ok := c.aNodes[key]; ok {
		return an.Data, ok
	}
	return nil, false
}

func (c *CHash) All() (ans []*ANode) {
	c.RLock()
	defer c.RUnlock()

	for _, an := range c.aNodes {
		ans = append(ans, an)
	}
	return ans
}

func (c *CHash) hashStr(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}
