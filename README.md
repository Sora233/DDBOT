# DDBOT

一个基于MiraiGO的多功能QQ群机器人。

-----

## **基本功能：**

- **B站直播/动态推送**
    - 让阁下在DD的时候不错过任何一场突击。
- **斗鱼直播推送**
    - 没什么用，主要用来看爽哥。
- **油管直播/视频推送** *New*
    - 支持推送预约直播信息及视频更新
- **人脸识别**
    - 主要用来玩，支持二次元人脸
- **倒放**
    - 主要用来玩
- **Roll**
    - 没什么用的roll点
- **签到**
    - 没什么用的签到
- **权限管理**
    - 可配置整个命令的启用和禁用，也可对单个用户配置命令权限，防止滥用。
- **帮助**
    - 输出一些没什么帮助的信息

<details>
  <summary>里命令</summary>

以下命令默认禁用，使用enable命令后才能使用

- **随机图片**
    - 由 [api.olicon.app](https://api.lolicon.app/#/) 提供
- **色图判定**
    - 由阿里云提供

</details>

### 用法示例

详细介绍及示例请查看：[详细示例](/EXAMPLE.md)

阁下可添加Demo机器人体验，QQ：1561991863

<img src="https://user-images.githubusercontent.com/11474360/108590360-150afa00-739e-11eb-86f7-77f68d845505.jpeg" width="300" height="450">

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

配置完成后将文件改名为application.yaml

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

## 已知问题

- 一些情况下无法正常识别QQ群管理员，属于MiraiGo问题，无法在本项目解决。

## 注意事项

- bot掉线无法重连时将自动退出，请自行实现保活机制
- bot使用 [buntdb](https://github.com/tidwall/buntdb) 作为embed database，会在当前目录生成文件`.lsp.db`
  ，删除该文件将导致bot恢复出厂设置，可以使用 [buntdb-cli](https://github.com/Sora233/buntdb-cli) 作为运维工具，但注意不要在bot运行的时候使用（buntdb不支持多写）。

## 敬告

- 请勿滥用
- 禁止商用

## 贡献

*Feel free to make your first pull request.*

想要为开源做一点微小的贡献但不会Go语言? [点我入门！](https://github.com/justjavac/free-programming-books-zh_CN#go)

发现问题或功能建议请到 [issues](https://github.com/Sora233/DDBOT/issues)

其他用法问题请到 [discussions](https://github.com/Sora233/DDBOT/discussions)