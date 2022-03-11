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
	Reserved
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
	if boilerplate != nil {
		if content, found := boilerplate[k.key]; found {
			return message.NewText(fmt.Sprintf("%v", content))
		}
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
