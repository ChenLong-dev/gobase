<!--
 * @Description: 
 * @Author: Chen Long
 * @Date: 2020-12-16 12:36:43
 * @LastEditTime: 2020-12-16 12:36:54
 * @LastEditors: Chen Long
 * @Reference: 
-->
# 日志组件

## 功能清单

* *log.go*: 标准日志接口实现

### 功能说明

标准日志接口实现

### 使用说明

```Golang
    //日志初始化
    mlog.Init(&mlog.Params{
        Path:       "./server_name/logs.log", //文件路径
        MaxSize:    2, //MB 单个日志文件最大
        MaxBackups: 3, //备份个数
        MaxAge:     10, //保存时间,天
        Level:      1,  //# 日志级别 [0：All,所有,1:TRACE 跟踪,2:DEBUG 调试,3:INFO 信息,4:WARN 警告,5:ERROR 一般错误,6:FATAL 致命错误]
    })  
    //初始化后，直接使用 包 级别函数
    mlog.Info("看下日志写到哪里去了！")
```

