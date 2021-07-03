package command

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/utils"
	"strings"
)

type Parser struct {
	Command string
	Args    []string
}

func (p *Parser) Parse(e []message.IMessageElement) {
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

func NewParser() *Parser {
	return &Parser{
		Command: "",
		Args:    nil,
	}
}
