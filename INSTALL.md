# 部署指南

首先需要获取DDBOT程序，可以选择下载预编译的版本，也可以选择从源码编译。

如果您不清楚如何进行编译，请选择预编译版本。

## 预编译版本

*预编译版本均为nocv版本，即无法正常使用face命令（本来也没啥人用这个命令）。*

从 [Release](https://github.com/Sora233/DDBOT/releases) 中下载适合您的版本，版本规则：

- 根据系统选择windows / linux / darwin
- 根据架构选择 32位->386 / 64位->amd64 / arm->arm

<details>
<summary>从源码编译（不推荐）</summary>

## 从源码编译

go >= 1.16

编译可选择是否启动gocv，如果不需要使用face命令，推荐禁用gocv。（因为安装opencv容易出现问题）

- ### 启用gocv

请先安装 [gocv](https://github.com/hybridgroup/gocv)

```
make build

# 如果没有安装make程序，则用下面的命令编译
go build
```

- ### 禁用gocv

禁用gocv将导致face命令无法正常工作

```
NOCV=1 make build

# 如果没有安装make程序，则用下面的命令编译
go build -tags nocv
```

</details>

## 配置

下载后可以直接运行，首次运行默认会通过二维码登陆，并且会自动生成所有配置。

DDBOT的配置文件为`application.yaml`，该文件可以用文本编辑器打开。

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

### device.json

device.json是运行时使用的设备信息（可以理解为伪装的手机型号），应尽量使用同一个，否则可能会触发安全机制无法登陆。

如果检测到不存在，该文件会自动生成，正常情况无需关心。

### application.yaml

DDBOT运行时的配置文件，可以用记事本打开修改，如果检测到不存在，会生成一个最小配置。

#### 最小配置

测试时推荐使用扫码登陆，即不指定帐号和密码（注意：掉线或重启后无法自动重连，仍需扫码，仅推荐测试使用）。

最小配置可以正常使用斗鱼、虎牙订阅功能。

如果可以直连油管，那也可以正常使用YTB订阅，如果无法直连油管，则需要配置代理功能，YTB订阅才能正常工作。

如果要使用b站订阅功能，需要填入b站账号和密码。

```yaml
bot:
  account:  # 你的qq号，不填则使用扫码登陆
  password: # 你的qq密码

# b站相关的功能需要一个b站账号，建议使用新注册的小号
# bot将使用您b站帐号的以下功能：
# 关注用户 / 取消关注用户 / 查看关注列表
# 请注意，订阅一个账号后，此处使用的b站账号将自动关注该账号
bilibili:
  account:  # 你的b站账号 
  password: # 你的b站密码
  interval: 20s

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
  onDisconnected: "exit" # 设置掉线时处理方式，exit为退出，不填或者其他值为尝试重连

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
  account:  # 你的b站账号 
  password: # 你的b站密码
  interval: 30s # 直播状态和动态检测间隔，过快可能导致ip被暂时封禁
  imageMergeMode: "auto" # 设置图片合并模式，支持 "auto" / "only9" / "off"
                         # auto 为默认策略，存在比较刷屏的图片时会合并
                         # only9 表示仅当恰好是9张图片的时候合并
                         # off 表示不合并


# 用于涩图鉴定功能
# https://api.aliyun.com/#/?product=imageaudit
aliyun: # 阿里云鉴权，用于图片识别，如果为空则图片识别功能无法工作
  accessKeyID: # 鉴权的账号应该开通内容审核功能（该服务2021年3月25日开始收费）
  accessKeySecret:

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

concern:
  emitInterval: 5s # 订阅的刷新频率，5s表示每5秒刷新一个ID，过快可能导致ip被暂时封禁

zhimaProxyPool: # 代理池配置，芝麻http代理 http://h.zhimaruanjian.com/
  api:
  type: socks
  BackUpCap: 50
  ActiveCap: 4
  ClearTime: 600
  TimeLimit: 175

imagePool:
  type: "off" # localPool / loliconPool

proxy:
  type: "off" # pyProxyPool / zhimaProxyPool

debug: # debug模式，只有以下群或qq号可以触发命令
  group:
    - 0
  uin:
    - 0

message-marker: # 是否禁用自动已读功能，true为禁用，false为不禁用
  disable: false

logLevel: info # 日志等级
```

</details>