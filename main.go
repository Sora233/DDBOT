package main

import (
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"

	_ "github.com/Logiase/MiraiGo-Template/modules/logging"
	_ "github.com/Sora233/Sora233-MiraiGo/lsp"
)

func init() {
	utils.WriteLogToFS()
	config.Init()
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)

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

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
	bot.Stop()
}
