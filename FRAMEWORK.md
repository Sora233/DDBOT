# 为DDBOT编写插件

**警告：现阶段DDBOT还会继续重构一部分，会尽力（但不是绝对）保证代码的兼容性。**

*阅读这个内容需要一定的开发能力。*

为DDBOT编写插件的基本步骤为：

- 实现`concern.Concern`接口
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

更多文档正在施工中。