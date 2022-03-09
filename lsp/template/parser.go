package template

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"strings"
)

type NodeType int

const (
	String NodeType = iota
	Key
	Pic
)

type INode interface {
	NodeType() NodeType
	ToElement(boilerplate map[string]interface{}) message.IMessageElement
}

type stringNode struct {
	s string
}

func (s *stringNode) NodeType() NodeType {
	return String
}

func (s *stringNode) ToElement(boilerplate map[string]interface{}) message.IMessageElement {
	return message.NewText(s.s)
}

type keyNode struct {
	key string
}

func (k *keyNode) NodeType() NodeType {
	return Key
}

func (k *keyNode) ToElement(boilerplate map[string]interface{}) message.IMessageElement {
	if content, found := boilerplate[k.key]; found {
		return message.NewText(fmt.Sprintf("%v", content))
	}
	return message.NewText(fmt.Sprintf("{!missing key: <%s>}", k.key))
}

type picNode struct {
	uri string
}

func (p *picNode) NodeType() NodeType {
	return Pic
}

func (p *picNode) ToElement(boilerplate map[string]interface{}) message.IMessageElement {
	if strings.HasPrefix("http://", p.uri) || strings.HasPrefix("https://", p.uri) {
		return mmsg.NewImageByUrl(p.uri)
	}
	return mmsg.NewImageByLocal(p.uri)
}

type Parser struct {
	nodes []INode
	cur   int
}

func (p *Parser) Next() bool {
	if p == nil {
		return false
	}
	return p.cur < len(p.nodes)
}

func (p *Parser) Peek() INode {
	if !p.Next() {
		return nil
	}
	node := p.nodes[p.cur]
	p.cur++
	return node
}

func (p *Parser) Parse(format string) error {
	p.nodes = nil
	p.cur = 0
	isBegin := func(s string, cur int) bool {
		if len(s[cur:]) < 2 {
			return false
		}
		return strings.HasPrefix(s[cur:], "{{")
	}
	isEnd := func(s string, cur int) bool {
		if cur == 0 || len(s) < 2 {
			return false
		}
		return strings.HasPrefix(s[cur-1:], "}}")
	}
	parseRaw := func(orig string) (INode, error) {
		s := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(orig), "{{"), "}}"))
		if len(s) == 0 {
			return nil, fmt.Errorf("empty template verb inside <%v>", orig)
		}
		if strings.HasPrefix(s, ".") {
			if strings.HasPrefix(s, ".pic ") {
				return &picNode{uri: strings.TrimPrefix(s, ".pic ")}, nil
			}
			return nil, fmt.Errorf("invalid template verb <%v> inside <%v>",
				strings.TrimPrefix(s, "."), orig)
		}
		return &keyNode{key: s}, nil
	}
	var st = 0
	var beginCount = 0
	for idx := range format {
		if beginCount == 0 && isBegin(format, idx) {
			if idx > st {
				p.nodes = append(p.nodes, &stringNode{s: format[st:idx]})
			}
			st = idx
		}
		if isBegin(format, idx) {
			beginCount++
		} else if isEnd(format, idx) {
			beginCount--
		}
		if beginCount == 0 && isEnd(format, idx) {
			if idx > st {
				node, err := parseRaw(format[st : idx+1])
				if err != nil {
					return err
				}
				p.nodes = append(p.nodes, node)
			}
			st = idx + 1
		}
		if beginCount < 0 {
			return fmt.Errorf("wrong template at pos: %v", idx)
		}
	}
	if st < len(format) {
		p.nodes = append(p.nodes, &stringNode{s: format[st:]})
	}
	return nil
}
