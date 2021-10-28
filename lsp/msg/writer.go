package msg

import (
	"bytes"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"strings"
)

// MSG 线程不安全
type MSG struct {
	elements []message.IMessageElement

	textBuf strings.Builder
}

func NewMSG() *MSG {
	return &MSG{}
}

func NewText(s string) *MSG {
	msg := NewMSG()
	msg.Text(s)
	return msg
}

func NewTextf(format string, args ...interface{}) *MSG {
	msg := NewMSG()
	msg.Textf(format, args...)
	return msg
}

func (m *MSG) Append(e message.IMessageElement) *MSG {
	if e.Type() == message.Text {
		if textE, ok := e.(*message.TextElement); ok {
			m.textBuf.WriteString(textE.Content)
			return m
		}
	}
	m.flushText()
	m.elements = append(m.elements, e)
	return m
}

func (m *MSG) flushText() {
	if m.textBuf.Len() > 0 {
		m.elements = append(m.elements, message.NewText(m.textBuf.String()))
		m.textBuf.Reset()
	}
}

func (m *MSG) Text(s string) *MSG {
	m.textBuf.WriteString(s)
	return m
}

func (m *MSG) Textf(format string, args ...interface{}) *MSG {
	m.textBuf.WriteString(fmt.Sprintf(format, args...))
	return m
}

func (m *MSG) Image(buf *bytes.Reader) *MSG {
	m.Append(NewImage(buf))
	return m
}

func (m *MSG) ImageByUrl(url string, prefer proxy_pool.Prefer, opts ...requests.Option) *MSG {
	var img = NewImage(nil)
	b, err := utils.ImageGet(url, prefer, opts...)
	if err == nil {
		img.Buf = bytes.NewReader(b)
	}
	m.Append(img)
	return m
}

func (m *MSG) ToMessage(client *client.QQClient, target Target) *message.SendingMessage {
	var sending = message.NewSendingMessage()
	m.flushText()
	for _, e := range m.elements {
		if custom, ok := e.(CustomElement); ok {
			packed := custom.PackToElement(client, target)
			if packed != nil {
				sending.Append(packed)
			}
			continue
		}
		sending.Append(e)
	}
	return sending
}
