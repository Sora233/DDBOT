# Example Concern

这个包包含一个样例concern实现以供参考，支持：

- /watch -s example -t example ID

使用上面的site和type参数订阅，订阅后每分钟会有50%的概率推送一次订阅的ID

- /unwatch -s example -t example ID

取消上面的订阅

- /list

查看当前订阅

使用时需要在main中引入：

```_ "github.com/Sora233/DDBOT/_exapmle/concern"```

## 目录结构

### notify.go

这个文件通过实现`concern.Notify`定义了推送的格式

### config.go

这个文件重写了`concern.Hook`，通过`FilterHook`实现50%概率推送

### concern.go

这个文件实现了`concern.Concern`，并使用了自带的`EmitQueue`作为状态刷新机制，产生`concern.Notify`事件，并发送给框架进行推送