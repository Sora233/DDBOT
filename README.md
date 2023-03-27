# DDBOT

[<img src="https://github.com/Sora233/DDBOT/actions/workflows/ci.yml/badge.svg"/>](https://github.com/Sora233/DDBOT/actions/workflows/ci.yml)
[<img src="https://coveralls.io/repos/github/Sora233/DDBOT/badge.svg?branch=master"/>](https://coveralls.io/github/Sora233/DDBOT)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/Sora233/DDBOT)

DDBOT是一个基于 [MiraiGO](https://github.com/Mrs4s/MiraiGo) 的QQ群推送框架， 内置支持b站直播/动态，斗鱼直播，YTB直播/预约直播，虎牙直播，ACFUN直播，微博动态，
也可以通过插件支持任何订阅源。

*DDBOT不是一个聊天机器人。*

[Bilibili专栏](https://www.bilibili.com/read/cv10602230)

-----

## 设计理念

制作bot的本意是为了减轻一些重复的工作负担，bot只会做好bot份内的工作：

- ddbot的交互被刻意设计成最小程度，正常交流时永远不必担心会误触ddbot。
- ddbot只有两种情况会主动发言，更新动态和直播，以及答复命令结果。

## **基本功能：**

- **B站直播/动态推送**
  - 让阁下在DD的时候不错过任何一场突击。
  - 支持按关键字过滤，只推送有关键字的动态。
  - 支持按动态类型过滤，例如：不推送转发的动态，只推送视频/专栏投稿，只推动带图片的动态等等。
- **斗鱼直播推送**
  - 没什么用，主要用来看爽哥。
- **油管直播/视频推送**
  - 支持推送预约直播信息及视频更新。
- **虎牙直播推送**
  - 不知道能看谁。
- **ACFUN直播推送**
  - 好像也有一些虚拟主播
- **微博动态推送**
- 支持自定义**插件**，可通过插件支持任意订阅来源
  - 需要写代码
- 可配置的 **@全体成员**
  - 只建议单推群开启。
- **倒放**
  - 主要用来玩。
- **Roll**
  - 没什么用的roll点。
- **签到**
  - 没什么用的签到。
- **权限管理**
  - 可配置整个命令的启用和禁用，也可对单个用户配置命令权限，防止滥用。
- **帮助**
  - 输出一些没什么帮助的信息。

<details>
  <summary>里命令</summary>

以下命令默认禁用，使用enable命令后才能使用

- **随机图片**
  - 由 [api.lolicon.app](https://api.lolicon.app/#/) 提供

</details>

### 推送效果

<img src="https://user-images.githubusercontent.com/11474360/111737379-78fbe200-88ba-11eb-9e7e-ecc9f2440dd8.jpg" width="300">

### 用法示例

详细介绍及示例请查看：[详细示例](/EXAMPLE.md)

~~阁下可添加官方Demo机器人体验~~

不再提供官方的公开BOT，你可以加入交流群申请使用群友搭建的BOT，也可以选择自己搭建。

## 使用与部署

对于普通用户，推荐您选择使用开放的官方Demo机器人。

您也可以选择私人部署，[详见部署指南](/INSTALL.md)。

私人部署的好处：

- 保护您的隐私，bot完全属于您，我无法得知您bot的任何信息（我甚至无法知道您部署了一个私人bot）
- 稳定的@全体成员功能
- 可定制BOT账号的头像、名字、签名
- 减轻我的服务器负担
- 很cool

如果您遇到任何问题，或者有任何建议，可以加入**唯一指定交流群：755612788**

## 最近更新

请参考[更新文档](/UPDATE.md)。

## 常见问题FAQ

提问前请先查看[FAQ文档](/FAQ.md)，如果仍然未能解决，请咨询唯一指定交流群。

## 增加推送来源 （为DDBOT编写插件）

DDBOT可以作为一个通用的QQ推送框架来使用。

您可以通过为DDBOT编写插件，DDBOT会为您完成大部分工作，您只需要实现少量代码，就能支持一个新的来源。

如果您对此有兴趣，请查看[框架文档](/FRAMEWORK.md) 。

## 自定义消息模板 & 自定义命令回复

DDBOT已实现消息模板功能，一些内置命令和推送可通过模板自定义格式。

同时支持自定义命令，自动回复模板内容。

详细介绍请看[模板文档](/TEMPLATE.md) 。

## 注意事项

- **bot只在群聊内工作，但命令可以私聊使用，以避免在群内刷屏**（少数次要娱乐命令暂不支持，详细列表请看用法指南）
- **建议bot秘密码设置足够强，同时不建议把bot设置为QQ群管理员，因为存在密码被恶意爆破的可能（包括但不限于盗号、广告等）**
- **您应当知道，bot账号可以人工登陆，请注意个人隐私**
- bot掉线无法重连时将自动退出，请自行实现保活机制
- bot使用 [buntdb](https://github.com/tidwall/buntdb) 作为embed database，会在当前目录生成文件`.lsp.db`
  ，删除该文件将导致bot恢复出厂设置，可以使用 [buntdb-cli](https://github.com/Sora233/buntdb-cli) 作为运维工具，但注意不要在bot运行的时候使用（buntdb不支持多写）

## 声明

- 您可以免费使用DDBOT进行其他商业活动，但不允许通过出租、出售DDBOT等方式进行商业活动。
- 如果您运营了私人部署的BOT，可以接受他人对您私人部署的BOT进行捐赠以帮助BOT运行，但该过程必须本着自愿的原则，不允许用BOT使用权来强制他人进行捐赠。
- 如果您使用了DDBOT的源代码，或者对DDBOT源代码进行修改，您应该用相同的开源许可（AGPL3.0）进行开源，并标明著作权。

## 贡献

*Feel free to make your first pull request.*

想要为开源做一点微小的贡献？

[Golang点我入门！](https://github.com/justjavac/free-programming-books-zh_CN#go)

您也可以选择点一下右上角的⭐星⭐

发现问题或功能建议请到 [issues](https://github.com/Sora233/DDBOT/issues)

其他用法问题请到**唯一指定交流群：755612788**

## 赞助

（排名按时间先后顺序）

| 赞助者          | 渠道  | 金额      |
|--------------|-----|---------|
| VE-H Maw     | 爱发电 | ￥30.00  |
| 饱受突击的3737民   | 爱发电 | ￥168.00 |
| 刀光流水         | 爱发电 | ￥5.00   |
| 爱发电用户_4QBx   | 爱发电 | ￥5.00   |
| XDMrSmile_鸟鸟 | 爱发电 | ￥120.00 |
| 别赋清颜         | 爱发电 | ￥300.00 |
| 赫尔博达         | 爱发电 | ￥50.00  |
| 冰王星          | 爱发电 | ￥5.00   |
| yzr6991      | 爱发电 | ￥30.00  |
| 么鱼           | 爱发电 | ￥5.00   |
| Darren今天社恐了嘛 | 爱发电 | ￥5.00   |
| Ed😓         | 爱发电 | ￥5.00   |
| 殿九           | 爱发电 | ￥30.00  |
| ACR          | 爱发电 | ￥50.00  |
| 南宫           | 爱发电 | ￥30.00  |
| 宇            | QQ  | ￥100.00 |
| 殿九           | QQ  | ￥50.00  |
| 梦幻七罪         | 爱发电 | ￥5.00   |
| 电子鸽子         | 爱发电 | ￥30.00  |
| 星落泪羽         | 爱发电 | ￥30.00  |

## 鸣谢

> [Goland](https://www.jetbrains.com/go/) 是一个非常适合Gopher的智能IDE，它极大地提高了开发人员的效率。

特别感谢 [JetBrains](https://jb.gg/OpenSource) 为本项目提供免费的 [Goland](https://www.jetbrains.com/go/) 等一系列IDE的授权

[<img src="https://user-images.githubusercontent.com/11474360/112592917-baa00600-8e41-11eb-9da4-ecb53bb3c2fa.png" width="200"/>](https://jb.gg/OpenSource)

## DDBOT:star:趋势图

[![Stargazers over time](https://starchart.cc/Sora233/DDBOT.svg)](https://starchart.cc/Sora233/DDBOT)
