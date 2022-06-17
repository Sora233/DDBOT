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
	// AtArgs 记录命令后的@
	AtArgs []int64

	commandName   string
	commandPrefix string
	o             sync.Once
}

func (p *Parser) Parse(elems []message.IMessageElement) {
	if len(elems) > 0 {
		var search []message.IMessageElement
		if elems[0].Type() == message.Reply {
			if elems[1].Type() == message.At {
				search = elems[2:]
			} else {
				search = elems[1:]
			}
		} else {
			search = elems[:]
		}
		if len(search) > 0 && search[0].Type() == message.At {
			p.AtTarget = search[0].(*message.AtElement).Target
			search = search[1:]
		}
		var afterCmd = false
		for _, e := range search {
			if afterCmd && e.Type() == message.At {
				p.AtArgs = append(p.AtArgs, e.(*message.AtElement).Target)
			}
			if !afterCmd && e.Type() != message.At {
				afterCmd = true
			}
		}
	}
	var buf strings.Builder
	textElems := utils.MessageFilter(elems, func(element message.IMessageElement) bool {
		return element.Type() == message.Text
	})
	for _, element := range textElems {
		if te, ok := element.(*message.TextElement); ok {
			text := strings.TrimSpace(strings.Replace(te.Content, " ", " ", -1))
			if text == "" {
				continue
			}
			buf.WriteString(text)
			buf.WriteString(" ")
		}
	}
	splitStr := utils.ArgSplit(strings.TrimSpace(buf.String()))
	if len(splitStr) >= 1 {
		p.Command = strings.TrimSpace(splitStr[0])
		for _, s := range splitStr[1:] {
			p.Args = append(p.Args, strings.TrimSpace(s))
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

func (p *Parser) GetAtArgs() []int64 {
	return p.AtArgs
}

func (p *Parser) GetCmdArgs() []string {
	result := []string{p.Command}
	result = append(result, p.Args...)
	return result
}

func (p *Parser) AtCheck() bool {
	if p.AtTarget <= 0 {
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
