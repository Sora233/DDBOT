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

- {{ weekday }}

返回当前时间是本周的第几天，1~7分别表示周一至周日，范围为[1,7]，类型为int

- 变量 `{{ .at_targets }}`

自定义命令现在支持@成员， 可以通过 {{ .at_targets }}来获取本次命令触发时@的成员的QQ号。

该变量为数组类型，可以搭配`len`和`index`等命令使用，例如：

```text
{{- if gt (len .at_targets) 0 -}}
你@了{{ index .at_targets 0}}
{{- else -}}
请@TA使用命令喵
{{- end -}}
```

- 获取群成员信息 `{{ member_info .group_code .member_code }}`

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

- 带权重的`choose`

现在`choose`的每个元素后面支持一个可选的权重参数（必须是大于0的整数，如果省略则默认为1），进行随机时，将尊重权重设置，例如：

```text
{{ choose "a"  "b" 1 "c" 5 }}
```

在上面的结果中，结果是"c"的概率将是"a"和"b"的5倍，"a"和"b"的概率相同（"a"的权重省略默认为1，b"的权重1也可以省略）。

数值上说，选中"c"的概率是 $\frac{5}{7}$，

选中"a"和"b"的概率分别是 $\frac{1}{7}$。

感觉上来说就像写了5次"c"。

- 完整文字参数 `{{ .full_args }}`

返回本次命令触发时附带的完整文字参数，与`{{.args}}`的区别是：它不是一个数组，而是包含文字和空格的完整参数。

例如，当触发命令`/test a b c`时：

`{{.args}}`的内容是一个数组`[]string{"a", "b", "c"}`，需要配合`index`等命令使用。

`{{ .full_args }}`的内容是"a b c"，它是一个字符串而非数组。

- 一组类型转换函数

<details>
  <summary>点击查看详情</summary>

- `{{ float64 123 }}`

转换参数为`float64`类型

- `{{ int 123 }}`

转换参数为`int`类型

*int在32位和64位下表现不一致*

- `{{ int64 123}}`

转换参数为`int64`类型

- `{{ toString 123 }}`

转换参数为`string`类型

</details>

- 一组数学函数

<details>
  <summary>点击查看详情</summary>

函数名字末尾带**f**的表示它返回值是`float64`，不带**f**的返回值为`int64`

- add / addf

`{{ add 1 2}}` 计算加法

- sub / subf

`{{ sub 1 2}}` 计算减法

- mul / mulf

`{{ mul 2 2 }}` 计算乘法

- div / divf

`{{ div 10 5 }}` 计算除法

- mod / modf

`{{ mod 10 5 }}` 计算余数

</details>

- 一组最值函数

<details>
  <summary>点击查看详情</summary>

函数名字末尾带**f**的表示它返回值是`float64`，不带**f**的返回值为`int64`

- max / maxf

`{{ max 1 2 3 4 5 }}` 返回最大值

- min / minf

`{{ min 1 2 3 4 5 }}` 返回最小值

</details>

- 一组哈希函数

<details>
  <summary>点击查看详情</summary>

- base64encode

`{{ base64encode "hello world" }}` base64的加密

- base64decode

`{{ base64decode "aGVsbG8gd29ybGQ=" }}` base64的解密

- md5sum

`{{ md5sum "hello world" }}` md5加密

- sha1sum

`{{ sha1sum "hello world" }}` sha1加密

- sha256sum

`{{ sha256sum "hellow world" }}` sha256sum加密

- adler32sum

`{{ adler32sum "hellow world" }}` adler32sum加密

- uuid

{{ uuid }} 生成一个UUID


</details>

- 一组字符串函数

<details>
  <summary>点击查看详情</summary>

- hasPrefix

`{{ hasPrefix "aaa" "aaabcd" }}` 检查字符串是否有指定前缀

- hasSuffix

`{{ hasSuffix "aaa" "bcdaaa" }}` 检查字符串是否有指定后缀

- contains

`{{ contains "aaa" "bcdaaabcd" }}` 检查字符串是否包含一个子字符串

- trim

`{{ trim "   aaa   " }}` 去掉字符串前后的空白字符

- trimSuffix

`{{ trimSuffix "aaa" "bcdaaa" }}` 去掉字符串的指定后缀，如果字符串没有指定后缀，则无效果

- trimPrefix

