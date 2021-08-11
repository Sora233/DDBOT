package main

import (
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/lsp"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/alecthomas/kong"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	_ "github.com/Sora233/DDBOT/logging"
	_ "github.com/Sora233/DDBOT/lsp"
	_ "github.com/Sora233/DDBOT/miraigo-logging"
	_ "github.com/Sora233/DDBOT/msg-marker"
	_ "net/http/pprof"
)

func init() {
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

func main() {
	var cli struct {
		Play     bool  `optional:"" help:"run the play function"`
		Debug    bool  `optional:"" help:"enable debug mode"`
		SetAdmin int64 `optional:"" xor:"c" help:"set QQ number to Admin"`
		Version  bool  `optional:"" xor:"c" short:"v" help:"print the version info"`
	}
	kong.Parse(&cli)

	if cli.Version {
		fmt.Printf("Tags: %v\n", Tags)
		fmt.Printf("COMMIT_ID: %v\n", CommitId)
		fmt.Printf("BUILD_TIME: %v\n", BuildTime)
		os.Exit(0)
	}

	if b, _ := utils.FileExist("device.json"); !b {
		fmt.Println("警告：没有检测到device.json，正在生成，如果是第一次运行，可忽略")
		bot.GenRandomDevice()
	} else {
		fmt.Println("检测到device.json，使用存在的device.json")
	}

	if b, _ := utils.FileExist("application.yaml"); !b {
		fmt.Println("警告：没有检测到配置文件application.yaml，正在生成，如果是第一次运行，可忽略")
		if err := ioutil.WriteFile("application.yaml", []byte(exampleConfig), 0755); err != nil {
			fmt.Printf("application.yaml生成失败 - %v\n", err)
		} else {
			fmt.Println("最小配置application.yaml已生成，请按需修改，如需高级配置请查看帮助文档")
		}
	}

	if cli.SetAdmin != 0 {
		if err := localdb.InitBuntDB(""); err != nil {
			fmt.Println("can not init buntdb")
			os.Exit(1)
		}
		defer localdb.Close()
		sm := permission.NewStateManager()
		err := sm.GrantRole(cli.SetAdmin, permission.Admin)
		if err != nil {
			fmt.Printf("set role failed %v\n", err)
			os.Exit(1)
		}
		return
	}

	if Tags != "UNKNOWN" {
		fmt.Printf("DDBOT版本：Release版本【%v】\n", Tags)
	} else {
		if CommitId == "UNKNOWN" {
			fmt.Println("DDBOT版本：编译版本未知")
		} else {
			fmt.Printf("DDBOT版本：编译版本【%v-%v】\n", BuildTime, CommitId)
		}
	}
	fmt.Println("DDBOT唯一指定交流群：755612788")

	config.Init()

	// 快速初始化
	bot.Init()

	if cli.Debug {
		lsp.Debug = true
		go http.ListenAndServe("localhost:6060", nil)
	}

	if cli.Play {
		play()
		return
	}

	// 初始化 Modules
	bot.StartService()

	go CheckUpdate()

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
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-ch
	bot.Stop()
}

var exampleConfig = func() string {
	s := `
# 注意，填写时请把井号及后面的内容删除，并且冒号后需要加一个空格
bot:
  account:  # 你bot的qq号，不填则使用扫码登陆
  password: # 你bot的qq密码

# b站相关的功能需要一个b站账号，建议使用小号
# bot将使用您b站帐号的以下功能：
# 关注用户 / 取消关注用户 / 查看关注列表
# 请注意，订阅一个账号后，此处使用的b站账号将自动关注该账号
bilibili:
  account:  # 你的b站账号 
  password: # 你的b站密码
  interval: 15s

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
