# 部署指南

首先需要获取DDBOT程序，可以选择下载预编译的版本，也可以选择从源码编译。

如果您不清楚如何进行编译，请选择预编译版本。

## 预编译版本

*预编译版本均为nocv版本，即无法正常使用face命令（本来也没啥人用这个命令）。*

从 [Release](https://github.com/Sora233/DDBOT/releases) 中下载适合您的版本

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

## 配置

### 设置管理员

初次运行之前推荐设置一个管理员帐号，否则您可能会失去对bot的控制权。

使用下面的命令设置：

```shell
./DDBOT --set-admin 你的QQ号码
```

### device.json

device.json是运行时使用的设备信息（可以理解为伪装的手机型号），应尽量使用同一个，否则可能会触发安全机制无法登陆。

如果检测到不存在，该文件会自动生成，正常情况无需关心。

### application.yaml

DDBOT运行时的配置文件，如果检测到不存在，会生成一个最小配置。

#### 最小配置

测试时推荐使用扫码登陆，即不指定帐号和密码（注意：掉线或重启后无法自动重连，仍需扫码，仅推荐测试使用）。

最小配置可以正常使用斗鱼、虎牙订阅功能。

如果可以直连油管，那也可以正常使用YTB订阅，如果无法直连油管，则需要配置代理功能，YTB订阅才能正常工作。

如果要使用b站订阅功能，需要填入下面bilibili的SESSDATA和bili_jct两个配置，请参考下面的说明。

```yaml
bot:
  account:  # 你的qq号，不填则使用扫码登陆
  password: # 你的qq密码

# b站登陆后的cookie字段，从cookie中找到这两个填进去，如果不会请百度搜索如何查看网站cookies
# 请注意，bot将使用您b站帐号的以下功能，建议使用新注册的小号：
# 关注用户 / 取消关注用户 / 查看关注列表
# 警告：
# SESSDATA和bili_jct等价于您的帐号凭证
# 请绝对不要透露给他人，更不能上传至Github等公开平台
# 否则将导致您的帐号被盗
bilibili:
  SESSDATA:
  bili_jct:
  interval: 15s

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

# b站登陆后的cookie字段，从cookie中找到这两个填进去
# 请注意，bot将使用您b站帐号的以下功能，建议使用新注册的小号：
# 关注用户 / 取消关注用户 / 查看关注列表
#
# 警告：
# SESSDATA和bili_jct等价于您的帐号凭证
# 请绝对不要透露给他人，更不能上传至Github等公开平台
# 否则将导致您的帐号被盗
bilibili:
  SESSDATA: ""
  bili_jct: ""
  interval: 30s # 直播状态和动态检测间隔，过快可能导致ip被暂时封禁


# 用于涩图鉴定功能
# https://api.aliyun.com/#/?product=imageaudit
aliyun: # 阿里云鉴权，用于图片识别，如果为空则图片识别功能无法工作
  accessKeyID: # 鉴权的账号应该开通内容审核功能（该服务2021年3月25日开始收费）
  accessKeySecret:

localPool: # 图片功能，使用本地图库
  imageDir: # 本地路径

loliconPool: # 图片功能，使用api.lolicon.app图库
  apikey: # 看api.lolicon.app网站内说明
  cacheMin: 10
  cacheMax: 50

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

logLevel: info # 日志等级
```

</details>