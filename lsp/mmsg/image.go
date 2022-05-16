package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"io/ioutil"
)

type ImageBytesElement struct {
	Buf         []byte
	alternative string
}

func NewImage(buf []byte) *ImageBytesElement {
	return &ImageBytesElement{Buf: buf, alternative: "[图片]"}
}

// NewImageByUrl 默认会对相同的url使用缓存
func NewImageByUrl(url string, opts ...requests.Option) *ImageBytesElement {
	var img = NewImage(nil)
	b, err := utils.ImageGet(url, opts...)
	if err == nil {
		img.Buf = b
	} else {
		logger.WithField("url", url).Errorf("ImageGet error %v", err)
	}
	return img
}

// NewImageByUrlWithoutCache 默认情况下相同的url会存在缓存，
// 如果url会随机返回不同的图片，则需要禁用缓存
// 这个函数就是不使用缓存的版本
func NewImageByUrlWithoutCache(url string, opts ...requests.Option) *ImageBytesElement {
	var img = NewImage(nil)
	b, err := utils.ImageGetWithoutCache(url, opts...)
	if err == nil {
		img.Buf = b
	} else {
		logger.WithField("url", url).Errorf("ImageGet error %v", err)
	}
	return img
}

func NewImageByLocal(filepath string) *ImageBytesElement {
	var img = NewImage(nil)
	b, err := ioutil.ReadFile(filepath)
	if err == nil {
		img.Buf = b
	} else {
		logger.WithField("filepath", filepath).Errorf("ReadFile error %v", err)
	}
	return img
}

func (i *ImageBytesElement) Norm() *ImageBytesElement {
	if i == nil || i.Buf == nil {
		return i
	}
	b, err := utils.ImageNormSize(i.Buf)
	if err == nil {
		i.Buf = b
	} else {
		logger.Errorf("mmsg: ImageBytesElement Norm error %v", err)
	}
	return i
}

func (i *ImageBytesElement) Resize(width, height uint) *ImageBytesElement {
	if i == nil || i.Buf == nil {
		return i
	}
	b, err := utils.ImageResize(i.Buf, width, height)
	if err == nil {
		i.Buf = b
	} else {
		logger.Errorf("mmsg: ImageBytesElement Resize error %v", err)
	}
	return i
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
	if i.alternative == "" {
		return message.NewText("")
	}
	return message.NewText(i.alternative + "\n")
}
