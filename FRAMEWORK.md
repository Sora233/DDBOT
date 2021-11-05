# 为DDBOT编写插件

*阅读这个内容需要一定的开发能力。*

为DDBOT编写插件的基本步骤为：

- 通过`concern.StateManager`，实现`concern.Concern`接口
- 在`init()`函数中进行注册
- 在`main`中引入刚刚编写的包

请参考`_example/concern`目录，该目录为一个样例插件，展示了为DDBOT编写插件的方法。

这个例子为网站`example`新增了一个类型`example`，可以使用对应的`watch`命令进行订阅：

```
/watch -s example -t example ID
```

使用unwatch进行取消订阅：

```
/unwatch -s example -t example ID
```

使用list进行查看订阅：

```
/list
```

## 一些开箱即用的组件

**示例代码为了简洁，均忽略错误处理，在实际编写代码的过程中，请注意错误处理。**

在编写插件的过程中，第一步通常是通过创建新的`StateManager`，下面的示例均基于这个结构：

```go
type StateManager struct {
*concern.StateManager
}
var s *StateManager = ... // 初始化
```

### 持久化

在编写插件的过程中，总会遇到想要存储一些数据的需求，DDBOT默认提供了一个key-value数据库可以使用，支持事务，key过期，并且已经包含在`StateManager`中：

```go
s.SetInt64("myInt64", 123456) // 即可往数据库中设置kv对 "myInt64" - 123456 
v, _ := s.GetInt64("myInt64") // 获取刚刚写入的值
// v == 123456
```

还有更多方法请参考`buntdb/shortcut.go`

### 更多文档正在施工中。