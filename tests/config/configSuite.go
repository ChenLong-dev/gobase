/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:32:35
 * @LastEditTime: 2020-12-17 09:32:35
 * @LastEditors: Chen Long
 * @Reference:
 */

package config

import (
	mconfig "github.com/ChenLong-dev/gobase/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"reflect"
)

type ConfigTestSuite struct {
	suite.Suite

	Mongo        *mconfig.Mongo
	Etcd         *mconfig.Etcd
	RedisCluster *mconfig.RedisCluster
}

//初始化
func (suite *ConfigTestSuite) SetupTest() {
	suite.Mongo = &mconfig.Mongo{
		Host:     "mongodb://10.107.30.158:27017",
		Port:     27017,
		UserName: "root",
		Password: "123",
		DbName:   "demo",
		AuthDB:   "admin",
	}
	suite.Etcd = &mconfig.Etcd{
		Hosts: []string{
			"10.227.30.129:2379", "10.227.30.130:2379", "10.227.30.131:2379",
		},
		UserName:    "root",
		Password:    "saas_etcd_root123",
		EtcdCert:    "/saasdata/etcd/data/ssl/server.pem",
		EtcdCertKey: "/saasdata/etcd/data/ssl/server-key.pem",
		EtcdCa:      "/saasdata/etcd/data/ssl/ca.pem",
	}
	//suite.RedisCluster = &mconfig.RedisCluster{
	//	Addrs: []string{
	//		"10.107.30.161:6379", "10.107.30.161:6379", "10.107.30.171:6379", "10.107.30.161:6380", "10.107.30.161:6380", "10.107.30.171:6380",
	//	},
	//	Password: "saas_redis_root123",
	//	PoolSize: 100,
	//}
}

func (suite *ConfigTestSuite) TestMongo() {
	assert.Equal(suite.T(), true, reflect.DeepEqual(suite.Mongo, mconfig.Conf.Mongo))
}

func (suite *ConfigTestSuite) TestEtcd() {
	assert.Equal(suite.T(), true, reflect.DeepEqual(suite.Etcd, mconfig.Conf.Etcd))
}

//func (suite *ConfigTestSuite) TestRedisCluster() {
//	assert.Equal(suite.T(), true, reflect.DeepEqual(suite.RedisCluster, mconfig.Conf.RedisCluster))
//}
