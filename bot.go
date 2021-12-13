package DDBOT

import (
	"fmt"
	"github.com/Sora233/DDBOT/lsp"
	"github.com/Sora233/MiraiGo-Template/bot"
	"github.com/Sora233/MiraiGo-Template/config"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	_ "github.com/Sora233/DDBOT/logging"
	_ "github.com/Sora233/DDBOT/lsp/acfun"
	_ "github.com/Sora233/DDBOT/lsp/douyu"
	_ "github.com/Sora233/DDBOT/lsp/huya"
	_ "github.com/Sora233/DDBOT/lsp/twitcasting"
	_ "github.com/Sora233/DDBOT/lsp/weibo"
	_ "github.com/Sora233/DDBOT/lsp/youtube"
	_ "github.com/Sora233/DDBOT/miraigo-logging"
	_ "github.com/Sora233/DDBOT/msg-marker"
)

// SetUpLog 使用默认的日志格式配置，会写入到logs文件夹内，日志会保留七天
func SetUpLog() {
	writer, err := rotatelogs.New(
		path.Join("logs", "%Y-%m-%d.log"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		logrus.WithError(err).Error("unable to write logs")
		return
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
	})
	logrus.AddHook(lfshook.NewHook(writer, &logrus.TextFormatter{
		FullTimestamp:    true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
		ForceQuote:       true,
	}))
}

// Run 启动bot，这个函数会阻塞直到收到退出信号
func Run() {
	if fi, err := os.Stat("device.json"); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("警告：没有检测到device.json，正在生成，如果是第一次运行，可忽略")
			bot.GenRandomDevice()
		} else {
			fmt.Printf("检查device.json文件失败 - %v", err)
			os.Exit(1)
		}
	} else {
		if fi.IsDir() {
			fmt.Println("检测到device.json，但目标是一个文件夹！请手动确认并删除该文件夹！")
			os.Exit(1)
		} else {
			fmt.Println("检测到device.json，使用存在的device.json")
		}
	}

	if fi, err := os.Stat("application.yaml"); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("警告：没有检测到配置文件application.yaml，正在生成，如果是第一次运行，可忽略")
			if err := ioutil.WriteFile("application.yaml", []byte(exampleConfig), 0755); err != nil {
				fmt.Printf("application.yaml生成失败 - %v\n", err)
				os.Exit(1)
			} else {
				fmt.Println("最小配置application.yaml已生成，请按需修改，如需高级配置请查看帮助文档")
			}
		} else {
			panic(fmt.Sprintf("检查application.yaml文件失败 - %v", err))
		}
	} else {
		if fi.IsDir() {
			fmt.Printf("检测到application.yaml，但目标是一个文件夹！请手动确认并删除该文件夹！")
			os.Exit(1)
		} else {
			fmt.Println("检测到application.yaml，使用存在的application.yaml")
		}
	}

	config.Init()

	// 快速初始化
	bot.Init()

	// 初始化 Modules
	bot.StartService()

	// 使用协议
	// 不同协议可能会有部分功能无法使用
	// 在登陆前切换协议
	bot.UseProtocol(bot.AndroidPhone)

	// 登录
	bot.Login()

	// 刷新好友列表，群列表
	bot.RefreshList()

	lsp.Instance.PostStart(bot.Instance)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	bot.Stop()
}

var exampleConfig = func() string {
	s := `
# 注意，填写时请把井号及后面的内容删除，并且冒号后需要加一个空格
bot:
  account:  # 你bot的qq号，不填则使用扫码登陆
  password: # 你bot的qq密码
  onJoinGroup: 
    rename: "【bot】"  # BOT进群后自动改名，默认改名为“【bot】”，如果留空则不自动改名

# b站相关的功能需要一个b站账号，建议使用小号
# bot将使用您b站帐号的以下功能：
# 关注用户 / 取消关注用户 / 查看关注列表
# 请注意，订阅一个账号后，此处使用的b站账号将自动关注该账号
bilibili:
  account:  # 你的b站账号 
  password: # 你的b站密码
  interval: 25s

# 参阅 https://apiv2-doc.twitcasting.tv/#registration
twitcasting:
  clientId:  abc
  clientSecret: xyz
  # 为防止风控，可选择性广播以下元素
  broadcaster:
	title: false
	created: true
	image: true

concern:
  emitInterval: 5s

logLevel: info
`
	// win上用记事本打开不会正确换行
	if runtime.GOOS == "windows" {
		s = strings.ReplaceAll(s, "\n", "\r\n")
	}
	return s
}()
