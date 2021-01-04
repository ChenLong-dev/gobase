/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:22:19
 * @LastEditTime: 2020-12-17 09:22:20
 * @LastEditors: Chen Long
 * @Reference:
 */

package mredis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ChenLong-dev/gobase/config"
	"github.com/ChenLong-dev/gobase/etcd"
	"github.com/ChenLong-dev/gobase/mlog"
	rds "github.com/go-redis/redis/v8"
	"time"
)

var (
	redisCluster *rds.ClusterClient
)

func init() {
	if config.Conf.RedisCluster == nil {
		return
	}
	mlog.Debugf("[RedisCluster] initial")
	if config.Conf.RedisCluster.Key != "" {
		if _, err := etcd.Connect(); err != nil {
			mlog.Errorf("[RedisCluster] Connect ETCD Fail")
			panic(err)
		}
		defer etcd.Close()
		redisConf, err := etcd.Get(config.Conf.Redis.Key)
		if err != nil {
			panic(fmt.Sprintf("[RedisCluster] Get etcd_config err: %s \n", err))
		}
		if err = json.Unmarshal(redisConf, &config.Conf.Redis); err != nil {
			panic(fmt.Sprintf("[RedisCluster] Get etcd_config err: %s \n", err))
		}
	}
	// 创建RedisCluster
	NewRedisCluster(config.Conf.RedisCluster)
}

func NewRedisCluster(cluster *config.RedisCluster) {
	clusterOpt := rds.ClusterOptions{
		Addrs:       cluster.Addrs,
		Password:    cluster.Password,
		PoolSize:    cluster.PoolSize,
		PoolTimeout: 10 * time.Second,
	}
	redisCluster = rds.NewClusterClient(&clusterOpt)
	cmd := redisCluster.Ping(context.TODO())
	if cmd.Val() != "PONG" {
		panic("[RedisCluster] NOT Connect")
	}
}

func RedisCluster() *rds.ClusterClient {
	return redisCluster
}
