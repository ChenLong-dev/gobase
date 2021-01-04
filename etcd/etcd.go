/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 11:56:43
 * @LastEditTime: 2020-12-16 11:56:46
 * @LastEditors: Chen Long
 * @Reference:
 */

//包级别的Put、Get、Delete、Close、Lock方法，需要首先显示的调用NewClient或NewTLSClient进行初始化
package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/ChenLong-dev/gobase/config"
	"github.com/ChenLong-dev/gobase/mlog"
	"go.etcd.io/etcd/clientv3"
	"io/ioutil"
	"time"
)

var (
	client *clientv3.Client
	//kv      clientv3.KV    //KV客户端
	//lease   clientv3.Lease //租约客户端
	timeout = 5 * time.Second
)

func Connect() (*clientv3.Client, error) {
	if config.Conf.Etcd.EtcdCa != "" {
		if err := NewTLSClient(); err != nil {
			mlog.Errorf("ETCD CONNECT FAIL: %s \n", err)
			return nil, err
		}
	} else {
		if err := NewClient(); err != nil {
			mlog.Errorf("ETCD CONNECT FAIL: %s \n", err)
			return nil, err
		}
	}
	mlog.Infof("[ETCD] CONNECT SUCCESSFUL \n")
	return client, nil
}

func NewClient() error {
	conf := clientv3.Config{
		Endpoints:   config.Conf.Etcd.Hosts,
		DialTimeout: timeout,
		Username:    config.Conf.Etcd.UserName,
		Password:    config.Conf.Etcd.Password,
	}

	var err error
	if client, err = clientv3.New(conf); err != nil {
		return err
	}
	mlog.Infof("etcd NewClient success")
	return nil
}

func NewTLSClient() error {
	mlog.Infof("etcd conf info:%v", config.Conf.Etcd)
	cert, err := tls.LoadX509KeyPair(config.Conf.Etcd.EtcdCert, config.Conf.Etcd.EtcdCertKey)
	if err != nil {
		mlog.Error(err)
		return err
	}

	caData, e := ioutil.ReadFile(config.Conf.Etcd.EtcdCa)
	if e != nil {
		return e
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caData)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}
	conf := clientv3.Config{
		Endpoints:   config.Conf.Etcd.Hosts,
		DialTimeout: timeout,
		Username:    config.Conf.Etcd.UserName,
		Password:    config.Conf.Etcd.Password,
		TLS:         tlsConfig,
	}

	if client, err = clientv3.New(conf); err != nil {
		return err
	}

	mlog.Infof("etcd NewTLSClient success")
	return nil
}

func Close() {
	client.Close()
}

func Put(key, val string, opts ...clientv3.OpOption) error {
	if _, err := client.KV.Put(context.TODO(), key, val, opts...); err != nil {
		return err
	}
	return nil
}

func Get(key string, opts ...clientv3.OpOption) ([]byte, error) {
	var getResp *clientv3.GetResponse
	var err error

	if getResp, err = client.KV.Get(context.TODO(), key, opts...); err != nil {
		return nil, err
	}

	for _, resp := range getResp.Kvs {
		mlog.Infof("getResp.Kvs key:%s\n", string(resp.Key))
		if string(resp.Key) == key {
			return resp.Value, nil
		}
	}

	return nil, nil
}

func GetAll(key string, opts ...clientv3.OpOption) (rsts [][]byte, err error) {
	var getResp *clientv3.GetResponse
	if getResp, err = client.KV.Get(context.TODO(), key, opts...); err != nil {
		return nil, err
	}
	for _, resp := range getResp.Kvs {
		//mlog.Debugf("getResp.Kvs key:%s, value:%s\n", string(resp.Key), string(resp.Value))
		rsts = append(rsts, resp.Value)
	}
	return
}

func Delete(key string, opts ...clientv3.OpOption) error {
	if _, err := client.KV.Delete(context.TODO(), key, opts...); err != nil {
		return err
	}

	return nil
}

//ETCD的分布式锁，封装，主要在租约机制上增加了续约goroutinue
//ttl租约过期时间
func Lock(lockKey string, ttl int64) (succeeded bool, err error) {

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*30)
	defer cancel()
	//新建一个租约，过期时间30秒
	leaseGrantResponse, err := client.Lease.Grant(ctx, ttl)
	if err != nil {
		fmt.Println(err)
		return
	}
	leaseId := leaseGrantResponse.ID

	//租约自动过期，立刻过期。//cancelfunc 取消续租，而revoke 则是立即过期
	// ctx, cancelFunc = context.WithCancel(context.TODO())
	// defer cancelFunc()
	// defer lease.Revoke(context.TODO(), leaseId)

	leaseKeepAliveChan, err := client.Lease.KeepAlive(context.TODO(), leaseId)
	if err != nil {
		fmt.Println(err)
		return
	}
	//启动取出续租结果Goroutinue，每ttl/3秒会自动续约
	go func() {
		defer client.Revoke(context.TODO(), leaseId)
		for {
			select {
			case leaseKeepAliveResponse := <-leaseKeepAliveChan:
				//println(time.Now().Format("2006 01 02 15:04:05.000"))
				if leaseKeepAliveResponse != nil {
					mlog.Trace("续租成功,leaseID :", leaseKeepAliveResponse.ID)
				} else {
					mlog.Warn("续租失败!!!!!!")
					return
				}
			}
			//time.Sleep(time.Second * 5)
		}
	}()
	//锁逻辑。
	txn := client.KV.Txn(context.TODO())

	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).Then(
		clientv3.OpPut(lockKey, "占用", clientv3.WithLease(leaseId))).Else(
		clientv3.OpGet(lockKey))
	txnResponse, err := txn.Commit()
	if err != nil {
		fmt.Println(err)
		return
	}
	if txnResponse.Succeeded {
		fmt.Println("抢到锁了")
	} else {
		mlog.Tracef("没抢到锁,%v", txnResponse.Responses[0].GetResponseRange().Kvs[0].Value)
	}
	succeeded = txnResponse.Succeeded
	return
}
