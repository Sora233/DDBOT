package main

import (
	"fmt"
	"github.com/Sora233/DDBOT"
	"github.com/Sora233/DDBOT/lsp"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/alecthomas/kong"
	"net/http"
	"os"

	_ "github.com/Sora233/DDBOT/logging"
	_ "github.com/Sora233/DDBOT/lsp/acfun"
	_ "github.com/Sora233/DDBOT/lsp/douyu"
	_ "github.com/Sora233/DDBOT/lsp/huya"
	_ "github.com/Sora233/DDBOT/lsp/weibo"
	_ "github.com/Sora233/DDBOT/lsp/youtube"
	_ "github.com/Sora233/DDBOT/miraigo-logging"
	_ "github.com/Sora233/DDBOT/msg-marker"
	_ "net/http/pprof"
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
		fmt.Printf("Tags: %v\n", Tags)
		fmt.Printf("COMMIT_ID: %v\n", CommitId)
		fmt.Printf("BUILD_TIME: %v\n", BuildTime)
		os.Exit(0)
	}

	if cli.SetAdmin != 0 {
		if err := localdb.InitBuntDB(""); err != nil {
			fmt.Printf("初始化Buntdb失败 %v\n", err)
			os.Exit(1)
		}
		defer localdb.Close()
		sm := permission.NewStateManager()
		err := sm.GrantRole(cli.SetAdmin, permission.Admin)
		if err != nil {
			fmt.Printf("设置Admin权限失败 %v\n", err)
		}
		return
	}

	if cli.SyncBilibili {
		config.Init()
		if err := localdb.InitBuntDB(""); err != nil {
			fmt.Printf("初始化buntdb失败 %v \n", err)
			return
		}
		defer localdb.Close()
		c := bilibili.NewConcern(nil)
		c.StateManager.FreshIndex()
		bilibili.Init()
		c.SyncSub()
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

	if cli.Debug {
		lsp.Debug = true
		go http.ListenAndServe("localhost:6060", nil)
	}

	if cli.Play {
		play()
		return
	}

	DDBOT.Run()
}
