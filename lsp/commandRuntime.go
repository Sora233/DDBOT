package lsp

import (
	"errors"
	"fmt"
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	"github.com/Sora233/DDBOT/lsp/command"
	"github.com/Sora233/DDBOT/lsp/douyu"
	"github.com/Sora233/DDBOT/lsp/huya"
	"github.com/Sora233/DDBOT/lsp/youtube"
	"github.com/Sora233/DDBOT/utils"
	"github.com/alecthomas/kong"
	"strings"
)

type Runtime struct {
	bot *miraiBot.Bot
	l   *Lsp
	*command.Parser

	debug bool
	exit  bool
}

func (r *Runtime) Exit(int) {
	r.exit = true
}
func (r *Runtime) Debug() {
	r.debug = true
}

func (r *Runtime) parseCommandSyntax(ast interface{}, name string, options ...kong.Option) (*kong.Context, string) {
	args := r.GetArgs()
	cmdOut := &strings.Builder{}
	// kong 错误信息不太友好
	options = append(options, kong.Name(name), kong.UsageOnError(), kong.Exit(r.Exit))
	k, err := kong.New(ast, options...)
	if err != nil {
		logger.Errorf("kong new failed %v", err)
		r.Exit(0)
		return nil, ""
	}
	k.Stdout = cmdOut
	ctx, err := k.Parse(args)
	if r.exit {
		logger.WithField("content", args).Debug("exit")
		return ctx, cmdOut.String()
	}
	if err != nil {
		logger.WithField("content", args).Errorf("kong parse failed %v", err)
		r.Exit(0)
		return nil, fmt.Sprintf("参数解析失败 - %v", err)
	}
	return ctx, ""
}

func (r *Runtime) ParseRawSiteAndType(rawSite string, rawType string) (string, concern.Type, error) {
	var (
		site      string
		_type     string
		found     bool
		watchType concern.Type
		err       error
	)
	rawSite = strings.Trim(rawSite, `"`)
	rawType = strings.Trim(rawType, `"`)
	site, err = r.ParseRawSite(rawSite)
	if err != nil {
		return "", concern.Empty, err
	}
	_type, found = utils.PrefixMatch([]string{"live", "news"}, rawType)
	if !found {
		return "", concern.Empty, errors.New("不支持的类型参数")
	}

	switch _type {
	case "live":
		if site == bilibili.Site {
			watchType = concern.BibiliLive
		} else if site == douyu.Site {
			watchType = concern.DouyuLive
		} else if site == youtube.Site {
			watchType = concern.YoutubeLive
		} else if site == huya.Site {
			watchType = concern.HuyaLive
		} else {
			return "", concern.Empty, errors.New("不支持的类型参数")
		}
	case "news":
		if site == bilibili.Site {
			watchType = concern.BilibiliNews
		} else if site == youtube.Site {
			watchType = concern.YoutubeVideo
		} else {
			return "", concern.Empty, errors.New("不支持的类型参数")
		}
	default:
		return "", concern.Empty, errors.New("不支持的类型参数")
	}
	return site, watchType, nil
}

func (r *Runtime) ParseRawSite(rawSite string) (string, error) {
	var (
		found bool
		site  string
	)

	site, found = utils.PrefixMatch([]string{bilibili.Site, douyu.Site, youtube.Site, huya.Site}, rawSite)
	if !found {
		return "", errors.New("不支持的网站参数")
	}
	return site, nil
}

func NewRuntime(bot *miraiBot.Bot, l *Lsp) *Runtime {
	r := &Runtime{
		bot:    bot,
		l:      l,
		Parser: command.NewParser(),
	}
	return r
}
