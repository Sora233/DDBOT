package lsp

import (
	"fmt"
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Sora233/Sora233-MiraiGo/lsp/command"
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

func (r *Runtime) parseCommandSyntax(ast interface{}, name string, options ...kong.Option) string {
	args := r.GetArgs()
	cmdOut := &strings.Builder{}
	options = append(options, kong.Name(name), kong.UsageOnError(), kong.Exit(r.Exit))
	k, err := kong.New(ast, options...)
	if err != nil {
		logger.Errorf("kong new failed %v", err)
		r.Exit(0)
		return ""
	}
	k.Stdout = cmdOut
	_, err = k.Parse(args)
	if r.exit {
		logger.WithField("content", args).Debug("exit")
		return cmdOut.String()
	}
	if err != nil {
		logger.WithField("content", args).Errorf("kong parse failed %v", err)
		r.Exit(0)
		return fmt.Sprintf("失败 - %v", err)
	}
	return ""
}

func NewRuntime(bot *miraiBot.Bot, l *Lsp) *Runtime {
	r := &Runtime{
		bot:    bot,
		l:      l,
		Parser: command.NewParser(),
	}
	return r
}
