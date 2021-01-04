# 服务治理组件

## 功能清单

* *服务注册*: 向etcd注册，注册信息包括：ip，port，weight，并定时主动上报服务健康状态
* *服务发现*: 从etcd中读取指定service_name的节点列表， 为避免每次都访问etcd服务，本地用读写锁存储发现的服务信息
* *监听ETCD服务*: 主动监听ETCD服务变更，更新本地可用node节点列表，以及相应事件处理

## 服务注册

### 流程图

![注册流程图](../asset/etcd_reg.png)

### 功能说明

| 流程           | 目的（原因）
| ----------------    | -----
| 将注册信息（service_name，ip/port/weight）保存到本地       | 注册失败的情况，或者与etcd连接意外中断了的情况，后台有一条协程定期去尝试注册，直到成功为止
| 将此服务标记为已注册        | 后台协程根据这判断走register逻辑还是keepalive逻辑
| 根据租赁id获取keepalive的channel，保存到本地变量中        | go etcd的机制是，不断从这个channel取数据，就实现了keepalive，如果取到的数据是nil，说明和etcd连接断了，需要重新注册register逻辑还是keepalive逻辑

### 使用说明

```Golang
//初始化注册组件
reg := etcd.NewEtcdRegister("cluster_name")
//向ETCD集群注册服务信息
err := reg.Register("service", "key", "info_json")
if err != nil {
 t.Error("注册出错：", err)
 return
}
//向ETCD集群更新服务信息
err := reg.UpdateInfo("service", "key", "info_json")
if err != nil {
 t.Error("更新出错：", err)
 return
}
```

## 服务发现

### 流程图

![发现流程图](../asset/etcd_disc.png)

### 功能说明

业务在调用discovery(service_name)拿到可用的节点列表（ip，noce，weight），供负载均衡使用

在main()函数调用discovery(service_name)，失败时，业务自行决定是否panic

定期从ETCD拉取最新信息，并更新本地缓存。

### 使用说明

``` Golang
    //初始化发现组件
    disc := etcd.NewEtcdDis("cluster_name")
    //监听目标服务信息
    err := disc.Watch("service_name")
    if err != nil {
        t.Error("获取服务出错：", err)
        return
    }
    //获取服务健康的节点信息
    node, has := disc.GetServiceInfoRandom("service_name")
    if !has {
        t.Error("该服务无可用节点")
        return
    }
    //获取服务所有节点信息
    nodes, has := disc.GetServiceInfoAllNode("service_name")
    if !has {
        t.Error("该服务无可用节点")
        return
    }
```

## watch etcd（监听订阅）

### 功能说明

主动监听订阅ETCD相关目录，捕获PUT、DELETE等事件，实现配置的更新等。

### 使用说明

```Golang
暂无实现
```
