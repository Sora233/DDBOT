# 部署指南

首先需要获取DDBOT程序，可以选择下载预编译的版本，也可以选择从源码编译。

如果您不清楚如何进行编译，请选择预编译版本。

## 预编译版本

*预编译版本均为nocv版本，即无法正常使用face命令。*

从 [Release](https://github.com/Sora233/DDBOT/releases) 中下载适合您的版本

## 从源码编译

go >= 1.16

编译可选择是否启动gocv，如果不需要使用face命令，推荐禁用gocv。（因为安装opencv容易出现问题）

- ### 启用gocv

请先安装 [gocv](https://github.com/hybridgroup/gocv)

```
go build
```

- ### 禁用gocv

禁用gocv将导致face命令无法正常工作

```
go mod edit -droprequire gocv.io/x/gocv
go build -tags nocv
```

## 配置

### device.json

device.json是运行时使用的设备信息，应尽量使用同一个，否则可能会触发安全机制无法登陆。

第一次运行时请运行下面的命令生成device.json。

```
./DDBOT --generate-device
```

### aplication.json

DDBOT运行时的配置文件， 配置完成后将文件改名为application.yaml，和DDBOT放在一起。

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


moderatecontent:
  apikey: # 已弃用

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
