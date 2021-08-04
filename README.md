# DDBOT

[<img src="https://github.com/Sora233/DDBOT/actions/workflows/ci.yml/badge.svg"/>](https://github.com/Sora233/DDBOT/actions/workflows/ci.yml)
[<img src="https://coveralls.io/repos/github/Sora233/DDBOT/badge.svg?branch=master"/>](https://coveralls.io/github/Sora233/DDBOT)

一个基于MiraiGO的多功能QQ群机器人。

[Bilibili专栏](https://www.bilibili.com/read/cv10602230)

-----

## 设计理念

制作bot的本意是为了减轻一些重复的工作负担，bot只会做好bot份内的工作：

- ddbot的交互被刻意设计成最小程度，正常交流时永远不必担心会误触ddbot。
- ddbot只有两种情况会主动发言，更新动态和直播，以及答复命令结果。

## **基本功能：**

- **B站直播/动态推送**
    - 让阁下在DD的时候不错过任何一场突击。
- **斗鱼直播推送**
    - 没什么用，主要用来看爽哥。
- **油管直播/视频推送**
    - 支持推送预约直播信息及视频更新。
- **虎牙直播推送** *新增*
    - 不知道能看谁。
- 可配置的 **@全体成员** *新增*
    - 只建议单推群开启。
- **人脸识别**
    - 主要用来玩，支持二次元人脸。
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
    - 由 [api.olicon.app](https://api.lolicon.app/#/) 提供
- **色图判定**
    - 由阿里云提供
    - **注意：阿里云该服务2021年3月25日开始收费**

</details>

### 推送效果

<img src="https://user-images.githubusercontent.com/11474360/111737379-78fbe200-88ba-11eb-9e7e-ecc9f2440dd8.jpg" width="300">

### 用法示例

详细介绍及示例请查看：[详细示例](/EXAMPLE.md)

阁下可添加Demo机器人体验，QQ号：

- 368236249 （二号机）
- 1561991863 （初号机）

<details>
<summary>点此扫码二号机</summary>
<img src="https://user-images.githubusercontent.com/11474360/122684719-a8afe280-d239-11eb-9089-b8ce6613c819.jpg" width="300" height="450">
</details>

<details>
<summary>点此扫码初号机</summary>
<img src="https://user-images.githubusercontent.com/11474360/108590360-150afa00-739e-11eb-86f7-77f68d845505.jpeg" width="300" height="450">
</details>

推荐您优先选择二号机，由于目前初号机负载较高。

**尝试同时使用多个Demo机器人会导致您被暂时加入黑名单**

## 使用与部署

对于普通用户，推荐您选择使用开放的Demo机器人。

您也可以选择私人部署，[详见部署指南](/INSTALL.md)。

私人部署的好处：

- 保护您的隐私，bot完全属于您，我无法得知您bot的任何信息（我甚至无法知道您部署了一个私人bot）
- 稳定的@全体成员功能
- 可定制BOT账号的头像、名字、签名
- 减轻我的服务器负担
- 很cool

如果您遇到任何问题，或者有任何建议，可以加入**唯一指定交流群：755612788**

## **最近更新**

- 增加了管理员命令/quit，用于退出群聊，或清除未正常退出的群的数据。
- 重构buntdb支持嵌套事务，现在被禁言时不会再尝试推送了
- 增加了b站使用账号密码登陆，可以不用再手动填cookie了（非常容易出错）
- 增加了禁用单个命令的管理员命令，对所有群组生效
- 改善了初次使用的体验（增加扫码登陆，自动生成device.json和application.yaml）
- 增加了config配置下播推送，直播间标题推送，开启@全体成员，b站动态推送过滤器。
- 增加了config命令，目前支持配置推送同时@成员（实验阶段，并不保证稳定，目前只支持直播推送）。
- 增加了私聊的grant命令，现在可以通过私聊配置qq群内的群友命令权限了。
- 更新了list命令，现在list命令会展示所有网站的所有订阅，不必再指定-s和-t了。
- 增加了私聊的list命令，现在可以通过私聊查看订阅列表了。
- 增加了虎牙直播的推送。
- 增加了私聊的enable和disable命令，现在可以通过私聊配置qq群内的命令了。
- 增加了私聊的watch和unwatch命令，现在可以通过私聊配置qq群内的订阅了（注意仍然是qq群推送）。
- 更换船新b站监控方案，预计单帐号可支持1000订阅，最大延迟30秒。

## 常见问题FAQ

### 为什么我设置了@全体成员却没效果？

- 我使用的是官方bot：

首先@全体成员无法在官方demo bot上正常使用，如果您使用的是官方bot，该问题是已知的，请使用@特定用户作为替代。
（这是因为一个账号一天只能@全体成员十次，该次数在所有群内共享，因为官方bot用户众多，很容易把次数用尽，用尽后即无法再@全体成员）

- 我使用的是私人部署的bot：

请检查是否把bot设为qq群管理员，检查`/config at_all`配置是否设置为`on`，如果确认配置无问题，请把该问题提交给开发者。

### 为什么我在群里发指令它没反应？

- 我使用的是官方bot：

请确认该命令是正确的，以及该命令没有被管理员禁用，最后确认bot处于正常运行的状态（可以使用/help测试有没有反应），如果确认后仍然没有解决，请把该问题提交给开发者。

- 我使用的是私人部署的bot：

请检查后台有无反应，如果后台有发送消息失败的日志但是qq没有收到，大概率是qq帐号处于风控帐号，需要人工登陆帐号正常使用一段时间以解除风控（注意不要再继续使用BOT了，否则会导致被封号）。

如果后台没有消息发送失败的日志，但是qq也没有收到，该情况小概率触发，一般交替使用别的命令数次，以及等待一段时间可以解决。

如果后台无法接受消息，请用手机登陆bot帐号，在设备管理页面把`mirai`名字的设备全部删除，等待一段时间重新登陆。

### 为什么我无法订阅b站动态/直播？

- 我使用的是官方bot：

请把该问题提交给开发者。

- 我使用的是私人部署的bot：

请检查b站配置完整，需要配置bilibili一栏的account和password，同时检查运行后日志内有如下或类似内容，否则请看ERROR日志：

```
AddVerifyOption 使用帐号刷新cookie
login to fresh cookie
login success
使用cookieInfo设置 bili_jct
使用cookieInfo设置 SESSDATA
```

### 为什么我订阅了动态/直播没有推送？

- 我使用的是官方bot：

检查是否使用`/config filter`设置了动态过滤。

- 我使用的是私人部署的bot

检查bot日志是否有该条`notify`，以及是否有消息发送失败。

### 为什么我私人部署的bot修改了配置后运行时闪退？

因为配置格式填写的不对，一些填写时注意事项：

- 填写时请把井号及后面的内容删除
- 所有引号、冒号均需要使用英文（半角）格式
- 冒号后需要加一个英文空格

### 为什么我私人部署的bot会出现发送群消息失败？

如果您是刚开始使用BOT，大概是因为qq账号被风控导致，建议使用bot挂机3-7天（期间请不要使用bot的功能，否则可能导致封号），即可解除风控。

## 注意事项

- **bot只在群聊内工作，但命令可以私聊使用，以避免在群内刷屏**（少数次要娱乐命令暂不支持，详细列表请看用法指南）
- **建议bot秘密码设置足够强，同时不建议把bot设置为QQ群管理员，因为存在密码被恶意爆破的可能（包括但不限于盗号、广告等）**
- **您应当知道，bot账号可以人工登陆，请注意个人隐私**
- bot掉线无法重连时将自动退出，请自行实现保活机制
- bot使用 [buntdb](https://github.com/tidwall/buntdb) 作为embed database，会在当前目录生成文件`.lsp.db`
  ，删除该文件将导致bot恢复出厂设置，可以使用 [buntdb-cli](https://github.com/Sora233/buntdb-cli) 作为运维工具，但注意不要在bot运行的时候使用（buntdb不支持多写）

## 敬告

- 请勿滥用
- 禁止商用

## 贡献

*Feel free to make your first pull request.*

想要为开源做一点微小的贡献？

[Golang点我入门！](https://github.com/justjavac/free-programming-books-zh_CN#go)

您也可以选择点一下右上角的⭐星⭐

发现问题或功能建议请到 [issues](https://github.com/Sora233/DDBOT/issues)

其他用法问题请到 [discussions](https://github.com/Sora233/DDBOT/discussions)

## 鸣谢

> [Goland](https://www.jetbrains.com/go/) 是一个非常适合Gopher的智能IDE，它极大地提高了开发人员的效率。

特别感谢 [JetBrains](https://jb.gg/OpenSource) 为本项目提供免费的 [Goland](https://www.jetbrains.com/go/) 等一系列IDE的授权

[<img src="https://user-images.githubusercontent.com/11474360/112592917-baa00600-8e41-11eb-9da4-ecb53bb3c2fa.png" width="200"/>](https://jb.gg/OpenSource)

## DDBOT:star:趋势图

[![Stargazers over time](https://starchart.cc/Sora233/DDBOT.svg)](https://starchart.cc/Sora233/DDBOT)
