# Example Concern

这个包包含一个样例concern实现以供参考，支持：

- /watch -s example -t example ID

使用上面的site和type参数订阅，订阅后每分钟会推送一次订阅的ID

- /unwatch -s example -t example ID

取消上面的订阅

- /list

查看当前订阅

使用时需要在main中引入：

```_ "github.com/Sora233/DDBOT/_exapmle/concern"```
