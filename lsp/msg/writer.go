package msg

import (
	"bytes"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
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

func (m *MSG) append(e message.IMessageElement) {
	if e.Type() == message.Text {
		if textE, ok := e.(*message.TextElement); ok {
			m.textBuf.WriteString(textE.Content)
			return
		}
	}
	m.flushText()
	m.elements = append(m.elements, e)
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
	m.append(NewImage(buf))
	return m
}

func (m *MSG) ImageByUrl(url string, opts ...requests.Option) *MSG {
	var img = NewImage(nil)
	var body = new(bytes.Buffer)
	err := requests.Get(url, nil, body, opts...)
	if err == nil {
		img.Buf = bytes.NewReader(body.Bytes())
	}
	m.append(img)
	return m
}

func (m *MSG) Raw(e message.IMessageElement) *MSG {
	m.append(e)
	return m
}

func (m *MSG) ToMessage(client *client.QQClient, target Target) (*message.SendingMessage, []func()) {
	var sending = message.NewSendingMessage()
	var f []func()
	m.flushText()
	for _, e := range m.elements {
		if custom, ok := e.(CustomElement); ok {
			packed := custom.PackToElement(client, target)
			if packed.Type() == Fn {
				f = append(f, packed.(*FnElement).f)
				continue
			}
			if packed != nil {
				sending.Append(packed)
			}
			continue
		}
		sending.Append(e)
	}
	return sending, f
}

//// Send 根据TargetType返回message.GroupMessage或者message.PrivateMessage
//func (m *MSG) Send(client *client.QQClient, target Target) interface{} {
//	msg, callback := m.ToMessage(client, target)
//	var result interface{}
//	switch target.TargetType() {
//	case TargetGroup:
//		result = client.SendGroupMessage(target.TargetCode(), msg)
//	case TargetPrivate:
//		result = client.SendPrivateMessage(target.TargetCode(), msg)
//	default:
//		panic("MSG Send: unknown target type")
//	}
//	for _,f := range callback {
//		f()
//	}
//	return result
//}
