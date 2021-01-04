/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 11:55:55
 * @LastEditTime: 2020-12-16 11:56:22
 * @LastEditors: Chen Long
 * @Reference:
 */

package etcd

import (
	"github.com/ChenLong-dev/gobase/mlog"
	"strings"
	"sync"
)

type EtcdDis struct {
	Cluster string
	// 监听相关的服务
	MapWatch map[string]*EtcdMaster
	watchMu  *sync.RWMutex
}

func NewEtcdDis(cluster string) *EtcdDis {
	return &EtcdDis{
		Cluster:  cluster,
		MapWatch: make(map[string]*EtcdMaster),
		watchMu:  &sync.RWMutex{},
	}
}

func (d *EtcdDis) Watch(service string) error {
	w := newMaster(d.Cluster, service)

	if d.MapWatch == nil {
		d.MapWatch = make(map[string]*EtcdMaster)
	}

	if _, ok := d.getEtcdMaster(service); ok {
		mlog.Infof("Service:%s Have Watch!\n", service)
		return nil
	}

	mlog.Debugf("w is %v", w.nodesMap)
	d.addEtcdMaster(service, w)

	return nil
}

// 获取服务的节点信息-随机获取
func (d *EtcdDis) GetServiceInfoRandom(service string) (node EtcdNode, has bool) {
	if len(d.MapWatch) == 0 {
		mlog.Warn("MapWatch is empty")
		return
	}

	if v, ok := d.getEtcdMaster(service); ok {
		mlog.Debugf("V is %v, ok is %v\n", v, ok)
		if v != nil {
			if n, ok1 := v.GetNodeRandom(); ok1 {
				mlog.Debugf("n is %v, ok1 is %v\n", v, ok1)
				return n, true
			}
			mlog.Debugf("n is %v\n", v)
		}
	} else {
		mlog.Errorf("Service:%s Not Be Watched!\n", service)
	}

	return
}

// 获取服务的节点信息-全部获取
func (d *EtcdDis) GetServiceInfoAllNode(service string) (nodes []*EtcdNode, has bool) {
	if len(d.MapWatch) == 0 {
		mlog.Warn("MapWatch is empty")
		return
	}

	if v, ok := d.getEtcdMaster(service); ok {
		if v != nil {
			return v.GetAllNodes(), true
		}
	} else {
		mlog.Errorf("Service:%s Not Be Watched!\n", service)
	}

	return
}

//直接通过ETCD实时获取服务配置信息，主要用于初始化时异步服务发现不够快的问题
func (d *EtcdDis) GetNodesByETCD(service string) (nodes []*EtcdNode, has bool) {
	if v, ok := d.getEtcdMaster(service); ok {
		nodes = v.allNodeByETCD()
		if len(nodes) > 0 {
			has = true
		}
	}
	return
}

//获取单个grpc路径
func (d *EtcdDis) GetGrpcNodeByPath(service string, key string) (EtcdNode, bool) {
	etcdMst, ok := d.getEtcdMaster(service)
	mlog.Debugf("Found %s: %s", service, etcdMst)
	if !ok {
		return EtcdNode{}, false
	}
	node, ok := etcdMst.getNode(key)
	if !ok {
		return EtcdNode{}, ok
	}
	return node, true
}

func (d *EtcdDis) addEtcdMaster(service string, w *EtcdMaster) {
	d.watchMu.Lock()
	defer d.watchMu.Unlock()
	d.MapWatch[service] = w
}

func (d *EtcdDis) getEtcdMaster(service string) (w *EtcdMaster, has bool) {
	d.watchMu.RLock()
	defer d.watchMu.RUnlock()
	if w, has = d.MapWatch[service]; has {
		return
	}
	return
}

// 拆分service name、key；返回bool true表示成功;false表示失败
func SplitServiceNameKey(dir string) (string, string, bool) {
	if idx := strings.Index(dir, "/"); -1 != idx {
		name := dir[:idx]
		key := dir[idx+1:]
		return name, key, true
	}
	return "", "", false
}
