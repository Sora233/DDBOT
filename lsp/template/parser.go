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
	// TODO parse
	return nil
}
