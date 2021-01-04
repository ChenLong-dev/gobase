<!--
 * @Description: 
 * @Author: Chen Long
 * @Date: 2020-12-17 09:20:20
 * @LastEditTime: 2020-12-17 09:20:30
 * @LastEditors: Chen Long
 * @Reference: 
-->
# MongoDB组件

## 功能清单

* *mongo.go*: 封装了MongoDB的默认初始化，并提供全局的单例客户端和默认DB

### 功能说明

强依赖默认的配置文件组件，只解决最理想情况下，DB持久化功能

### 使用说明

```Golang
//显式制定进行初始化
import (
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
    "mongodb"
    )

//初始化后可直接使用
mongodb.GlobalDatabase.
        Collection("CollectionName").InsertOne(ctx, bson.M{})

```