`{{ trimPrefix "aaa" "aaabcd" }}` 去掉字符串的指定前缀，如果字符串没有指定前缀，则无效果

- split

`{{ split " " "foo bar baz" }}` 按照指定字符串分割字符串，返回一个list

- join

`{{ join " " (list "foor" "bar" "baz") }}` 按照指定字符串拼接字符串，返回一个字符串

- trunc

`{{ trunc 2 "abcde" }}` 按照长度截取字符串，如果长度大于字符串，则返回整个字符串

- upper

`{{ upper "abc" }}` 把字符串中所有英文字母改成大写

- lower

`{{ lower "ABC" }}` 把字符串中所有英文字母改成小写

- title

`{{ title "hello world" }}` 把字符串中的英文单词改为首字母大写

- snakecase

`{{ snakecase "FirstName" }}` 把字符串中所有命名改成蛇形命名法

- camelcase

`{{ camelcase "first_name" }}` 把字符串中所有命名改成驼峰式命名法

</details>

- 一组默认值与空值函数

<details>
  <summary>点击查看详情</summary>

- empty

`{{ if empty .args }}empty{{else}}not empty{{end}}`  检查变量是否是空值

- nonEmpty

`{{ if nonEmpty .args }}not empty{{else}}empty{{end}}`  检查变量是否不是空值

- coalesce

返回第一个非空的值，支持变长参数，常用于设置默认值

```
{{- $name := "" -}}
{{- $name = coalesce $name "没有名字" -}}
{{ $name }}
```

- ternary

三元运算符的语法糖，当第三个参数为true时返回第一个参数，否则返回第二个参数

`{{ ternary "有参数" "没有参数" (nonEmpty .args)}}`

- all

检查参数是否全部非空，支持变长参数

`{{ if all "" 0 1}}all is not empty{{end}}`

- any

检查参数是否有一个非空，支持变长参数

`{{ if any "" 0 1}}someone is not empty{{end}}`

</details>

- 一组list函数

<details>
  <summary>点击查看详情</summary>

- list

`{{ list "a" "b" "c" }}` 创建一个list，返回创建的list

- append

向list末尾添加一个元素，返回新的list

```
{{ $old := list "a" "b" "c" }}
{{ $new := append $old "d" }}
```

- prepend

向list开头添加一个元素，返回新的list

```
{{ $old := list "a" "b" "c" }}
{{ $new := prepend $old "d" }}
```

- concat

连接一组list，支持变长参数，返回新的list

```
{{ $l1 := list "a" }}
{{ $l2 := list "b" }}
{{ $l3 := list "c" }}
{{ $new := concat $l1 $l2 $l3 }}
```

</details>

- 一组dict函数

<details>
  <summary>点击查看详情</summary>

- dict

`{{ dict "a" 1 "b" 2 "c" 3 }}` 创建一个dict，其中key必须是字符串类型，返回创建的dict：

- get

从dict中获取值

```
{{ $d := dict "a" 1 "b" 2 "c" 3 }}
{{ $value := get $d "a" }}
```

- set

设置dict的值，其中key必须是字符串类型，返回dict

```
{{ $d := dict "a" 1 "b" 2 "c" 3 }}
{{ set $d "a" 100 }}
```

- unset

删除dict的值，返回dict

```
{{ $d := dict "a" 1 "b" 2 "c" 3 }}
{{ unset $d "a" }}
```

- hasKey

检查dict中是否有指定key

```
{{ $d := dict "a" 1 "b" 2 "c" 3 }}
{{ if hasKey $d "a" }}Yes.{{ end }}
```

- merge

合并dict，支持变长参数，不会覆盖已经存在的key，如果需要覆盖，请使用`mergeOverwrite`

返回合并后的dict

```
{{ $d1 := dict "a" 1 "b" 2 "c" 3 }}
{{ $d2 := dict "a" 2 "b" 3 "c" 4 }}
{{ $d3 := merge $d1 $d2 }}
```

- mergeOverwrite

合并dict，支持变长参数，会覆盖已经存在的key

返回合并后的dict

```
{{ $d1 := dict "a" 1 "b" 2 "c" 3 }}
{{ $d2 := dict "a" 2 "b" 3 "c" 4 }}
{{ $d3 := mergeOverwrite $d1 $d2 }}
```

- pick

从dict中取出指定的key创建新dict，支持变长参数，返回新创建的dict

