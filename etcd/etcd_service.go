/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 11:59:03
 * @LastEditTime: 2020-12-16 11:59:17
 * @LastEditors: Chen Long
 * @Reference:
 */

package etcd

import (
	"context"
	"encoding/json"
	"github.com/ChenLong-dev/gobase/mlog"
	"go.etcd.io/etcd/clientv3"
)

type EtcdService struct {
	Cluster string // 集群名称
	Name    string // 服务名称
	Info    string // 节点信息
	stop    chan error
	leaseid clientv3.LeaseID
	client  *clientv3.Client
}

// 创建ETCD服务
func newService(cluster, name, info string) (*EtcdService, error) {
	cli, err := Connect()
	if err != nil {
		mlog.Error(err)
		return nil, err
	}
	// 返回服务对象
	return &EtcdService{
		Cluster: cluster,
		Name:    name,
		Info:    info,
		stop:    make(chan error),
		client:  cli,
	}, err
}

// 启动
func (s *EtcdService) Start() error {
	// 获取心跳的通道
	ch, err := s.keepLive()
	if err != nil {
		mlog.Error(err)
		return err
	}
	go func() {
		// 死循环
		for {
			select {
			case <-s.stop:
				s.revoke()
				return
			case <-s.client.Ctx().Done():
				mlog.Debugf("server closed")
				return
			case ka, ok := <-ch:
				if !ok {
					mlog.Debugf("keep live channel closed")
					s.revoke()
					return
				} else {
					mlog.Tracef("recv reply frem service:[%s], ttl:[%d],", s.Name, ka.TTL)
				}
			}
		}
	}()
	return nil
}

// 停止
func (s *EtcdService) Stop() {
	s.stop <- nil
}

// 保持心跳
func (s *EtcdService) keepLive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	key := "/" + s.Cluster + "/" + s.Name
	value, _ := json.Marshal(s.Info)

	// minimum lease TTL is 60-second
	resp, err := s.client.Grant(context.TODO(), 10)
	if err != nil {
		mlog.Error(err)
		return nil, err
	}

	mlog.Infof("Register, Key:%s, Value:%s\n", key, string(value))
	_, err = s.client.Put(context.TODO(), key, string(value), clientv3.WithLease(resp.ID))
	if nil != err {
		mlog.Error(err)
		return nil, err
	}

	s.leaseid = resp.ID
	return s.client.KeepAlive(context.TODO(), resp.ID)
}

// 设置节点信息
func (s *EtcdService) setValue(info string) error {
	s.Info = info
	tmp, _ := json.Marshal(info)
	key := "/" + s.Cluster + "/" + s.Name
	if _, err := s.client.Put(context.TODO(), key, string(tmp), clientv3.WithLease(s.leaseid)); err != nil {
		mlog.Errorf("etcd set value failed! key:%s;value:%s", key, info)
		return err
	}

	return nil
}

// 撤销
func (s *EtcdService) revoke() error {
	_, err := s.client.Revoke(context.TODO(), s.leaseid)
	if err != nil {
		mlog.Error(err)
	}

	mlog.Infof("service:%s stop\n", s.Name)
	return nil
}
