/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 10:44:33
 * @LastEditTime: 2020-12-16 11:50:23
 * @LastEditors: Chen Long
 * @Reference:
 */

package config

import (
	"github.com/BurntSushi/toml"
	"github.com/ChenLong-dev/gobase/mlog"
	"os"
	"reflect"
	"runtime"
)

//属性设计为指针类型
type Config struct {
	AppMode      string
	Http         *ServerCfg
	GRPC         *ServerCfg
	Es           *Es
	Etcd         *Etcd
	Mongo        *Mongo
	RedisCluster *RedisCluster
	Redis        *Redis
	Sms          *Sms
	Email        *Email
}

type ServerCfg struct {
	Host   string
	Port   string
	PreFix string
}
type Es struct {
	Host     string
	UserName string
	Password string
	Index    string
	Key      string
}

type Etcd struct {
	Hosts       []string
	UserName    string
	Password    string
	EtcdCert    string
	EtcdCertKey string
	EtcdCa      string
}

type Mongo struct {
	Host     string
	UserName string
	Password string
	DbName   string
	AuthDB   string
	MaxConns uint64
	Port     int
	Key      string
}

type RedisCluster struct {
	Addrs    []string
	Password string
	PoolSize int
	Key      string
}

type Redis struct {
	Addr     string
	Password string
	DB       int
	Key      string
}

type LogConfig struct {
	Log *Log
}

type Log struct {
	Path       string
	Level      int
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

type Sms struct {
	Host     string
	Api      string
	Account  string
	Password string
	Switch   string
	Key      string
}

type Email struct {
	From     string
	Name     string
	Server   string
	Port     int
	Mail     string
	Password string
	Switch   string
	Key      string
}

var Conf Config
var LogConf LogConfig

func init() {
	SAAS_COMMON_CONFIG := os.Getenv("SAAS_COMMON_CONFIG")
	path := "./conf/" + SAAS_COMMON_CONFIG + ".toml"
	if runtime.GOOS == "windows" {
		path = ".\\conf\\" + SAAS_COMMON_CONFIG + ".toml"
	}
	mlog.Infof("[CONFIG] current config path: %s", path)
	//读取本地配置获取服务IP、端口、监听url和ETCD配置信息
	if _, err := toml.DecodeFile(path, &Conf); err != nil {
		panic(err)
	}
	//读取本地配置获取服务日志配置
	if _, err := toml.DecodeFile(path, &LogConf); err != nil {
		panic(err)
	}

}

func InitExt(v interface{}) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		panic("Decode of non-pointer or nil ")
	}
	SAAS_COMMON_CONFIG := os.Getenv("SAAS_COMMON_CONFIG")
	path := "./conf/" + SAAS_COMMON_CONFIG + ".toml"
	if runtime.GOOS == "windows" {
		path = ".\\conf\\" + SAAS_COMMON_CONFIG + ".toml"
	}
	//读取本地配置获取,项目自定义配置项
	if _, err := toml.DecodeFile(path, v); err != nil {
		panic(err)
	}
}
