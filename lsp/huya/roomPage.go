package huya

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/proxy_pool/requests"
	"github.com/Sora233/DDBOT/utils"
	"strings"
	"time"
)

type LiveInfo struct {
	RoomId   string `json:"room_id"`
	RoomUrl  string `json:"room_url"`
	Avatar   string `json:"avatar"`
	Name     string `json:"name"`
	RoomName string `json:"room_name"`
	Living   bool   `json:"living"`
}

func (m *LiveInfo) Type() EventType {
	return Live
}

func (m *LiveInfo) ToString() string {
	if m == nil {
		return ""
	}
	bin, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(bin)
}

func RoomPage(roomId string) (*LiveInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := HuyaPath(roomId)
	resp, err := requests.Get(ctx, url, nil, 3,
		requests.AddUAOption(),
		requests.ProxyOption(proxy_pool.PreferAny),
	)
	if err != nil {
		return nil, err
	}
	b, err := resp.Content()
	if err != nil {
		return nil, err
	}
	if strings.Contains(string(b), "找不到这个主播") {
		return nil, errors.New("房间不存在")
	}
	ri := new(LiveInfo)
	ri.RoomId = roomId
	ri.RoomUrl = url
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	livingSpan := doc.Find("span.host-prevStartTime,span.host-spectator")
	if livingSpan.Size() != 1 {
		logger.WithField("room_id", roomId).Errorf("living span size %v", livingSpan.Size())
		return nil, errors.New("can not determine live status")
	}
	for _, attr := range livingSpan.Get(0).Attr {
		if attr.Key != "class" {
			continue
		}
		if attr.Val == "host-prevStartTime" {
			ri.Living = false
		} else {
			ri.Living = true
		}
	}

	nameH3 := doc.Find(".host-name")
	if nameH3.Size() == 1 {
		name, found := nameH3.Attr("title")
		if found {
			ri.Name = name
		} else {
			logger.WithField("room_id", roomId).Errorf("h3.host-name[title] not found")
		}
	} else {
		logger.WithField("room_id", roomId).Errorf("h3.host-name not found")
	}

	avaImg := doc.Find("#avatar-img")
	if avaImg.Size() == 1 {
		ava, found := avaImg.Attr("src")
		if found {
			if strings.HasPrefix(ava, "//") {
				ava = "https:" + ava
			}
			ri.Avatar = ava
		} else {
			logger.WithField("room_id", roomId).Errorf("#avatar-img[src] not found")
		}
	} else {
		logger.WithField("room_id", roomId).Errorf("#avatar-img not found")
	}

	roomNameH1 := doc.Find(".host-title")
	if roomNameH1.Size() == 1 {
		roomName, found := roomNameH1.Attr("title")
		if found {
			ri.RoomName = roomName
		} else {
			logger.WithField("room_id", roomId).Errorf("h1.host-title[title] not found")
		}
	} else {
		logger.WithField("room_id", roomId).Errorf("h1.host-title no found")
	}

	return ri, nil
}
