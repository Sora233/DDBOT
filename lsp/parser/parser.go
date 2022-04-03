package parser

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/cfg"
	"github.com/Sora233/DDBOT/utils"
	"go.uber.org/atomic"
	"strings"
)

type Parser struct {
	Command string
	Args    []string
	// AtTarget 记录消息开头的@
	AtTarget int64

	commandName atomic.String
}

func (p *Parser) Parse(e []message.IMessageElement) {
	if len(e) > 0 {
		var atElem *message.AtElement
		if e[0].Type() == message.At {
			atElem, _ = e[0].(*message.AtElement)
		} else if e[0].Type() == message.Reply && len(e) > 1 && e[1].Type() == message.At {
			atElem, _ = e[1].(*message.AtElement)
		}
		if atElem != nil {
			p.AtTarget = atElem.Target
		}
	}
	for _, element := range e {
		if te, ok := element.(*message.TextElement); ok {
			text := strings.TrimSpace(strings.Replace(te.Content, " ", " ", -1))
			if text == "" {
				continue
			}
			splitStr := utils.ArgSplit(text)
			if len(splitStr) >= 1 {
				p.Command = strings.TrimSpace(splitStr[0])
				for _, s := range splitStr[1:] {
					p.Args = append(p.Args, strings.TrimSpace(s))
				}
			}
			break
		}
	}
}

// GetCmd 返回包括commandPrefix在内的command字符串
func (p *Parser) GetCmd() string {
	return p.Command
}

func (p *Parser) GetArgs() []string {
	return p.Args
}

func (p *Parser) GetCmdArgs() []string {
	result := []string{p.Command}
	result = append(result, p.Args...)
	return result
}

func (p *Parser) AtCheck() bool {
	if p.AtTarget == 0 {
		return true
	}
	return p.AtTarget == utils.GetBot().GetUin()
}

// CommandName 返回command本身的名字，不包括command prefix
func (p *Parser) CommandName() string {
	if p == nil {
		return ""
	}
	x := p.commandName.Load()
	if x == "" {
		x = strings.TrimPrefix(p.GetCmd(), cfg.GetCommandPrefix())
		p.commandName.Store(x)
	}
	return x
}

func NewParser() *Parser {
	return &Parser{
		Command: "",
		Args:    nil,
	}
}
