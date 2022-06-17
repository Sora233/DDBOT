# DDBOT 模板介绍

DDBOT的模板大致与GO标准库`text/template`与相同，想深入了解请参考[官方文档](https://pkg.go.dev/text/template) 。

**注：DDBOT模板从`v1.0.5`版本开始支持**

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

<img src="https://user-images.githubusercontent.com/11474360/161202913-63f2064d-de49-4d54-b1d8-748b62b91294.jpg" width="300">

### 模板的高级功能

在模板中，所有的`{{ ..... }}`都有特殊意义，例如：

- 使用`{{ pic "uri" }}`发送图片：

*创建一个模板，编辑其内容为：*

```text
发送一张图片 {{ pic "https://i2.hdslb.com/bfs/face/0bd7082c8c9a14ef460e64d5f74ee439c16c0e88.jpg" }}
```

*该模板将发送为：*

<img src="https://user-images.githubusercontent.com/11474360/161203028-e53a4fe0-c8ca-4e3e-a098-45c1f7ea78c9.jpg" width="300">

在这个例子中，使用了一个模板函数`pic`，它可以用来发送一张图片。

- 使用`{{ cut }}`发送分段消息：

*创建一个模板，编辑其内容为：*

```text
一、啦啦啦
{{- cut -}}
二、啦啦啦
```

*该模板将发送为：*

<img src="https://user-images.githubusercontent.com/11474360/161203068-0d26024f-37e6-4a04-b88e-cfedd0235924.jpg" width="300">

在这个例子中，使用了一个模板函数`cut`，它可以用来切分消息。

*请注意括号内的短横线`-`符号，它的作用是控制前后换行符*。

- 使用模板变量：

`.`符号表示引用一个模板变量，**根据模板应用的场合，能够使用的模板变量也不尽相同**。

例如：当自定义`/签到`命令模板时，能够使用`.score`变量与`.success`变量。

*创建一个模板，编辑其内容为：*

```text
{{- reply .msg -}}
{{ if .success -}}
签到大大大大大成功！获得1只萝莉，当前拥有{{.score}}只萝莉
{{- else -}}
明天再来吧，当前拥有{{.score}}只萝莉
{{- end }}
```

*该模板将发送为：*

<img src="https://user-images.githubusercontent.com/11474360/161203356-84f28ac5-a883-4213-92ed-3c03ad2e524e.jpg" width="300">

这个例子还展示了**回复消息语法**：`{{- reply .msg -}}`以及模板的**条件控制语法**：`{{if}} {{else}} {{end}}`

## 通过模板创建自定义命令回复

得益于模板的高度定制化能力，DDBOT现在支持通过模板发送消息的自定义命令：

例如可以创建一个`/群主女装`命令，并让这个命令发送定义在模板内的群主女装照。

首先需要在配置文件`application.yaml`中定义命令：

```yaml
autoreply:
  group:
    command: [ "群命令1", "群命令2" ]
  private:
    command: [ "私聊命令1", "私聊命令2" ]
```

在上面这段配置中，自定义了两个群命令`/群命令1`和`/群命令2`，两个私聊命令`/私聊命令1`和`/私聊命令2`。

完成后需要创建对应的模板文件：

- `custom.command.group.群命令1.tmpl`
- `custom.command.group.群命令2.tmpl`
- `custom.command.private.私聊命令1.tmpl`
- `custom.command.private.私聊命令2.tmpl`

当触发`/群命令1`的时候，则会自动发送模板消息`custom.command.group.群命令1.tmpl`。

当触发`/私聊命令1`的时候，则会自动发送模板消息`custom.command.private.私聊命令1.tmpl`。

其他命令也遵守这个规则。

## 通过模板创建定时消息

DDBOT现在支持通过模板发送消息的定时命令：

首先需要在配置文件`application.yaml`中定义定时消息：

```yaml
cronjob:
  - cron: "* * * * *"
    templateName: "定时1"
    target:
      private: [ 123 ]
      group: [ ]
  - cron: "0 * * * *"
    templateName: "定时2"
    target:
      private: [ ]
      group: [ 456 ]
```

在上面这段配置中，自定义了两条定时消息，完成后需要创建对应的模板文件：

- `custom.cronjob.定时1.tmpl`
- `custom.cronjob.定时2.tmpl`

定时消息的定时条件使用`cron表达式`定义，可以在[工具网站](https://tool.lu/crontab/)上选择`类型：Linux`编辑和测试Cron表达式。

DDBOT使用五个字段的Cron表达式，这意味着最小的定时粒度为`1分钟`。

完成配置后，`定时1`会每分钟触发一次，触发时会私聊QQ号为123发送消息模板`custom.cronjob.定时1.tmpl`。

`定时2`会每小时触发一次，触发时会在QQ群456内发送消息模板`custom.cronjob.定时2.tmpl`。

## DDBOT新增的模板函数

- {{- cut -}}

用于发送分段消息，上面已介绍过

- {{ reply .msg }}

用于回复消息，上面已介绍过

- {{ prefix }}

引用配置中的command prefix，默认为`/`

- {{ pic "图片地址" }}

用于发送图片，支持`http/https`链接，以及本地路径。

图片格式支持 jpg / png / gif。

*如果路径是一个文件夹，则会在文件夹内随机选择一张图片。*

- {{ roll a b }}

在a ~ b范围内中随机一个数字，a b 返回值均为int64类型。

- {{ choose "a" "b" "c" }}

从传入的参数中随机返回一个，参数类型为string，支持变长参数。

*以下为v1.0.6新增*

- {{ at 123456 }}

发送@指定的qq号

- {{ icon 123456 }}

发送指定qq号的头像

*以下为v1.0.7新增*

- {{ hour }}

返回当前时间的小时数, 范围为[0, 23]，类型为int

- {{ minute }}

返回当前时间的分钟数，范围为[0, 59]，类型int

- {{ second }}

返回当前时间的秒数，范围为[0, 59]，类型为int

- {{ month }}

返回当前时间的月份，范围为[1,12]，类型为int

- {{ year }}

返回当前时间的年份，类型为int

- {{ day }}

返回当前时间为当月的第几天，类型为int

- {{ yearday }}

返回当前时间为当年的第几天，闰年范围为[1,366]，非闰年范围为[1,365]，类型为int

*以下为v1.0.8新增*

- 变量{{ .at_targets }}

自定义命令现在支持@成员， 可以通过 {{ .at_targets }}来获取本次命令触发时@的成员的QQ号。

该变量为数组类型，可以搭配`len`和`index`等命令使用，例如：

```text
{{- if gt (len .at_targets) 0 -}}
你@了{{ index .at_targets 0}}
{{- else -}}
请@TA使用命令喵
{{- end -}}
```

- {{ member_info .group_code .member_code }}

获取群成员信息，.at_targets中的qq号可以搭配这个命令来获取被@的成员的信息，例如：

```text
{{- if gt (len .at_targets) 0 -}}
	{{- $t := index .at_targets 0 -}}
	{{- $info := member_info .group_code $t -}}
	你@了{{- $info.name }}
{{ $info.name }} 的性别是 {{ if eq $info.gender 2 }}男生{{ else if eq $info.gender 1 }}女生{{else}}秘密{{end}} 喵
{{- else -}}
请@TA使用命令喵
{{- end -}}
```

返回结果中可以使用的信息有：

`name`：成员的群聊名字，如果没有设置，就是QQ资料名字

`gender`：QQ资料性别，2表示男生，1表示女生，0表示未公开

`permission`：QQ群权限，10表示群主，5表示管理员，1表示普通成员



## 当前支持的命令模板

命令通用模板变量：

| 模板变量        | 类型       | 含义                    | 备注            |
|-------------|----------|-----------------------|---------------|
| group_code  | int      | 本次命令触发的QQ群号码（私聊则为空）   |               |
| group_name  | string   | 本次命令触发的QQ群名称（私聊则为空）   |               |
| member_code | int      | 本次命令触发的成员QQ号          |               |
| member_name | string   | 本次命令触发的成员QQ名称         |               |
| cmd         | string   | 本次触发的命令名称             | 从v1.0.7版本开始支持 |
| args        | []string | 本次命令触发时附带的参数数组（只支持文字） | 从v1.0.7版本开始支持 |
| at_targets  | []int64  | 本次命令触发时附带的@成员的QQ号码    | 从v1.0.8版本开始支持 |

- /签到

模板名：`command.group.checkin.tmpl`

| 模板变量    | 类型   | 含义                             |
|---------|------|--------------------------------|
| success | bool | 表示本次签到是否成功，一天内只有第一次签到成功，后续签到失败 |
| score   | int  | 表示目前拥有的签到分数                    |

<details>
  <summary>默认模板</summary>

```text
{{ reply .msg }}{{if .success}}签到成功！获得1积分，当前积分为{{.score}}{{else}}明天再来吧，当前积分为{{.score}}{{end}}
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
{{- cut -}}
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

- /lsp

模板名：`command.group.lsp.tmpl`

| 模板变量 | 类型  | 含义  |
|------|-----|-----|
| 无    |     ||

<details>
  <summary>默认模板</summary>

```text
{{ reply .msg -}}
LSP竟然是你
```

</details>

- /ping(私聊)

模板名：`command.private.ping.tmpl`

| 模板变量 | 类型  | 含义  |
|------|-----|-----|
| 无    |     ||

<details>
  <summary>默认模板</summary>

```text
pong
```

</details>

## 当前支持的推送模板

- b站直播推送

模板名：`notify.group.bilibili.live.tmpl`

| 模板变量   | 类型     | 含义          |
|--------|--------|-------------|
| living | bool   | 是否正在直播      |
| name   | string | 主播昵称        |
| title  | string | 直播标题        |
| url    | string | 直播间链接       |
| cover  | string | 直播间封面或者主播头像 |

<details>
  <summary>默认模板</summary>

```text
{{ if .living -}}
{{ .name }}正在直播【{{ .title }}】
{{ .url -}}
{{ pic .cover "[封面]" }}
{{- else -}}
{{ .name }}直播结束了
{{ pic .cover "[封面]" }}
{{- end -}}
```

</details>

- ACFUN站直播推送

模板名：`notify.group.acfun.live.tmpl`

| 模板变量   | 类型     | 含义          |
|--------|--------|-------------|
| living | bool   | 是否正在直播      |
| name   | string | 主播昵称        |
| title  | string | 直播标题        |
| url    | string | 直播间链接       |
| cover  | string | 直播间封面或者主播头像 |

<details>
  <summary>默认模板</summary>

```text
{{ if .living -}}
ACFUN-{{ .name }}正在直播【{{ .title }}】
{{ .url -}}
{{ pic .cover "[封面]" }}
{{- else -}}
ACFUN-{{ .name }}直播结束了
{{ pic .cover "[封面]" }}
{{- end -}}
```

</details>

- 斗鱼直播推送

模板名：`notify.group.douyu.live.tmpl`

| 模板变量   | 类型     | 含义          |
|--------|--------|-------------|
| living | bool   | 是否正在直播      |
| name   | string | 主播昵称        |
| title  | string | 直播标题        |
| url    | string | 直播间链接       |
| cover  | string | 直播间封面或者主播头像 |

<details>
  <summary>默认模板</summary>

```text
{{ if .living -}}
斗鱼-{{ .name }}正在直播【{{ .title }}】
{{ .url -}}
{{ pic .cover "[封面]" }}
{{- else -}}
斗鱼-{{ .name }}直播结束了
{{ pic .cover "[封面]" }}
{{- end -}}
```

</details>

- 虎牙直播推送

模板名：`notify.group.huya.live.tmpl`

| 模板变量   | 类型     | 含义          |
|--------|--------|-------------|
| living | bool   | 是否正在直播      |
| name   | string | 主播昵称        |
| title  | string | 直播标题        |
| url    | string | 直播间链接       |
| cover  | string | 直播间封面或者主播头像 |

<details>
  <summary>默认模板</summary>

```text
{{ if .living -}}
虎牙-{{ .name }}正在直播【{{ .title }}】
{{ .url -}}
{{ pic .cover "[封面]" }}
{{- else -}}
虎牙-{{ .name }}直播结束了
{{ pic .cover "[封面]" }}
{{- end -}}
```

</details>

## 当前支持的事件模板

- 有新成员加入群

模板名：`trigger.group.member_in.tmpl`

| 模板变量        | 类型     | 含义        |
|-------------|--------|-----------|
| group_code  | int64  | 群号码       |
| group_name  | string | 群名称       |
| member_code | int64  | 新加入的成员QQ号 |
| member_name | string | 新加入的成员昵称  |

<details>
  <summary>默认模板</summary>

*该模板默认为空，即不发送消息*

```text
```

</details>

- 有成员退出群

模板名：`trigger.group.member_out.tmpl`

| 模板变量        | 类型     | 含义       |
|-------------|--------|----------|
| group_code  | int64  | 群号码      |
| group_name  | string | 群名称      |
| member_code | int64  | 退出的成员QQ号 |
| member_name | string | 退出的成员昵称  |

<details>
  <summary>默认模板</summary>

*该模板默认为空，即不发送消息*

```text
```

</details>

- bot添加新好友

模板名：`trigger.private.new_friend_added.tmpl`

| 模板变量                 | 类型     | 含义               |
|----------------------|--------|------------------|
| member_code          | int64  | 添加的好友QQ号         |
| member_name          | string | 添加的好友QQ昵称        |
| .command.HelpCommand | string | 帮助命令名称，默认是`help` |

<details>
  <summary>默认模板</summary>

```text
阁下的好友请求已通过，请使用<{{ prefix .command.HelpCommand }}>(不含括号)查看帮助，然后在群成员页面邀请bot加群（bot不会主动加群）。
```

</details>

- bot接受加群邀请

模板名：`trigger.private.group_invited.tmpl`

| 模板变量        | 类型     | 含义       |
|-------------|--------|----------|
| group_code  | int64  | 邀请加入的群号码 |
| group_name  | string | 邀请加入的群名称 |
| member_code | int64  | 邀请人的QQ号  |
| member_name | string | 邀请人的QQ昵称 |

<details>
  <summary>默认模板</summary>

```text
阁下的群邀请已通过，基于对阁下的信任，阁下已获得本bot在群【{{ .group_name }}】的控制权限，相信阁下不会滥用本bot。
```

</details>

