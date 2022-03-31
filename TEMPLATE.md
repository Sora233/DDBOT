# DDBOT 模板介绍

DDBOT的模板大致与GO标准库`text/template`与相同，想深入了解请参考[官方文档](https://pkg.go.dev/text/template) 。

## 配置

DDBOT默认不启用自定义模板，当需要使用时，需要修改配置文件`application.yaml`，在其中增加一段配置，并且重启bot：

```yaml
template:
  enable: true
```

如果配置成功，启动时将显示日志信息：`已启用模板`。

配置成功后，DDBOT将会自动创建一个`template`文件夹，用于存放所有的模板文件。

DDBOT会**监控**该文件夹目录，这意味着对模板的创建/修改/删除等操作，无需重启BOT即可生效，DDBOT会自动使用最新的模板。

## 认识模板

模板是`template`文件夹下的一个文件，后缀名为`.tmpl`。

它是一个文本文件，可以使用文本编辑器进行编辑，它的文件名即是模板的名字，它的内容即是模板的内容。

### 文字模板

*创建一个模板，编辑其内容为：*

```text
这是一段文字，也是一段模板
```

*该模板将发送为：*

XXXXXXXXXXX

### 模板的高级功能

在模板中，所有的`{{ ..... }}`都有特殊意义，例如：

- 使用`{{ pic "uri" }}`发送图片：

*创建一个模板，编辑其内容为：*

```text
发送一张图片 {{ pic "https://i2.hdslb.com/bfs/face/0bd7082c8c9a14ef460e64d5f74ee439c16c0e88.jpg" }}
```

*该模板将发送为：*

XXXXXX

在这个例子中，使用了一个模板函数`pic`，它可以用来发送一张图片。

- 使用`{{ cut }}`发送分段消息：

*创建一个模板，编辑其内容为：*

```text
一、啦啦啦
{{ cut }}
二、啦啦啦
```

*该模板将发送为：*

XXXXX

在这个例子中，使用了一个模板函数`cut`，它可以用来切分消息。

- 使用模板变量：

`.`符号表示引用一个模板变量，**根据模板应用的场合，能够使用的模板变量也不尽相同**。

例如：当自定义`/签到`命令模板时，能够使用`.score`变量与`.success`变量。

*创建一个模板，编辑其内容为：*

```text
{{ if .success }}
签到大大大大大成功！获得1只萝莉，当前拥有{{.score}}只萝莉
{{else}}
明天再来吧，当前拥有{{.score}}
{{end}}
```

*该模板将发送为：*

XXXXX

这个例子同时展示了模板的条件控制语法：`if else end`

## 当前支持的命令模板

- /签到

模板名：`command.group.checkin.tmpl`

| 模板变量    | 类型   | 含义                             |
|---------|------|--------------------------------|
| success | bool | 表示本次签到是否成功，一天内只有第一次签到成功，后续签到失败 |
| score   | int  | 表示目前拥有的签到分数                    |

<details>
  <summary>默认模板</summary>

```text
{{if .success}}签到成功！获得1积分，当前积分为{{.score}}{{else}}明天再来吧，当前积分为{{.score}}{{end}}
```

</details>

- /help （私聊版）

模板名：`command.private.help.tmpl`

| 模板变量 | 类型  | 含义  |
|------|-----|-----|
| 无    |     ||

<details>
  <summary>默认模板</summary>

```text
常见订阅用法：
以作者UID:97505为例
首先订阅直播信息：{{ prefix }}watch 97505
然后订阅动态信息：{{ prefix }}watch -t news 97505
由于通常动态内容较多，可以选择不推送转发的动态
{{ prefix }}config filter not_type 97505 转发
还可以选择开启直播推送时@全体成员：
{{ prefix }}config at_all 97505 on
以及开启下播推送：
{{ prefix }}config offline_notify 97505 on
BOT还支持更多功能，详细命令介绍请查看命令文档：
https://gitee.com/sora233/DDBOT/blob/master/EXAMPLE.md
使用时请把作者UID换成你需要的UID
当您完成所有配置后，可以使用{{ prefix }}silence命令，让bot专注于推送，在群内发言更少
{{ cut }}
B站专栏介绍：https://www.bilibili.com/read/cv10602230
如果您有任何疑问或者建议，请反馈到唯一指定交流群：755612788
```

</details>

- /help （群聊版）

模板名：`command.group.help.tmpl`

| 模板变量 | 类型  | 含义  |
|------|-----|-----|
| 无    |     ||

<details>
  <summary>默认模板</summary>

```text
DDBOT是一个多功能单推专用推送机器人，支持b站、斗鱼、油管、虎牙推送
```

</details>