/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:21:45
 * @LastEditTime: 2020-12-17 09:21:46
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
)

var (
	redisClient *rds.Client
)

func init() {
	if config.Conf.Redis == nil {
		return
	}
	mlog.Debugf("[Redis] initial")
	//判断config中mongo字段中key是否为空
	if config.Conf.Redis.Key != "" {
		//从etcd获取字段
		if _, err := etcd.Connect(); err != nil {
			panic("[REDIS] Connect ETCD Fail")
		}
		defer etcd.Close()
		redisClusterConf, err := etcd.Get(config.Conf.Redis.Key)
		if err != nil {
			panic(fmt.Sprintf("[REDIS] Get etcd_config err: %s \n", err))
		}
		if err = json.Unmarshal(redisClusterConf, &config.Conf.RedisCluster); err != nil {
			panic(fmt.Sprintf("[REDIS] Get etcd_config err: %s \n", err))
		}
	}
	// 创建Redis
	NewRedisClient(config.Conf.Redis)
}

func NewRedisClient(redis *config.Redis) {
	redisOpt := rds.Options{
		Addr:     redis.Addr,
		Password: redis.Password,
		DB:       redis.DB,
	}
	redisClient = rds.NewClient(&redisOpt)
	cmd := redisClient.Ping(context.TODO())
	if cmd.Val() != "PONG" {
		panic("[REDIS] NOT Connect")
	}
}

func RedisClient() *rds.Client {
	return redisClient
}
