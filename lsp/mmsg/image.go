package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
)

type ImageBytesElement struct {
	Buf         []byte
	alternative string
}

func NewImage(buf []byte) *ImageBytesElement {
	return &ImageBytesElement{Buf: buf, alternative: "[图片]"}
}

func NewImageByUrl(url string, prefer proxy_pool.Prefer, opts ...requests.Option) *ImageBytesElement {
	var img = NewImage(nil)
	var b []byte
	var err error
	b, err = utils.ImageGet(url, prefer, opts...)
	if err == nil {
		img.Buf = b
	} else {
		logger.WithField("url", url).Errorf("ImageGet error %v", err)
	}
	return img
}

func NewNormImageByUrl(url string, prefer proxy_pool.Prefer, opts ...requests.Option) *ImageBytesElement {
	var img = NewImage(nil)
	var b []byte
	var err error
	b, err = utils.ImageGetAndNorm(url, prefer, opts...)
	if err == nil {
		img.Buf = b
	} else {
		logger.WithField("url", url).Errorf("ImageGet error %v", err)
	}
	return img
}

func (i *ImageBytesElement) Alternative(s string) *ImageBytesElement {
	i.alternative = s
	return i
}

func (i *ImageBytesElement) Type() message.ElementType {
	return ImageBytes
}

func (i *ImageBytesElement) PackToElement(target Target) message.IMessageElement {
	if i == nil {
		return message.NewText("[nil image]\n")
	}
	switch target.TargetType() {
	case TargetPrivate:
		if i.Buf != nil {
			img, err := utils.UploadPrivateImage(target.TargetCode(), i.Buf, false)
			if err == nil {
				return img
			}
			logger.Errorf("TargetPrivate %v UploadGroupImage error %v", target.TargetCode(), err)
		} else {
			logger.Debugf("TargetPrivate %v nil image buf", target.TargetCode())
		}
	case TargetGroup:
		if i.Buf != nil {
			img, err := utils.UploadGroupImage(target.TargetCode(), i.Buf, false)
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
	return message.NewText(i.alternative + "\n")
}
