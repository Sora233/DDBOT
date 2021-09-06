package msg

import (
	"bytes"
	"fmt"
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
		m.textBuf.WriteString(e.(*message.TextElement).Content)
	} else {
		m.flushText()
		m.elements = append(m.elements, e)
	}
}

func (m *MSG) flushText() {
	if m.textBuf.Len() > 0 {
		m.elements = append(m.elements, message.NewText(m.textBuf.String()))
		m.textBuf.Reset()
	}
}

func (m *MSG) Text(s string) {
	m.textBuf.WriteString(s)
}

func (m *MSG) Textf(format string, args ...interface{}) {
	m.textBuf.WriteString(fmt.Sprintf(format, args...))
}

func (m *MSG) ImageBytes(buf *bytes.Reader) {
	m.append(NewImageFromBytes(buf))
}

func (m *MSG) ImageByUrl(url string, opts ...requests.Option) {
	var img = NewImageFromBytes(nil)
	var body = new(bytes.Buffer)
	err := requests.Get(url, nil, body, opts...)
	if err == nil {
		img.Buf = body
	}
	m.append(img)
}

func (m *MSG) Raw(e message.IMessageElement) {
	m.append(e)
}

func (m *MSG) ToPrivateMSG(target int64) message.SendingMessage {
	m.flushText()
}
