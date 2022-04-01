package mmsg

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/requests"
	"strings"
)

// MSG 线程不安全
type MSG struct {
	elements []message.IMessageElement
	textBuf  strings.Builder
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

func (m *MSG) Append(elems ...message.IMessageElement) *MSG {
	if len(elems) == 0 {
		return m
	}
	for _, e := range elems {
		if e == nil {
			continue
		}
		if textE, ok := e.(*message.TextElement); ok {
			m.textBuf.WriteString(textE.Content)
			continue
		}
		m.flushText()
		m.elements = append(m.elements, e)
	}
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

func (m *MSG) ImageWithNorm(buf []byte, alternative string) *MSG {
	img := NewImageWithNorm(buf)
	if len(alternative) > 0 {
		img.Alternative(alternative)
	}
	m.Append(img)
	return m
}

func (m *MSG) ImageByUrl(url string, alternative string, opts ...requests.Option) *MSG {
	img := NewImageByUrl(url, opts...)
	if len(alternative) > 0 {
		img.Alternative(alternative)
	}
	m.Append(img)
	return m
}

func (m *MSG) ImageByUrlWithNorm(url string, alternative string, opts ...requests.Option) *MSG {
	img := NewImageByUrlWithNorm(url, opts...)
	if len(alternative) > 0 {
		img.Alternative(alternative)
	}
	m.Append(img)
	return m
}

// ToCombineMessage 总是返回 non-nil
func (m *MSG) ToCombineMessage(target Target) *message.SendingMessage {
	var result = message.NewSendingMessage()
	sms := m.ToMessage(target)
	for _, sm := range sms {
		for _, e := range sm.Elements {
			result.Append(e)
		}
	}
	return result
}

// ToMessage 返回消息用于发送
func (m *MSG) ToMessage(target Target) []*message.SendingMessage {
	if m == nil {
		return nil
	}
	var result []*message.SendingMessage
	m.Cut()
	var sending = message.NewSendingMessage()
	for _, e := range m.elements {
		if custom, ok := e.(CustomElement); ok {
			if e.Type() == Cut {
				if len(sending.Elements) > 0 {
					result = append(result, sending)
					sending = message.NewSendingMessage()
				}
			} else {
				packed := custom.PackToElement(target)
				if packed != nil {
					sending.Append(packed)
				}
			}
			continue
		}
		sending.Append(e)
	}
	if len(sending.Elements) > 0 {
		result = append(result, sending)
	}
	return result
}

func (m *MSG) Cut() {
	m.flushText()
	if len(m.elements) > 0 {
		m.elements = append(m.elements, new(CutElement))
	}
}

func (m *MSG) Elements() []message.IMessageElement {
	m.flushText()
	return m.elements
}
