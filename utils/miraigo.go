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

const (
	internalMsgTypeGroup = "group"
)

type internalMsg struct {
	Type          string `json:"type"`
	MsgInfo       string `json:"msg_info"`
	ElementString string `json:"element_string"`
}

func SerializationGroupMsg(m *message.GroupMessage) (string, error) {
	elems := m.Elements
	m.Elements = nil

	defer func() {
		m.Elements = elems
	}()

	mString, err := json.MarshalToString(m)
	if err != nil {
		return "", err
	}

	elemString, err := SerializationElement(elems)
	if err != nil {
		return "", err
	}

	imsg := &internalMsg{
		Type:          internalMsgTypeGroup,
		MsgInfo:       mString,
		ElementString: elemString,
	}

	return json.MarshalToString(imsg)
}

func DeserializationGroupMsg(r string) (*message.GroupMessage, error) {
	var imsg *internalMsg
	err := json.UnmarshalFromString(r, &imsg)
	if err != nil {
		return nil, err
	}

	var m *message.GroupMessage
	err = json.UnmarshalFromString(imsg.MsgInfo, &m)
	if err != nil {
		return nil, err
	}
	elems, err := DeserializationElement(imsg.ElementString)
	if err != nil {
		return nil, err
	}
	m.Elements = elems
	return m, nil
}

type internalElem struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

const (
	internalTypeText       = "text"
	internalTypeGroupImage = "group_image"
)

// SerializationElement 序列化消息，只支持图片，文字
func SerializationElement(e []message.IMessageElement) (string, error) {
	var tmp []*internalElem

	for _, elem := range e {
		switch o := elem.(type) {
		case *message.TextElement:
			b, _ := json.Marshal(o)
			tmp = append(tmp, &internalElem{
				Type:    internalTypeText,
				Content: string(b),
			})
		case *message.GroupImageElement:
			b, _ := json.Marshal(o)
			tmp = append(tmp, &internalElem{
				Type:    internalTypeGroupImage,
				Content: string(b),
			})
		default:
			panic("unsupported element type")
		}
	}
	s, err := json.MarshalToString(tmp)
	return s, err
}

// DeserializationElement 反序列化消息，只支持图片，文字
func DeserializationElement(r string) ([]message.IMessageElement, error) {
	var tmp []*internalElem
	err := json.Unmarshal([]byte(r), &tmp)
	if err != nil {
		return nil, err
	}
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
	return result, nil
}
