# 部署指南

首先需要获取DDBOT程序，可以选择下载预编译的版本，也可以选择从源码编译。

如果您不清楚如何进行编译，请选择预编译版本。

## 预编译版本

从 [Release](https://github.com/Sora233/DDBOT/releases) 中下载适合您的版本，版本规则：

- 根据系统选择windows / linux / darwin
- 根据架构选择 32位->386 / 64位->amd64 / arm->arm

例如对于windows 7/8/10/11/server，64位系统，推荐选择`-windows-amd64.zip`

<details>
<summary>从源码编译（不推荐，需要编程能力）</summary>

## 从源码编译

首先安装go，go >= 1.17

然后执行下面的命令

```
make build

# 如果没有安装make程序，则用下面的命令编译
go build
```

</details>

## 配置

获取的DDBOT程序为二进制程序，可以直接运行。

**Linux**下请确认DDBOT程序对所在目录有读写权限，并需要赋予程序可执行权限，可通过下面的命令实现：`chmod +x ./DDBOT`

然后可以通过`./DDBOT`运行DDBOT程序。

**windows**版DDBOT默认应该有.exe后缀，即程序名为`DDBOT.exe`，如果没有后缀，请确认下载的程序是windows版本。

Windows版可以双击运行DDBOT程序。

首次运行默认会通过二维码登陆，并且会自动生成所有配置。

DDBOT的配置文件为`application.yaml`，该文件可以用文本编辑器（windows的记事本/vscode，linux的nano/vim等）打开。

### 设置BOT管理员

**方法一**

初次运行bot时，默认没有管理员，此时您可以对bot私聊发送`/whosyourdaddy`命令，成功后发送命令的账号将成为bot管理员。

*请放心，只有当bot没有管理员时可以使用此命令，后续发送此命令将无法生效*

<details>
<summary>方法二（不推荐）</summary>

bot未运行时，执行以下命令。

```shell
./DDBOT --set-admin 你的QQ号码
```

</details>

### 设置BOT运行模式

*该功能从v0.1.0版本开始支持，v0.1.0之前的版本可以跳过这一段*

初次运行bot时，bot默认为公开模式，会自动同意好友申请与加群邀请，您可以通过以下命令切换：

- 切换到私人模式，该模式下会自动拒绝好友申请与加群邀请，您可以手动登陆bot QQ主动加好友与群

```shell
/mode 私人
```

- 切换到审核模式，该模式下会保留好友申请与加群邀请，您可以选择同意或是拒绝。

```shell
/mode 审核
```

- 您也可以重新切换回公开模式

```shell
/mode 公开
```

### 减少bot的刷屏输出

*该功能从v0.1.3版本开始支持，v0.1.3之前的版本可以跳过这一段*

如果您已经对bot命令比较熟悉，同时您的bot为非公开地小范围内使用，您可以对bot私聊发送`/silence`命令减少bot的输出内容。

这个操作会全局设置沉默模式，bot的输出信息将会减少，不再输出权限不足、命令被禁用、命令帮助等信息。

沉默模式也可以对单个群开启，详情请看命令文档。

**如果您对bot的命令尚不熟悉，建议暂时不开启这个设置。**

### device.json

device.json是运行时使用的设备信息（可以理解为伪装的手机型号），应尽量使用同一个，否则可能会触发安全机制无法登陆。

如果检测到不存在，该文件会自动生成，正常情况无需关心。

### application.yaml

DDBOT运行时的配置文件，可以用记事本打开修改，如果检测到不存在，会生成一个最小配置。

#### 最小配置

测试时推荐使用扫码登陆，即不指定帐号和密码（注意：掉线或重启后无法自动重连，仍需扫码，仅推荐测试使用）。

最小配置可以正常使用斗鱼、虎牙、B站订阅功能（最小配置下，B站订阅推荐不超过5个，过多订阅会导致推送延迟大幅增加）。

如果可以直连油管，那也可以正常使用YTB订阅，如果无法直连油管，则需要配置代理功能，YTB订阅才能正常工作。

如果要大规模使用b站订阅功能，推荐配置B站帐号，订阅数最高可达2000（初号机已验证2000订阅）。

如果要使用 TwitCasting 订阅功能，请自行参考完整配置。

```yaml
# 注意，填写时请把井号及后面的内容删除，并且冒号后需要加一个空格
bot:
  account:  # 你bot的qq号，不填则使用扫码登陆
  password: # 你bot的qq密码
  onJoinGroup:
    rename: "【bot】"  # BOT进群后自动改名，默认改名为“【bot】”，如果留空则不自动改名

# 初次运行时将不使用b站帐号方便进行测试
# 如果不使用b站帐号，则推荐订阅数不要超过5个，否则推送延迟将上升
# b站相关的功能推荐使用一个b站账号，建议使用小号
# bot将使用您b站帐号的以下功能：
# 关注用户 / 取消关注用户 / 查看关注列表
# 请注意，订阅一个账号后，此处使用的b站账号将自动关注该账号
bilibili:
  SESSDATA: # 你的b站cookie
  bili_jct: # 你的b站cookie
  interval: 25s

concern:
  emitInterval: 5s

logLevel: info

```

#### 完整配置

<details>
<summary> 点此查看全部配置 （仅供高级用户参考） </summary>

```yaml
bot:
  account: # bot账号
  password: # bot密码
  commandPrefix: "/"     # bot触发命令的前缀，默认为单斜杠 /
  onDisconnected: "exit" # 设置掉线时处理方式，exit为退出，不填或者其他值为尝试重连
  onJoinGroup:
    rename: "【bot】"     # BOT进群后自动改名，默认改名为“【bot】”，如果留空则不自动改名

# 请注意，bot将使用您b站帐号的以下功能，建议使用新注册的小号：
# 关注用户 / 取消关注用户 / 查看关注列表
# 目前支持填cookie和账号两种方式 （选择任意一种方式即可，推荐使用账号密码）
# 若使用账号
    # 直接填入账号密码
# 若使用cookie
    # b站登陆后的cookie字段，从cookie中找到这两个填进去
    # 警告：
    # SESSDATA和bili_jct等价于您的帐号凭证
    # 请绝对不要透露给他人，更不能上传至Github等公开平台
    # 否则将导致您的帐号被盗
# 请注意，订阅一个账号后，此处使用的b站账号将自动关注该账号
bilibili:
  SESSDATA: ""
  bili_jct: ""
  account:  # 你的b站账号 （目前不可用）
  password: # 你的b站密码 （目前不可用）
  interval: 25s # 直播状态和动态检测间隔，过快可能导致ip被暂时封禁
  imageMergeMode: "auto" # 设置图片合并模式，支持 "auto" / "only9" / "off"
                         # auto 为默认策略，存在比较刷屏的图片时会合并
                         # only9 表示仅当恰好是9张图片的时候合并
                         # off 表示不合并
  hiddenSub: false    # 是否使用悄悄关注，默认不使用
  unsub: false        # 是否自动取消关注，默认不取消，如果您的b站账号有多个bot同时使用，取消可能导致推送丢失
  minFollowerCap: 0        # 设置订阅的b站用户需要满足至少有多少个粉丝，默认为0，设为-1表示无限制
  disableSub: false        # 禁止ddbot去b站关注帐号，这意味着只能订阅帐号已关注的用户，或者在b站手动关注
  onlyOnlineNotify: false  # 是否不推送Bot离线期间的动态和直播，默认为false表示需要推送，设置为true表示不推送

localPool: # 图片功能，使用本地图库
  imageDir: # 本地路径

loliconPool: # 图片功能，使用api.lolicon.app图库
  apikey:    # 由于该图库更新，此字段不再需要了，留空即可
  cacheMin: 10
  cacheMax: 50
  proxy:

pyProxyPool: # 代理池配置，py代理池 https://github.com/jhao104/proxy_pool
  host: http://127.0.0.1:5010

localProxyPool: # 代理池配置，固定代理
  oversea: # 可翻墙的代理，用于访问YTB或pixiv
    - 127.0.0.1:8888
  mainland: # 不可翻墙的代理，用于直连国内网站
    - 127.0.0.1:8888

# 加入 twitcasting 部分即启用 tc 订阅功能  
# 参阅 https://apiv2-doc.twitcasting.tv/#registration
# 你需要到 https://twitcasting.tv/developer.php 新增一个 App
# 填入所需资料后就可以获取 clientId 和 clientSecret
# 详细: https://www.codenong.com/7d930656a7499340f14a/
twitcasting:
  clientId: abc
  clientSecret: xyz
  # 为防止风控，可选择性广播以下元素
  broadcaster:
    title: false # 标题 (有风控机会)
    created: true # 开播时间
    image: false # 直播封面 (墙内无法获取TC直播封面，建议有代理才开启
  # 自定义名称显示，日文字太多很大机会风控
  # 例如 "(如何显示) 正在直播""
  nameStrategy: "name" # 如何显示名称, name= 显示用户名称, userid= 显示用户ID, both= 显示 "用户名称 (用户ID)"

concern:
  emitInterval: 5s # 订阅的刷新频率，5s表示每5秒刷新一个ID，过快可能导致ip被暂时封禁

imagePool:
  type: "off" # localPool / loliconPool

proxy:
  type: "off" # localProxyPool/ pyProxyPool

debug: # debug模式，只有以下群或qq号可以触发命令
  group:
    - 0
  uin:
    - 0

message-marker: # 是否禁用自动已读功能，true为禁用，false为不禁用，默认为不禁用
  disable: false

qq-logs: # 是否启用在命令行内展示qq聊天内容，true为启用，false为禁用，默认为禁用
  enable: false

dispatch:
  largeNotifyLimit: 50 # 巨量推送的判定配置，默认为50，当大于这个配置时，将增大推送延迟保证账号稳定
notify:
  parallel: 1          # 增加推送消息的并发配置，默认为1以优先保证账号稳定，当出现推送堆积的时候可以尝试调高

template:       # 是否启用模板功能，true为启用，false为禁用，默认为禁用
  enable: false # 需要了解模板请看模板文档
  
autoreply: # 自定义命令自动回复，自定义命令通过模板发送消息，且不支持任何参数，需要同时启用模板功能
  group:   # 需要了解该功能请看模板文档
    command: [ ]
  private:
    command: [ ]

# 重定义命令前缀，优先级高于bot.commandPrefix
# 如果有多个，可填写多项，prefix支持留空，可搭配自定义命令使用
# 例如下面的配置为：<Q命令1> <命令2> </help>
customCommandPrefix:
  命令1: "Q"
  命令2: ""

timezone: Asia/Shanghai # 使用TZ database name自定义通知消息时间戳的时区，默认为部署服务器的本地时区

logLevel: info # 日志等级
```

</details>