package utils

import (
	"bytes"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/proxy_pool"
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

type internalMsg struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

const (
	internalTypeText       = "text"
	internalTypeGroupImage = "group_image"
)

// ToStringMsg 序列化消息，只支持图片，文字
func ToStringMsg(e []message.IMessageElement) string {
	var tmp []*internalMsg

	for _, elem := range e {
		switch o := elem.(type) {
		case *message.TextElement:
			b, _ := json.Marshal(o)
			tmp = append(tmp, &internalMsg{
				Type:    internalTypeText,
				Content: string(b),
			})
		case *message.GroupImageElement:
			b, _ := json.Marshal(o)
			tmp = append(tmp, &internalMsg{
				Type:    internalTypeGroupImage,
				Content: string(b),
			})
		default:
			panic("unsupported element type")
		}
	}
	s, _ := json.MarshalToString(tmp)
	return s
}

// FromStringMsg 反序列化消息，只支持图片，文字
func FromStringMsg(r string) []message.IMessageElement {
	var tmp []*internalMsg
	json.Unmarshal([]byte(r), &tmp)
	var result []message.IMessageElement
	for _, e := range tmp {
		switch e.Type {
		case internalTypeGroupImage:
			var elem *message.GroupImageElement
			json.UnmarshalFromString(e.Content, &elem)
			if elem != nil {
				result = append(result, elem)
			}
		case internalTypeText:
			var elem *message.TextElement
			json.UnmarshalFromString(e.Content, &elem)
			if elem != nil {
				result = append(result, elem)
			}
		default:
			panic("unsupported element type")
		}
	}
	return result
}
