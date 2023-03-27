package main

import (
	"fmt"
	"github.com/Sora233/DDBOT"
	_ "github.com/Sora233/DDBOT/logging"
	"github.com/Sora233/DDBOT/lsp"
	_ "github.com/Sora233/DDBOT/lsp/acfun"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	_ "github.com/Sora233/DDBOT/lsp/douyu"
	_ "github.com/Sora233/DDBOT/lsp/huya"
	"github.com/Sora233/DDBOT/lsp/permission"
	_ "github.com/Sora233/DDBOT/lsp/weibo"
	_ "github.com/Sora233/DDBOT/lsp/youtube"
	_ "github.com/Sora233/DDBOT/msg-marker"
	"github.com/Sora233/DDBOT/warn"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/alecthomas/kong"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
)

func main() {
	var cli struct {
		Play         bool  `optional:"" help:"运行play函数，适用于测试和开发"`
		Debug        bool  `optional:"" help:"启动debug模式"`
		SetAdmin     int64 `optional:"" xor:"c" help:"设置admin权限"`
		Version      bool  `optional:"" xor:"c" short:"v" help:"打印版本信息"`
		SyncBilibili bool  `optional:"" xor:"c" help:"同步b站帐号的关注，适用于更换或迁移b站帐号的时候"`
	}
	kong.Parse(&cli)

	if cli.Version {
		fmt.Printf("Tags: %v\n", lsp.Tags)
		fmt.Printf("COMMIT_ID: %v\n", lsp.CommitId)
		fmt.Printf("BUILD_TIME: %v\n", lsp.BuildTime)
		os.Exit(0)
	}

	if err := localdb.InitBuntDB(""); err != nil {
		if err == localdb.ErrLockNotHold {
			warn.Warn("tryLock数据库失败：您可能重复启动了这个BOT！\n如果您确认没有重复启动，请删除.lsp.db.lock文件并重新运行。")
		} else {
			warn.Warn("无法正常初始化数据库！请检查.lsp.db文件权限是否正确，如无问题则为数据库文件损坏，请阅读文档获得帮助。")
		}
		return
	}

	if runtime.GOOS == "windows" {
		if err := exitHook(func() {
			localdb.Close()
		}); err != nil {
			localdb.Close()
			warn.Warn("无法正常初始化Windows环境！")
			return
		}
	} else {
		defer localdb.Close()
	}

	if cli.SetAdmin != 0 {
		sm := permission.NewStateManager()
		err := sm.GrantRole(cli.SetAdmin, permission.Admin)
		if err != nil {
			fmt.Printf("设置Admin权限失败 %v\n", err)
		}
		return
	}

	if cli.SyncBilibili {
		config.Init()
		c := bilibili.NewConcern(nil)
		c.StateManager.FreshIndex()
		bilibili.Init()
		c.SyncSub()
		return
	}

	fmt.Println("DDBOT唯一指定交流群：755612788")

	if cli.Debug {
		lsp.Debug = true
		go http.ListenAndServe("localhost:6060", nil)
	}

	if cli.Play {
		play()
		return
	}

	DDBOT.SetUpLog()

	DDBOT.Run()
}
