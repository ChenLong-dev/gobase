<!--
 * @Description: 
 * @Author: Chen Long
 * @Date: 2020-12-16 10:48:09
 * @LastEditTime: 2020-12-16 10:54:08
 * @LastEditors: Chen Long
 * @Reference: 
-->

# 配置文件组件

## 功能清单

* *config.go*: 常用的基本配置封装，公共组件里相关配置会直接依赖此项
* *.toml*: 配置文件模板示例，以及相关解释

### 功能说明

本组件利用init函数，再模块第一次被引用时，会自动尝试读取项目的配置文件，并进行解析。默认路径为："./conf/" + os.Getenv("SAAS_COMMON_CONFIG") + ".toml"

### 使用说明

```Golang
//显式制定进行初始化
import _ "config"

//初始化后可直接使用相关配置信息
println("服务启动模式：", config.Conf.AppMode)

//自定义扩展配置使用方式

type ExtendCfg struct {
    Extends *ExtendCtx
}

//自定义扩展配置的具体配置内容项，与toml里文件一致即可，根据实际业务自己扩展
type ExtendCtx struct {
    ExtCfgStr string
    ExtCfgInt int
}

extLog := ExtendCfg{}
//从默认配置文件中，主动初始化扩展配置，必须 指针类型
config.InitExt(&extLog)
println(extLog.Extends.ExtCfgStr)

```
