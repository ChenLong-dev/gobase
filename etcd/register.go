/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 12:00:30
 * @LastEditTime: 2020-12-16 12:00:38
 * @LastEditors: Chen Long
 * @Reference:
 */

package etcd

import (
	"sync"
	"github.com/ChenLong-dev/gobase/mlog"
)

type EtcdRegister struct {
	Cluster     string
	registerMap map[string]*EtcdService
	rwMu        *sync.RWMutex
}

func NewEtcdRegister(cluster string) *EtcdRegister {
	return &EtcdRegister{
		Cluster:     cluster,
		registerMap: make(map[string]*EtcdService),
		rwMu:        &sync.RWMutex{},
	}
}

// 注册
func (reg *EtcdRegister) Register(service, key, info string) error {
	var s *EtcdService
	var err error
	name := service + "/" + key
	if s, err = newService(reg.Cluster, name, info); err != nil {
		mlog.Errorf("Register service:%s error:%v", service, err.Error())
		return err
	}

	if reg.registerMap == nil {
		reg.registerMap = make(map[string]*EtcdService)
	}

	if _, ok := reg.getEtcdService(name); ok {
		mlog.Errorf("Service:%s Have Registered!", name)
		return nil
	}

	reg.addEtcdService(name, s)
	// w维持心跳
	s.Start()

	return nil
}

// 更新
func (reg *EtcdRegister) UpdateInfo(service, key, info string) {
	name := service + "/" + key
	if s, ok := reg.getEtcdService(name); ok && s != nil {
		s.setValue(info)
	}
}

func (reg *EtcdRegister) addEtcdService(name string, srv *EtcdService) {
	reg.rwMu.Lock()
	defer reg.rwMu.Unlock()
	reg.registerMap[name] = srv
}

func (reg *EtcdRegister) getEtcdService(name string) (srv *EtcdService, has bool) {
	reg.rwMu.RLock()
	defer reg.rwMu.RUnlock()
	if srv, has = reg.registerMap[name]; has {
		return
	}
	return
}
