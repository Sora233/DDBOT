# 命令详细介绍及用法示例

*本介绍并不要求您拥有编程能力，但如果有，您会非常容易理解（因为命令的格式被设计为贴近shell命令）。*

## 命令基本格式

**bot只在群聊内工作，除特殊标注外，私聊命令无效**

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
  -s, --site="bilibili"    bilibili / douyu / youtube / huya
  -t, --type="live"        news / live
```

`Flags`里展示了`watch`命令的三个命令参数，分别是`-s/--site`、`-t/--type`、`-h/--help`，其中`--site="bilibili"`表示`-s/--site`
的默认参数是bilibili，而后面`bilibili / douyu / youtube / huya`部分则表示--site可以指定为这四个网站。

`-t/--type`同理。

`Arguments: <id>`表示watch命令需要一个主命令参数<id>，在例子1中为爽哥的UID`1472906636`。

DDBOT的命令格式皆遵守这个规则。

## 命令详细介绍

### /watch

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|QQ群管理员 / bot群管理员|是|是|

**该命令与unwatch命令共享权限**

订阅推送，支持推送b站直播，b站动态，斗鱼直播，YTB视频，YTB直播，虎牙直播

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

- 订阅虎牙乐爷的直播：https://www.huya.com/xiaoleyan

```shell
/watch -s huya xiaoleyan
```

### /watch （私聊版本）

- 在QQ群123456内订阅b站UID为2的用户的动态信息

```shell
/watch -g 123456 -t news 2
```

**一句话来说，用法同群聊一样，只是需要增加`-g 要操作的qq群号码`参数。**

### /unwatch

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|QQ群管理员 / bot群管理员|是|是|

**该命令与watch命令共享权限**

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

- 取消订阅虎牙乐爷的直播：https://www.huya.com/xiaoleyan

```shell
/unwatch -s huya xiaoleyan
```

**一句话来说，把watch命令原封不动的复制过来，并把`watch`替换成`unwatch`即可取消订阅。**

### /unwatch （私聊版本）

- 在QQ群123456内取消订阅b站UID为2的用户的动态信息

```shell
/unwatch -g 123456 -t news 2
```

**一句话来说，用法同群聊一样，只是需要增加`-g 要操作的qq群号码`参数。**

### /list

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|所有人|是|是|

查看当前订阅列表，同时会展示订阅的id，可以用来取消订阅。

- 查看直播订阅列表，会展示所有网站的所有订阅

```shell
/list
```

### /list（私聊版本）

- 查询QQ群123456的b站动态订阅列表

```shell
/list -g 123456
```

**一句话来说，用法同群聊一样，只是需要增加`-g 要操作的qq群号码`参数。**

### /config

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|QQ群管理员 / bot群管理员|是|是|

定制BOT，目前支持设置：

- 配置启用@全体成员，默认不@
- 配置@特定成员（已成功@全体成员时不会生效），默认不@
- 配置启用直播间标题更新推送，默认不推送
- 配置下播推送，默认不推送

#### 配置@全体成员

**同一个主播每2小时内只会@全体成员一次**

- 推送b站UID为2的用户的直播信息时，同时@全体成员（需要将BOT设置为管理员，否则配置后无法@全体成员）

```shell
/config at_all --site bilibili 2 on
```

- 取消上面的配置，不再@全体成员

```shell
/config at_all --site bilibili 2 off
```

#### 配置@特定成员

- 推送b站UID为2的用户直播信息时，当@全体成员无法生效时（包括未设置@全体成员配置，未被设置管理员权限，@全员剩余次数为0等情况），@特定成员

```shell
# 配置@QQ号为10000和10001的成员，QQ号可以一次填多个
/config at --site bilibili 2 add 10000 10001
```

- 删除配置中的特定成员

```shell
# 取消配置@QQ号为10000和10001的成员，QQ号可以一次填多个
/config at --site bilibili 2 remove 10000 10001
```

- 查看配置的成员列表

```shell
/config at --site bilibili 2 show
```

- 清空配置的成员列表

```shell
/config at --site bilibili 2 clear
```

#### 配置推送直播间标题更改

- 推送b站UID为2的用户的直播信息时，每当他的直播间标题更改时重新进行推送。

```shell
/config title_notify --site bilibili 2 on
```

#### 配置下播推送

- 推送b站UID为2的用户的直播信息时，当他下播时也进行推送。

```shell
/config offline_notify --site bilibili 2 on
```

#### 配置b站动态推送过滤器

*只能同时设置一种过滤器，如果多次设置，则以最后一次为准*

- 只推送指定种类的动态

```shell
/config filter type 97505 图片 文字 专栏 投稿
```

- 不推送指定种类的动态

```shell
/config filter not_type 97505 转发
```

- 动态关键字，只推送包含任意关键字的动态

```shell
/config filter text 97505 关键字1 关键字2
```

- 查看当前过滤器配置

```shell
/config filter show 97505
```

- 清空当前过滤器配置

```shell
/config filter clear 97505
```

**（b站动态种类繁多，无法全部列举，此处只支持部分常见类型）**

支持的动态类型：

- 专栏
- 转发
- 投稿
- 文字
- 图片
- 直播分享

### /config（私聊版本）

在QQ群123456内设置，推送b站UID为2的用户的直播信息时，同时@全体成员（需要将BOT设置为管理员，否则配置后无法@全体成员）

```shell
/config -g 123456 at_all --site bilibili 2 on
```

- 在QQ群123456内取消上面的配置，不再@全体成员

```shell
/config -g 123456 at_all --site bilibili 2 off
```

- 其他配置类似，不再重复列出

**一句话来说，用法同群聊一样，只是需要增加`-g 要操作的qq群号码`参数。**

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

### /grant (私聊版)

- 在QQ群654321内给予QQ号为123456的成员使用`/watch`命令的权限

```shell
/grant -g 654321 -c watch 123456
```

**一句话来说，用法同群聊一样，只是需要增加`-g 要操作的qq群号码`参数。**

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

### /enable 与 /disable （私聊版）

- 在QQ群123456内禁用watch命令，调用`/watch`不再有任何反应，之前watch过的仍然正常推送，即无法新增订阅

```shell
/disable -g 123456 watch
```

- 在QQ群123456内启用watch命令，调用`/watch`恢复正常使用，即允许新增订阅

```shell
/enable -g 123456 watch
```

**一句话来说，用法同群聊一样，只是需要增加`-g 要操作的qq群号码`参数。**

### /help

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|所有人|是|是|

打印帮助信息，可以帮助宣传本项目。

*现在可以禁用help命令了*

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

***由于该功能已有失控的趋势，现在不能再指定数量和图片tag了。***

返回p站图片（NSFW但非r18），~~一次最多10张~~。

如果你发现图片有某种倾向，那它就是。

一些例子：

- 返回一张图片

```shell
/色图
```

<details>
  <summary>该功能已删除</summary>

- 返回3张`图片tag`包含`刻晴`的p站图片

```shell
/色图 -t 刻晴 3
```

</details>

### 涩图判定

|默认使用权限|默认启用|是否可禁用|
|----------|-------|--------|
|所有人|否|是|

***该命令默认关闭，需要手动使用enable命令开启***

***由于阿里云开始收费了，开启后有效期为24小时。***

对比较涩的图片返回`就这`，对足够涩的图片返回`多发点`，其他图片无反应。

**不会对BOT自己发的图片有反应。**

触发方式：启用后自动触发，使用`/enable ic`开启，使用`/disable ic`关闭。

一些例子：

```
[图片]
```

## 管理员命令

管理员命令，仅限于管理员使用，主要面向私有部署场景

**所有管理员命令只能私聊使用**

### /sysinfo

查询bot当前好友数、群组数、订阅数量

例子：

```shell
/sysinfo
```

返回结果：

```
当前好友数：88
当前群组数：84
当前Bilibili订阅数：207
当前Douyu订阅数：2
当前YTB订阅数：28
当前Huya订阅数：1
```

### /disable --global 与 /enable --global

用于管理员控制命令的启停

注意：

该命令不会覆盖单个群的设置，即：

假设命令C在A群被禁用，在B群开启，被全局禁用后在AB群均无法使用，此时再次全局启用，仍然保持A群禁用，B群启用的状态。

一些例子：

- 全局禁用help命令，在群组内无法再使用/help命令

```shell
/disable --global help
```

- 全局禁用watch命令，无法再使用/watch命令

```shell
/disable --global watch
```

- 全局启用watch命令，可以正常使用/watch命令

```shell
/enable --global watch
```

### /log

用于管理员临时查询bot日志

一些例子：

- 查询最后10条日志

```shell
/log -n 10
```

- 查询包含有关键字error的日志，返回其中的最后20条 （先按关键字过滤，再按条数过滤）

```shell
/log -n 20 -k error
```

### /ping

用于管理员ping bot

例子：

```shell
/ping
```

返回：

```
pong
```