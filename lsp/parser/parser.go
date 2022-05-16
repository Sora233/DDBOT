package parser

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/cfg"
	"github.com/Sora233/DDBOT/utils"
	"strings"
	"sync"
)

type Parser struct {
	Command string
	Args    []string
	// AtTarget 记录消息开头的@
	AtTarget int64

	commandName   string
	commandPrefix string
	o             sync.Once
}

func (p *Parser) Parse(elems []message.IMessageElement) {
	if len(elems) > 0 {
		var hasReply bool
		var atElem *message.AtElement
		for _, e := range elems {
			if e.Type() == message.Reply {
				hasReply = true
			}
		}
		ats := utils.MessageFilter(elems, func(element message.IMessageElement) bool {
			return element.Type() == message.At
		})
		if hasReply {
			if len(ats) >= 2 {
				atElem = ats[len(ats)-1].(*message.AtElement)
			} else if len(ats) <= 0 {
				p.AtTarget = -1 // bot reply maybe
			}
		} else {
			if len(ats) > 0 {
				atElem = ats[0].(*message.AtElement)
			}
		}
		if atElem != nil {
			p.AtTarget = atElem.Target
		}
	}
	for _, element := range elems {
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

func (p *Parser) CommandPrefix() string {
	if p == nil {
		return ""
	}
	p.match()
	return p.commandPrefix
}

// CommandName 返回command本身的名字，不包括command prefix
func (p *Parser) CommandName() string {
	if p == nil {
		return ""
	}
	p.match()
	return p.commandName
}

func (p *Parser) match() {
	p.o.Do(func() {
		var (
			err     error
			command string
			prefix  string
		)
		prefix, command, err = cfg.MatchCmdWithPrefix(p.GetCmd())
		if err != nil {
			return
		}
		p.commandPrefix = prefix
		p.commandName = command
	})
}

func NewParser() *Parser {
	return &Parser{
		Command: "",
		Args:    nil,
	}
}
