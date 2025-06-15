package utils

import (
	"bytes"
	"errors"

	"github.com/LagrangeDev/LagrangeGo/message"

	"github.com/Sora233/MiraiGo-Template/bot"
)

func UploadGroupImageByUrl(groupCode uint32, url string, isNorm bool) (*message.ImageElement, error) {
	img, err := ImageGet(url)
	if err != nil {
		return nil, err
	}
	return UploadGroupImage(groupCode, img, isNorm)
}

func UploadGroupImage(groupCode uint32, img []byte, isNorm bool) (image *message.ImageElement, err error) {
	if isNorm {
		img, err = ImageNormSize(img)
		if err != nil {
			return nil, err
		}
	}
	if !GetBot().IsOnline() {
		return nil, errors.New("bot offline")
	}
	return bot.QQClient.UploadImage(message.Source{SourceType: message.SourceGroup, PrimaryID: int64(groupCode)}, &message.ImageElement{Size: uint32(len(img)), Stream: bytes.NewReader(img)})
}

func UploadPrivateImage(uin uint32, img []byte, isNorm bool) (*message.ImageElement, error) {
	var err error
	if isNorm {
		img, err = ImageNormSize(img)
		if err != nil {
			return nil, err
		}
	}
	if !GetBot().IsOnline() {
		return nil, errors.New("bot offline")
	}
	return bot.QQClient.UploadImage(message.Source{SourceType: message.SourcePrivate, PrimaryID: int64(uin)}, &message.ImageElement{Size: uint32(len(img)), Stream: bytes.NewReader(img)})
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
	m.ToString()

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
	internalTypeText        = "text"
	internalTypeGroupImage  = "group_image"
	internalTypeFriendImage = "friend_image"
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
		case *message.ImageElement:
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
		case internalTypeGroupImage, internalTypeFriendImage:
			var elem *message.ImageElement
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
