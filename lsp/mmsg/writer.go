package mmsg

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
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
	if e == nil {
		return m
	}
	if textE, ok := e.(*message.TextElement); ok {
		m.textBuf.WriteString(textE.Content)
		return m
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

func (m *MSG) Image(buf []byte, alternative string) *MSG {
	img := NewImage(buf)
	if len(alternative) > 0 {
		img.Alternative(alternative)
	}
	m.Append(img)
	return m
}

func (m *MSG) ImageByUrl(url string, alternative string, prefer proxy_pool.Prefer, opts ...requests.Option) *MSG {
	img := NewImageByUrl(url, prefer, opts...)
	if len(alternative) > 0 {
		img.Alternative(alternative)
	}
	m.Append(img)
	return m
}

func (m *MSG) NormImageByUrl(url string, alternative string, prefer proxy_pool.Prefer, opts ...requests.Option) *MSG {
	img := NewNormImageByUrl(url, prefer, opts...)
	if len(alternative) > 0 {
		img.Alternative(alternative)
	}
	m.Append(img)
	return m
}

// ToMessage 总是返回 non-nil
func (m *MSG) ToMessage(target Target) *message.SendingMessage {
	var sending = message.NewSendingMessage()
	m.flushText()
	for _, e := range m.elements {
		if custom, ok := e.(CustomElement); ok {
			packed := custom.PackToElement(target)
			if packed != nil {
				sending.Append(packed)
			}
			continue
		}
		sending.Append(e)
	}
	return sending
}

func (m *MSG) Elements() []message.IMessageElement {
	m.flushText()
	return m.elements
}
