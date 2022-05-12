package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
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

func (i *ImageBytesElement) PackToElement(target mt.Target) message.IMessageElement {
	if i == nil {
		return message.NewText("[nil image]\n")
	}
	switch target.GetTargetType() {
	case mt.TargetPrivate:
		targetCode := target.(*mt.PrivateTarget).TargetCode()
		if i.Buf != nil {
			img, err := utils.UploadPrivateImage(targetCode, i.Buf, false)
			if err == nil {
				return img
			}
			logger.Errorf("TargetPrivate %v UploadGroupImage error %v", targetCode, err)
		} else {
			logger.Debugf("TargetPrivate %v nil image buf", targetCode)
		}
	case mt.TargetGroup:
		targetCode := target.(*mt.GroupTarget).TargetCode()
		if i.Buf != nil {
			img, err := utils.UploadGroupImage(targetCode, i.Buf, false)
			if err == nil {
				return img
			}
			logger.Errorf("TargetGroup %v UploadGroupImage error %v", targetCode, err)
		} else {
			logger.Debugf("TargetGroup %v nil image buf", targetCode)
		}
	case mt.TargetGuild:
		guildId := target.(*mt.GuildTarget).GuildId
		channelId := target.(*mt.GuildTarget).ChannelId
		if i.Buf != nil {
			img, err := utils.UploadGuildImage(guildId, channelId, i.Buf, false)
			if err == nil {
				return img
			}
			logger.Errorf("TargetGuild %v - %v UploadGroupImage error %v", guildId, channelId, err)
		} else {
			logger.Debugf("TargetGroup %v - %v nil image buf", guildId, channelId)
		}
	default:
		panic("ImageBytesElement PackToElement: unknown GetTargetType")
	}
	if i.alternative == "" {
		return message.NewText("")
	}
	return message.NewText(i.alternative + "\n")
}
