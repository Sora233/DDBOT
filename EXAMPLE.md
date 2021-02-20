# 命令详细介绍及用法示例

-----

*本介绍并不要求您拥有编程能力，但如果有，您会非常容易理解（因为命令的格式被设计为贴近shell命令）。*

## 命令基本格式

**bot只在群聊内工作，私聊命令无效**

```
/主命令 [命令选项 [选项参数] ...] [主命令参数 ...]
```

#### 例子1：我买了海盗床，我要看爽哥，如何使用watch命令订阅爽哥b站直播信息？

- 进入爽哥b站个人[空间主页](https://space.bilibili.com/1472906636)

- 找到爽哥的UID`1472906636`

```shell
/watch --site bilibili --type live 1472906636
```

在这个例子中，主命令为`watch`，它包含两个命令选项，分别是`--site`和`--type`，两个命令选项的参数分别是`bilibili`和`live`，有一个主命令参`1472906636`。

`--site`用来指定订阅网站，`--type`用来指定订阅类型，所以上面的命令表示：订阅**b站**的UID为**1472906636**的用户的**直播**信息。

命令选项有长和短两种模式，两种模式效果相同，这个例子中使用的是长模式，它的格式为`--`加上单词，短模式格式为`-`加上单个字母。

大多数长模式都可以用`-`加上单词的首字母简写成短模式。

例如`--site`简写成`-s`，`--type`简写成`-t`，那么命令可以写成：

```shell
/watch -s bilibili -t live 1472906636
```

更进一步，对于`watch`命令来说，`-s`的默认参数是`bilibili`，`-t`的默认参数是`live`。也就是说，如果不指定`-s`，那么默认网站是`bilibili`；如果不指定`-t`，那么默认类型是`live`
，所以命令进一步简写成：

```shell
/watch 1472906636
```

## 命令帮助信息

您可以使用选项参数`-h/--help`查看命令的帮助信息，例如输入`/watch -h`或者`/watch --help` （长模式和短模式），会输出：

```
Usage: watch <id>

Arguments:
  <id>

Flags:
  -h, --help               Show context-sensitive help.
  -s, --site="bilibili"    bilibili / douyu / youtube
  -t, --type="live"        news / live
```

`Flags`里展示了`watch`命令的三个命令参数，分别是`-s/--site`、`-t/--type`、`-h/--help`，其中`--site="bilibili"`表示`-s/--site`
的默认参数是bilibili，而后面`bilibili / douyu / youtube`部分则表示--site可以指定为这三个网站。

`-t/--type`同理。

`Arguments: <id>`表示watch命令需要一个主命令参数<id>，在例子1中为爽哥的UID`1472906636`。

一些例子：

```shell
// 订阅b站UID为2的用户的直播信息
/watch 2
```

```shell
// 订阅b站UID为2的用户的动态信息
/watch -t news 2
```

```shell
// 订阅YTB乙女音
```

DDBOT的命令格式皆遵守这个规则。

## 命令详细介绍

### TODO
