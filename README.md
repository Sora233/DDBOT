# Sora233-MiraiGo

一个基于MiraiGO的多功能QQ群机器人。

-----

## **基本功能：**

- **B站直播/动态推送**
    - 让阁下在DD的时候不错过任何一场突击。
    - 命令
        - /watch -s bilibili
- **斗鱼直播推送**
    - 没什么用，主要用来看爽哥。
    - 命令
        - /watch -s douyu
- **取消订阅**
    - 没什么用
    - 命令
        - /unwatch
- **查看订阅**
    - 主要用来刷屏
    - 命令
        - /list
- **人脸识别**
    - 主要用来玩
    - 支持二次元人脸
    - 命令
        - /face (电脑使用/face [图片] 或者 回复图片消息+/face触发)
    - *recall不高*
- **Roll**
    - 没什么用的roll点
    - 命令
        - /roll
- **签到**
    - 没什么用的签到
    - 命令
        - /签到
- **权限管理**
    - 可配置整个命令的启用和禁用，也可对单个用户配置命令权限，防止滥用。
    - 命令
        - /enable
        - /disable
        - /grant
- **帮助**
    - 输出一些没什么帮助的信息
    - 命令
        - /help
        - /about
    - 使用命令 -h可查看该命令的帮助信息，例如
        - /watch -h
        - /grant -h

<details>
  <summary>里命令</summary>

以下命令默认禁用

- **随机图片**
    - 由 [api.olicon.app](https://api.lolicon.app/#/) 提供
    - 命令
        - /色图
- **色图判定**
    - 由阿里云提供
    - 命令 ic (该命令启用后自动对图片触发，无需输入命令)

</details>

阁下可添加Demo机器人体验 1561991863

## 使用

自行编译

## 编译

go >= 1.15

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

模板为application.yaml.example，好后将文件改名为application.yaml

```yaml
bot:
  account: # bot账号
  password: # bot密码

moderatecontent:
  apikey: # 已弃用

# https://api.aliyun.com/#/?product=imageaudit
aliyun: # 阿里云鉴权，用于图片识别，如果为空则图片识别功能无法工作
  accessKeyID: # 鉴权的账号应该开通内容审核功能（目前免费）
  accessKeySecret:

localPool: # 图片功能，使用本地图库
  imageDir: # 本地路径

loliconPool: # 图片功能，使用api.lolicon.app图库
  apikey: # 看api.lolicon.app网站内说明

pyProxyPool: # 代理池配置，py代理池 https://github.com/jhao104/proxy_pool
  host: http://127.0.0.1:5010

localProxyPool: # 代理池配置，固定代理
  proxy:
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
  type: off # localPool / loliconPool

proxy:
  type: off # pyProxyPool / zhimaProxyPool

debug: # debug模式，只有以下群或qq号可以触发命令
  group:
    - 0
  uin:
    - 0

logLevel: info # 日志等级
```

## 敬告

- 请勿滥用
- 禁止商用

## 贡献

发现问题或功能建议请到 [issues](https://github.com/Sora233/Sora233-MiraiGo/issues)

其他用法问题请到 [discussions](https://github.com/Sora233/Sora233-MiraiGo/discussions)