```
{{ $d := dict "a" 1 "b" 2 "c" 3 }}
{{ $d2 := pick $d "a" "b" }}
```

</details>

- json处理

json处理使用 [gjson](https://github.com/tidwall/gjson) 库实现，请参考对应文档。

```
{{- $data := `{"name":{"first":"Janet","last":"Prichard"},"age":47}` -}}
{{- $j := toGJson $data -}}
{{- $name := ($j.Get "name.first").String -}}
{{- $age := ($j.Get "age").Int -}}

Name is {{ $name }}.
Age is {{ $age }}.  
```

- http请求

目前仅支持get / post 请求。

`httpGet` 用于发送get请求。

`httpPostJson` 用于发送json格式的post请求

`httpPostForm` 用于发送表单格式的post请求

三个函数的用法相同：

支持两个参数，第一个参数为url，第二个参数为请求参数（可选），请求参数的格式为dict，可使用dict创建。

```
{{/* 不带参数的httpGet */}}
{{- $j := httpGet "https://httpbin.sora233.me/get" | toGJson -}}
机器人的IP是：{{ coalesce ($j.Get "origin").String "获取失败" }}
```

```
{{/* 带参数的httpGet */}}
{{- $name := (member_info .group_code .member_code).name -}}
{{- $j := httpGet "https://httpbin.sora233.me/get" (dict "name" $name) | toGJson -}}
你好，{{ ($j.Get "args.name.0").String }}！
```

**http请求的特殊参数**

http请求存在一些特殊的保留参数，用来控制http行为，这些参数不会真正发送出去，当前支持：

- DDBOT_REQ_DEBUG

会详细输出本次http请求的细节（可能含有隐私信息，请注意不要随便复制给别人），用于DEBUG

- DDBOT_REQ_USER_AGENT

设置本次http请求的user_agent字段，大部分情况下不用设置，使用默认设置即可，默认值为：

```text
Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36
```

- DDBOT_REQ_HEADER

设置http请求的header，可以设置为list，每一项的格式为"A=B"，表示设置header名字为A，值为B

- DDBOT_REQ_COOKIE **（警告：不正确使用可能导致ccokie泄漏）**

设置http请求的cookie，可以设置为list，每一项的格式为"A=B"，表示设置cookie名字为A，值为B

- DDBOT_REQ_PROXY

用于控制请求代理，可以是：

`prefer_mainland` 使用mainland代理（要求已经配置proxy_pool)

`prefer_oversea` 使用oversea代理（要求已经配置proxy_pool)

`prefer_none` 不使用代理

`prefer_any` 随机选择mainland/oversea

`直接的proxy地址（例如http://localhost:7890）` 直接使用指定的代理

特殊参数的例子：

```
{{- /* 特殊参数的httpGet */ -}}

{{- $d := dict -}}

{{- /* 设置DDBOT_REQ_DEBUG，展示http 详细信息*/ -}}
{{- $d = set $d "DDBOT_REQ_DEBUG" "1" -}}

{{- /* 设置DDBOT_REQ_USER_AGENT，自定义USER_AGENT */ -}}
{{- $d = set $d "DDBOT_REQ_USER_AGENT" "DDBOT TEMPLATE 100%" -}}

{{- /* 设置DDBOT_REQ_HEADER，添加http自定义header */ -}}
{{- $d = set $d "DDBOT_REQ_HEADER" (list "FROM_DDBOT=yes") -}}

{{- /* 设置DDBOT_REQ_COOKIE，添加http自定义cookie */ -}}
{{- $d = set $d "DDBOT_REQ_COOKIE" (list "DDBOT_COOKIE=cookie_value") -}}

{{- /* 设置DDBOT_REQ_PROXY，不使用代理 */ -}}
{{- $d = set $d "DDBOT_REQ_PROXY" "prefer_none" -}}

{{- $_ := httpGet "https://httpbin.sora233.me/get" $d -}}
请查看命令行内的详细输出
```

*以下为v1.0.9新增*

- 发送戳一戳消息（只支持群聊）

戳一戳QQ号123456的用户

```
{{ poke 123456 }}
```

- 获取bot的qq号码

```
{{ bot_uin }}
```

- 设置命令冷却时间cooldown

cooldown设置后，设定时间内只有第一次会返回true，后续函数将返回false。

cooldown函数可以跨模板设置。

cooldown函数第一个参数为时间单位，支持如下简写，数字可以自由更换：

