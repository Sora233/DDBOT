package utils

import (
	"bytes"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
)

func MessageFilter(msg []message.IMessageElement, filter func(message.IMessageElement) bool) []message.IMessageElement {
	var result []message.IMessageElement
	for _, e := range msg {
		if filter(e) {
			result = append(result, e)
		}
	}
	return result
}

func MessageTextf(format string, args ...interface{}) *message.TextElement {
	return message.NewText(fmt.Sprintf(format, args...))
}

func UploadGroupImageByUrl(groupCode int64, url string, isNorm bool, prefer proxy_pool.Prefer) (*message.GroupImageElement, error) {
	img, err := ImageGet(url, prefer)
	if err != nil {
		return nil, err
	}
	return UploadGroupImage(groupCode, img, isNorm)
}

func UploadGroupImage(groupCode int64, img []byte, isNorm bool) (image *message.GroupImageElement, err error) {
	if isNorm {
		img, err = ImageNormSize(img)
		if err != nil {
			return nil, err
		}
	}
	image, err = bot.Instance.UploadGroupImage(groupCode, bytes.NewReader(img))
	if err != nil {
		return nil, err
	}
	return image, nil

}
