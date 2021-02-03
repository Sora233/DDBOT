package main

import (
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/lsp"
	"github.com/alecthomas/kong"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"

	_ "github.com/Sora233/Sora233-MiraiGo/logging"
	_ "github.com/Sora233/Sora233-MiraiGo/lsp"
)

func init() {
	utils.WriteLogToFS()
	config.Init()
}

func main() {
	var cli struct {
		Play  bool `optional:"" help:"run the play function"`
		Debug bool `optional:"" help:"enable debug mode"`
	}
	kong.Parse(&cli)

	if cli.Debug {
		lsp.Debug = true
		go http.ListenAndServe("localhost:6060", nil)
	}

	if cli.Play {
		play()
		return
	}

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

	if lsp.Instance != nil {
		lsp.Instance.FreshIndex()
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
	bot.Stop()
}
