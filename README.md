<!--
 * @Description: 
 * @Author: Chen Long
 * @Date: 2020-12-17 09:43:16
 * @LastEditTime: 2020-12-17 09:53:00
 * @LastEditors: Chen Long
 * @Reference: 
-->


# GoBase 公共组件

## 功能清单

* *config*: 配置文件解析
* *mlog*: 日志记录组件
* *etcd*: ETCD服务治理

## Golang/gRPC开发环境设置

### Golang 包镜像代理

设置golang包代理为公司镜像站
设置镜像代理才能自动下载依赖的第三方库。

windows
```sh
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off
```
linux or mac
```
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=off
```

### Protocolbuf 开发环境

将asset目录下的 protoc.exe+include文件夹，放到系统任一path环境目录下
protoc是用来编译proto脚本文件为目标语言代码的
include是以来的一些公共基础的pb
尝试一下命令，查看是否可用

```sh
　　protoc --version
```

### gRPC 的golang环境

安装以下开发插件
protoc-gen-go：Protocolbuf能生成golang源码的插件,
protoc-gen-go-grpc：能识别 rpc 关键字，并生成gRPC golang代理代码的插件

关于protoc-gen-go不能使用最新版本的问题，go.etcd.io/etcd v3.4.13依赖google.golang.org/grpc 最高版本为v1.29.1，高版本的protoc-gen-go依赖的grpc版本会有冲突

```sh
go get github.com/golang/protobuf/protoc-gen-go@v1.4.3 google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

## 私有仓库使用说明

### Golang Env设置

设置golang私有仓库

windows
```sh
go env -w GOPRIVATE=https://github.com/
```
linux or mac
```sh
export GOPRIVATE=https://github.com/
```

对私有仓库go get 时不启用默认的https请求

windows
```sh
go env -w GOINSECURE=https://github.com/
```

linux or mac
```
export GOINSECURE=https://github.com/
```

### Git 全局设置

由于仓库路径非go get标准命名，需要设置替换

```sh
　　git config --global url."git@github.com:ChenLong-dev/gobase.git".insteadof "https://github.com/ChenLong-dev/gobase.git"
```

### 验证获取

执行以下命令，能正确获取则设置成功

```sh
go get -v -u github.com/ChenLong-dev/gobase
```
