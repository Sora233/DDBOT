package huya

import (
	"bytes"
	"errors"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/Sora233/DDBOT/v2/proxy_pool"
	"github.com/Sora233/DDBOT/v2/requests"
	"github.com/Sora233/DDBOT/v2/utils"
)

func RoomPage(roomId string) (*LiveInfo, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	log := logger.WithField("RoomId", roomId)
	url := HuyaPath(roomId)
	var opts = []requests.Option{
		requests.AddUAOption(),
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.RetryOption(3),
		requests.TimeoutOption(time.Second * 10),
	}
	var body = new(bytes.Buffer)
	err := requests.Get(url, nil, body, opts...)
	if err != nil {
		return nil, err
	}
	if strings.Contains(body.String(), "找不到这个主播") {
		return nil, ErrRoomNotExist
	}
	if strings.Contains(body.String(), "涉嫌违规") {
		return nil, ErrRoomBanned
	}
	ri := new(LiveInfo)
	ri.RoomId = roomId
	ri.RoomUrl = url
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}
	livingSpan := doc.Find("span.host-prevStartTime,span.host-spectator")
	if livingSpan.Size() != 1 {
		log.Errorf("living span size %v", livingSpan.Size())
		return nil, errors.New("can not determine live status")
	}
	for _, attr := range livingSpan.Get(0).Attr {
		if attr.Key != "class" {
			continue
		}
		if attr.Val == "host-prevStartTime" {
			ri.IsLiving = false
		} else {
			ri.IsLiving = true
		}
	}

	nameH3 := doc.Find(".host-name")
	if nameH3.Size() == 1 {
		name, found := nameH3.Attr("title")
		if found {
			ri.Name = name
		} else {
			log.Errorf("h3.host-name[title] not found")
		}
	} else {
		log.Errorf("h3.host-name not found")
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
			log.Errorf("#avatar-img[src] not found")
		}
	} else {
		log.Errorf("#avatar-img not found")
	}

	roomNameH1 := doc.Find(".host-title")
	if roomNameH1.Size() == 1 {
		roomName, found := roomNameH1.Attr("title")
		if found {
			ri.RoomName = roomName
		} else {
			log.Errorf("h1.host-title[title] not found")
		}
	} else {
		log.Errorf("h1.host-title no found")
	}

	return ri, nil
}
