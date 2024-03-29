# Apollo Golang 客户端

## 功能

* 多 namespace 支持
* 容错，本地缓存
* 零依赖
* 实时更新通知

## 安装

```sh
    go get -u github.com/liamylian/apollo-client
```

## 使用

### 使用 app.properties 配置文件启动

```golang
    apollo.Start()
```

### 使用自定义配置启动

```golang
    apollo.StartWithConfFile(name)
```

### 监听配置更新

```golang
    events := apollo.WatchUpdate()
    changeEvent := <-event
    bytes, _ := json.Marshal(changeEvent)
    fmt.Println("event:", string(bytes))
```

### 获取配置

```golang
    apollo.GetStringValue(Key, defaultValue)
    apollo.GetStringValueWithNameSapce(namespace, key, defaultValue)
```

### 获取文件内容

```golang
    apollo.GetNameSpaceContent(namespace, defaultValue)
```

### 获取配置中所有的键

```golang
    apollo.GetAllKeys(namespace)
```

### 订阅namespace的配置

```golang
    apollo.SubscribeToNamespaces("newNamespace1", "newNamespace2")
```
