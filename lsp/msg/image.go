package msg

import (
	"bytes"
	"github.com/Mrs4s/MiraiGo/message"
	"io"
)

type ImageBytesElement struct {
	Buf         io.Reader
	alternative string
}

func NewImageFromBytes(buf *bytes.Reader) *ImageBytesElement {
	return &ImageBytesElement{Buf: buf, alternative: "[图片]"}
}

func (i *ImageBytesElement) Alternative(s string) *ImageBytesElement {
	i.alternative = s
	return i
}

func (i *ImageBytesElement) Type() message.ElementType {
	return ImageBytes
}
