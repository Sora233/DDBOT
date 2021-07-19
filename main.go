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
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/Sora233/DDBOT/logging"
	_ "github.com/Sora233/DDBOT/lsp"
)

func init() {
	utils.WriteLogToFS()
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

var exampleConfig = `
bot:
  account:  # 你的qq号，不填则使用扫码登陆
  password: # 你的qq密码

# b站登陆后的cookie字段，从cookie中找到这两个填进去，如果不会请百度搜索如何查看网站cookies
# 请注意，bot将使用您b站帐号的以下功能，建议使用新注册的小号：
# 关注用户 / 取消关注用户 / 查看关注列表
# 警告：
# SESSDATA和bili_jct等价于您的帐号凭证
# 请绝对不要透露给他人，更不能上传至Github等公开平台
# 否则将导致您的帐号被盗
bilibili:
  SESSDATA:
  bili_jct:
  interval: 15s

concern:
  emitInterval: 5s

logLevel: info
`
