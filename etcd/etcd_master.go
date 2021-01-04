/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 11:57:56
 * @LastEditTime: 2020-12-16 11:58:23
 * @LastEditors: Chen Long
 * @Reference:
 */

package etcd

import (
	"context"
	"github.com/ChenLong-dev/gobase/mlog"
	"go.etcd.io/etcd/clientv3"
	"math/rand"
	"strings"
	"sync"
)

// Etcd注册的节点，一个节点代表一个client
type EtcdNode struct {
	State   bool
	Cluster string // 集群
	Key     string // key
	Info    []byte // 节点信息
}

type EtcdMaster struct {
	Cluster  string // 集群
	Path     string // 路径
	nodesMap map[string]*EtcdNode
	Client   *clientv3.Client
	nodesMu  *sync.RWMutex
}

func newMaster(cluster string, watchPath string) *EtcdMaster {
	if _, err := Connect(); err != nil {
		panic(err)
	}
	master := &EtcdMaster{
		Cluster:  cluster,
		Path:     watchPath,
		nodesMap: make(map[string]*EtcdNode),
		Client:   client,
		nodesMu:  &sync.RWMutex{},
	}
	// 监听观察节点
	go master.WatchNodes()

	mlog.Debugf("master is %v; master.nodes is %v", master, master.nodesMap)
	return master
}

func newEtcdNode(ev *clientv3.Event) string {
	return string(ev.Kv.Value)
}

func (m *EtcdMaster) allNodeByETCD() (nodes []*EtcdNode) {
	resp, err := m.Client.Get(context.Background(), "/"+m.Cluster+"/"+m.Path, clientv3.WithPrefix())
	if nil != err {
		mlog.Error(err)
		return
	}
	for _, ev := range resp.Kvs {
		node := &EtcdNode{
			State:   true,
			Cluster: m.Cluster,
			Key:     string(ev.Key),
			Info:    ev.Value,
		}
		nodes = append(nodes, node)
	}
	return
}

// 监听观察节点
func (m *EtcdMaster) WatchNodes() {
	// 查看之前存在的节点
	mlog.Tracef("grpc's etcd addr:%v", "/"+m.Cluster+"/"+m.Path)
	resp, err := m.Client.Get(context.Background(), "/"+m.Cluster+"/"+m.Path, clientv3.WithPrefix())
	mlog.Tracef("etcd's responese to grpc:%v", resp)
	if nil != err {
		mlog.Error(err)
	} else {
		//mlog.Debugf("resp.Kvs is :%v", resp.Kvs)
		for _, ev := range resp.Kvs {
			mlog.Infof("add dir:%q, value:%q\n", ev.Key, ev.Value)
			m.addNode(string(ev.Key), ev.Value)
		}
	}

	rch := m.Client.Watch(context.Background(), "/"+m.Cluster+"/"+m.Path, clientv3.WithPrefix(), clientv3.WithPrevKV())
	mlog.Debugf("etcd's rch to grpc:%T %v", rch, rch)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			mlog.Debugf("etcd's ev to grpc:%T %v", ev, ev)
			switch ev.Type {
			case clientv3.EventTypePut:
				mlog.Infof("[%s] dir:%q, value:%q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				m.addNode(string(ev.Kv.Key), ev.Kv.Value)
			case clientv3.EventTypeDelete:
				mlog.Infof("[%s] dir:%q, value:%q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				k := ev.Kv.Key
				if len(ev.Kv.Key) > (len(m.Cluster) + 1) {
					k = ev.Kv.Key[len(m.Cluster)+1:]
				}
				m.delNode(string(k))
			default:
				mlog.Infof("[%s] dir:%q, value:%q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}
		}
	}
	//mlog.Debugf("m.nodes are %v", m.nodesMap)
}

// 获取该集群下所有的节点
func (m *EtcdMaster) GetAllNodes() (allNodes []*EtcdNode) {
	m.nodesMu.RLock()
	defer m.nodesMu.RUnlock()
	for i := range m.nodesMap {
		if m.nodesMap[i] != nil {
			allNodes = append(allNodes, m.nodesMap[i])
		}
	}
	return
}

func (m *EtcdMaster) GetNodeRandom() (EtcdNode, bool) {
	count := len(m.nodesMap)
	mlog.Debugf("count is %v:\n", count)
	// 该集群不存在节点时，直接返回false
	if count == 0 {
		return EtcdNode{}, false
	}
	m.nodesMu.RLock()
	defer m.nodesMu.RUnlock()

	idx := rand.Intn(count)
	for _, v := range m.nodesMap {
		if idx == 0 {
			return *v, true
		}
		idx = idx - 1
	}
	return EtcdNode{}, false
}

// 获取节点
func (m *EtcdMaster) getNode(key string) (EtcdNode, bool) {
	node, ok := m.nodesMap[key]
	if !ok {
		return EtcdNode{}, false
	}
	return *node, true
}

// 添加节点
func (m *EtcdMaster) addNode(key string, info []byte) {
	k := key
	if len(key) > (len(m.Cluster) + 1) {
		k = key[len(m.Cluster)+1:]
	}
	if strings.Contains(k, "/grpc") {
		k = strings.Split(k, "/grpc")[0]
	}
	node := &EtcdNode{
		State:   true,
		Cluster: m.Cluster,
		Key:     k,
		Info:    info,
	}
	mlog.Infof("[ETCD DISCOVERY] map add node key: %s, value: %s", k, info)
	m.nodesMu.Lock()
	defer m.nodesMu.Unlock()
	m.nodesMap[node.Key] = node
}

func (m *EtcdMaster) delNode(key string) {
	m.nodesMu.Lock()
	defer m.nodesMu.Unlock()
	delete(m.nodesMap, key)
}
