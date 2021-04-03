# 命令详细介绍及用法示例

*本介绍并不要求您拥有编程能力，但如果有，您会非常容易理解（因为命令的格式被设计为贴近shell命令）。*

## 命令基本格式

**bot只在群聊内工作，私聊命令无效**

```
/主命令 [命令选项 [选项参数 ...] ...] [主命令参数 ...]
```

#### 例子：

我买了海盗床，我要车爽哥，如何使用watch命令订阅爽哥b站直播信息？

- 进入爽哥b站个人[空间主页](https://space.bilibili.com/1472906636)

- 找到爽哥的UID`1472906636`

```shell
/watch --site bilibili --type live 1472906636
```

在这个例子中，主命令为`watch`，它包含两个命令选项，分别是`--site`和`--type`，两个命令选项的参数分别是`bilibili`和`live`，有一个主命令参`1472906636`。

`--site`用来指定订阅网站，`--type`用来指定订阅类型，所以上面的命令表示：订阅**b站**的UID为**1472906636**的用户的**直播**信息。

命令选项有长和短两种模式，两种模式效果相同，这个例子中使用的是长模式，它的格式为`--`加上单词，短模式格式为`-`加上单个字母。

大多数长模式都可以用`-`加上单词的首字母简写成短模式。

例如`--site`简写成`-s`，`--type`简写成`-t`，那么命令可以写成：

```shell
/watch -s bilibili -t live 1472906636
```

更进一步，对于`watch`命令来说，`-s`的默认参数是`bilibili`，`-t`的默认参数是`live`。也就是说，如果不指定`-s`，那么默认网站是`bilibili`；如果不指定`-t`，那么默认类型是`live`
，所以命令进一步简写成：

```shell
/watch 1472906636
```

## 命令帮助信息

几乎所有命令都带有自身的帮助信息，您可以使用选项参数`-h/--help`查看命令的帮助信息。

以`watch`命令为例，输入`/watch -h`或者`/watch --help` （短模式和长模式），会输出：

```
Usage: watch <id>

Arguments:
  <id>

Flags:
  -h, --help               Show context-sensitive help.
  -s, --site="bilibili"    bilibili / douyu / youtube
  -t, --type="live"        news / live
```

`Flags`里展示了`watch`命令的三个命令参数，分别是`-s/--site`、`-t/--type`、`-h/--help`，其中`--site="bilibili"`表示`-s/--site`
的默认参数是bilibili，而后面`bilibili / douyu / youtube`部分则表示--site可以指定为这三个网站。

`-t/--type`同理。

`Arguments: <id>`表示watch命令需要一个主命令参数<id>，在例子1中为爽哥的UID`1472906636`。

DDBOT的命令格式皆遵守这个规则。

## 命令详细介绍

### /watch

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|QQ群管理员 / bot群管理员|是|是|

订阅推送，支持推送b站直播，b站动态，斗鱼直播，YTB视频，YTB直播

一些例子：

- 订阅b站UID为2的用户的直播信息

```shell
/watch 2
```

- 订阅b站UID为2的用户的动态信息

```shell
/watch -t news 2
```

- 订阅斗鱼6655直播间 ~~钢之魂，我的钢之魂~~

```shell
/watch -s douyu 6655
```

- 订阅YTB乙女音频道的直播 https://www.youtube.com/channel/UCvEX2UICvFAa_T6pqizC20g

```shell
/watch -s youtube UCvEX2UICvFAa_T6pqizC20g
```

- 订阅YTB乙女音频道的视频 https://www.youtube.com/channel/UCvEX2UICvFAa_T6pqizC20g

```shell
/watch -s youtube -t news UCvEX2UICvFAa_T6pqizC20g
```

### /unwatch

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|QQ群管理员 / bot群管理员|是|是|

取消订阅推送，和`watch`用法一样

一些例子：

- 取消订阅b站UID为2的用户的直播信息

```shell
/unwatch 2
```

- 取消订阅b站UID为2的用户的动态信息

```shell
/unwatch -t news 2
```

- 取消订阅YTB乙女音频道的直播 https://www.youtube.com/channel/UCvEX2UICvFAa_T6pqizC20g

```shell
/unwatch -s youtube UCvEX2UICvFAa_T6pqizC20g
```

- 取消订阅YTB乙女音频道的视频 https://www.youtube.com/channel/UCvEX2UICvFAa_T6pqizC20g

```shell
/unwatch -s youtube -t news UCvEX2UICvFAa_T6pqizC20g
```

**一句话来说，把watch命令原封不动的复制过来，并把`watch`替换成`unwatch`即可取消订阅。**

### /list

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|所有人|是|是|

查看当前订阅列表，同时会展示订阅的id，可以用来取消订阅。

- 查看当前b站直播订阅列表

```shell
/list -s bilibili -t live
```

- 查看当前b站动态订阅列表

```shell
/list -s bilibili -t news
```

- 查看当前YTB直播订阅列表

```shell
/list -s youtube -t live
```

### /grant

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|QQ群管理员 / bot群管理员|是|否|

给予或撤销其他成员命令权限或者角色权限。

一些例子：

- 给予QQ号为123456的成员使用`/watch`命令的权限

```shell
/grant -c watch 123456
```

- 给予QQ号为123456的成员bot群管理员权限（触发命令的人必须有bot群管理员权限）

```shell
/grant -r GroupAdmin 123456
```

- 撤销QQ号为123456的成员使用`/watch`命令的权限

```shell
/grant -d -c watch 123456
```

**一句话来说，把grant命令原封不动的复制过来，并加上`-d`命令选项即可撤销权限。**

### /enable 与 /disable

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|bot群管理员|是|否|

启用命令与禁用命令，两者配套使用。

一些例子：

- 禁用watch命令，调用`/watch`不再有任何反应，之前watch过的仍然正常推送，即无法新增订阅

```shell
/disable watch
```

- 禁用unwatch命令，调用`/unwatch`不再有任何反应，即无法取消订阅

```shell
/disable unwatch
```

- 启用watch命令，调用`/watch`恢复正常使用，即允许新增订阅

```shell
/enable watch
```

### /help

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|所有人|是|否|

打印帮助信息，可以帮助宣传本项目。

一些例子：

```shell
/help
```

### /roll

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|所有人|是|是|

产生随机数，或者从多个选择中随机返回一个

一些例子：

- 生成一个1~100范围的随机数

```shell
/roll
```

- 生成一个1~10范围的随机数

```shell
/roll 10
```

- 生成一个100~200范围的随机数

```shell
/roll 100-200
```

- 从`画画`、`不画`、`明天再画`三个中随机返回一个

```shell
/roll 画画 不画 明天再画
```

### /倒放

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|所有人|是|是|

只对gif生效，生成倒放过的gif。

触发方式有两种：

```
/倒放 [图片]
```

由于手机端不太好使用上面的格式，因此也支持回复触发

```
【回复图片消息】/倒放
```

### /face

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|所有人|是|是|

对图片或gif进行人脸识别，并在人脸处用红框标识。

触发方式同`/倒放`

### /签到

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|所有人|是|是|

每天可签到1次，获得1积分（目前积分暂无其他用处）

一些例子：

```shell
/签到
```

### /色图

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|所有人|否|是|

***该命令默认关闭，需要手动使用enable命令开启***

返回p站图片（NSFW但非r18），一次最多10张。

一些例子：

- 返回一张图片

```shell
/色图
```

- 返回3张`图片tag`包含`刻晴`的p站图片

```shell
/色图 -t 刻晴 3
```

### 涩图判定

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|所有人|否|是|

***该命令默认关闭，需要手动使用enable命令开启***

对比较涩的图片返回`就这`，对足够涩的图片返回`多发点`，其他图片无反应。

触发方式：启用后自动触发，使用`/enable ic`开启，使用`/disable ic`关闭。

一些例子：

```
[图片]
```