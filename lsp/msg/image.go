package msg

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"io"
)

type ImageBytesElement struct {
	Buf         io.ReadSeeker
	alternative string
}

func NewImage(buf io.ReadSeeker) *ImageBytesElement {
	return &ImageBytesElement{Buf: buf, alternative: "[图片]"}
}

func (i *ImageBytesElement) Alternative(s string) *ImageBytesElement {
	i.alternative = s
	return i
}

func (i *ImageBytesElement) Type() message.ElementType {
	return ImageBytes
}

func (i *ImageBytesElement) PackToElement(client *client.QQClient, target Target) message.IMessageElement {
	if i == nil {
		return message.NewText("[nil image]")
	}
	switch target.TargetType() {
	case TargetPrivate:
		if i.Buf != nil {
			img, err := client.UploadPrivateImage(target.TargetCode(), i.Buf)
			if err == nil {
				return img
			}
			logger.Errorf("TargetPrivate %v UploadGroupImage error %v", target.TargetCode(), err)
		} else {
			logger.Debugf("TargetPrivate %v nil image buf", target.TargetCode())
		}
	case TargetGroup:
		if i.Buf != nil {
			img, err := client.UploadGroupImage(target.TargetCode(), i.Buf)
			if err == nil {
				return img
			}
			logger.Errorf("TargetGroup %v UploadGroupImage error %v", target.TargetCode(), err)
		} else {
			logger.Debugf("TargetGroup %v nil image buf", target.TargetCode())
		}
	default:
		panic("ImageBytesElement PackToElement: unknown TargetType")
	}
	return message.NewText(i.alternative)
}
