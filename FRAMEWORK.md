# 为DDBOT编写插件

*阅读这个内容需要一定的开发能力。*

文档地址：
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/Sora233/DDBOT)

为DDBOT编写插件的基本步骤为：

- 通过`concern.StateManager`，实现`concern.Concern`接口
- 在`init()`函数中进行注册
- 在`main`中引入刚刚编写的包

DDBOT提供了一个插件脚手架，可以快速创建一个插件模版：[DDBOT-template](https://github.com/Sora233/DDBOT-template)

在编写插件前，请先阅读一个示例插件：[DDBOT-example](https://github.com/Sora233/DDBOT-example)

示例插件展示了为DDBOT编写插件的方法。

```golang
// 引入刚刚编写的插件
import (
    _ "github.com/Sora233/DDBOT-example/concern"
)
```

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

在编写插件的过程中，第一步通常是创建新的`StateManager`，下面的示例均基于这个结构：

```golang
type StateManager struct {
    *concern.StateManager
}
var s *StateManager = ... // 初始化
```

### 持久化

在编写插件的过程中，总会遇到想要存储一些数据的需求，DDBOT默认提供了一个key-value数据库可以使用，支持串行事务，key ttl，并且已经包含在`StateManager`中：

```golang
s.SetInt64("myInt64", 123456) // 即可往数据库中设置kv对 "myInt64" - 123456 
v, _ := s.GetInt64("myInt64") // 获取刚刚写入的值
// v == 123456

// 还有 s.SetJson() s.Set() 等更多方法
```

还有更多方法请参考`buntdb/shortcut.go`


### 轮询器

当订阅网站暂时没有发现比较高效的爬虫方式时，我们只能选择不断地依次访问每一个订阅目标的特定页面来获取目标是否有新信息，这种方式叫做轮询。

DDBOT内置了一个轮询器来满足这种方式：

```golang
// 首先需要启动 EmitQueue 轮询器
s.UseEmitQueue()
// 然后可以使用轮询器提供的helper创建一个 FreshFunc
s.UseFreshFunc(s.EmitQueueFresher(func(p concern_type.Type, id interface{}) ([]concern.Event, error) { 
    // id 是此时轮到的目标信息，此时可以刷新这个目标
    // p是id所有订阅过的Type的集合
}))
```

### 更多文档正在施工中。