- 500ms（表示500毫秒）
- 1s（表示1秒钟）
- 20m（表示20分钟）
- 1.5h（表示1.5小时）
- 2h45m（表示2小时45分钟）

**如果设置为0或者负数，将自动替换为5m（即5分钟）**

cooldown函数的后续参数为设置cooldown的关键字，不同关键字的cooldown无任何关联，相同关键字的cooldown在时间范围内只能触发一次。

例子：

使用模板名作为关键字，实现单个模板所有人共享cooldown，不同模板独立cooldown：

```
{{- if (cooldown "10s" .template_name) -}}
成功
{{- else -}}
失败，正在冷却
{{- end -}}
```

把`.member_code`加入关键字，实现每个人对每个模板有独立cooldown：

```
{{- if (cooldown "10s" .member_code .template_name) -}}
成功
{{- else -}}
失败，正在冷却
{{- end -}}
```

使用固定关键字，实现不同模板共享cooldown：

```
{{- if (cooldown "10s" "my_cooldown_keyword_1") -}}
成功
{{- else -}}
失败，正在冷却
{{- end -}}
```

- 读取本地文件`openFile`

**警告：该函数并不会对参数做安全检查，在任何情况下都绝对不要把用户输入作为函数参数。**

返回值为`[]byte` ，例子：

```
{{ $data := openFile "path/myfile" }}
```

- 强制退出当前模板`abort`

退出当前模板并且丢弃已经产生的内容，如果有参数，则会发送参数

例子：

```
{{- abort -}}
```

```
{{- abort "出现错误" -}}
```

```
{{- if eq 1 5 -}}
  {{- abort (printf "出现错误: %v居然等于%v" 1 5) -}}
{{- end -}}
```

abort也支持图片参数

```
{{ abort (pic "https://i2.hdslb.com/bfs/face/0bd7082c8c9a14ef460e64d5f74ee439c16c0e88.jpg" ) }}
```


- 结束处理当前模板`fin`

退出当前模板并且发送已经产生的内容，未被处理的模板代码将被跳过

```
这句话会输出
{{- fin -}}
这句话不会输出
```

## 当前支持的命令模板

命令通用模板变量：

| 模板变量          | 类型       | 含义                    | 备注            |
|---------------|----------|-----------------------|---------------|
| group_code    | int      | 本次命令触发的QQ群号码（私聊则为空）   |               |
| group_name    | string   | 本次命令触发的QQ群名称（私聊则为空）   |               |
| member_code   | int      | 本次命令触发的成员QQ号          |               |
| member_name   | string   | 本次命令触发的成员QQ名称         |               |
| cmd           | string   | 本次触发的命令名称             | 从v1.0.7版本开始支持 |
| args          | []string | 本次命令触发时附带的参数数组（只支持文字） | 从v1.0.7版本开始支持 |
| at_targets    | []int64  | 本次命令触发时附带的@成员的QQ号码    | 从v1.0.8版本开始支持 |
| full_args     | string   | 本次命令触发时附带的完整参数（只支持文字） | 从v1.0.8版本开始支持 |
| template_name | string   | 本次命令触发时的模板名字          | 从v1.0.9版本开始支持 |

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

*以下为v1.0.9新增*

- bot收到群戳一戳

模板名：`trigger.group.poke.tmpl`

| 模板变量          | 类型     | 含义           |
|---------------|--------|--------------|
| group_code    | int64  | 发生戳一戳的群号码    |
| group_name    | string | 发生戳一戳的群名称    |
| member_code   | int64  | 发送戳一戳的用户QQ号  |
| member_name   | string | 发送戳一戳的用户QQ昵称 |
| receiver_code | int64  | 被戳的用户QQ号     |
| receiver_name | string | 被戳的用户QQ昵称    |

注意群内所有的戳一戳消息都会受到，如果只想处理bot被戳的消息，需要使用`receiver_code`进行判断

<details>
  <summary>默认模板</summary>

*该模板默认为空，即不发送消息*

```text
```

</details>

- bot收到好友私聊戳一戳

模板名：`trigger.private.poke.tmpl`

| 模板变量        | 类型     | 含义          |
|-------------|--------|-------------|
| member_code | int64  | 发送戳一戳的用户QQ号 |
| member_name | string | 发送戳一戳的QQ昵称  |

<details>
  <summary>默认模板</summary>

*该模板默认为空，即不发送消息*

```text
```

</details>