package lolicon_pool

import (
	"errors"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	requests "github.com/asmcos/requests"
	"time"
)

const Host = "https://api.lolicon.app/setu"

type R18Type int

const (
	R18_OFF R18Type = iota
	R18_ON
	R18_MIX
)

type Request struct {
	Apikey   string `json:"apikey"`
	R18      int    `json:"r18"`
	Keyword  string `json:"keyword"`
	Num      int    `json:"num"`
	Proxy    string `json:"proxy"`
	Size1200 bool   `json:"size1200"`
}

type Setu struct {
	Pid    int      `json:"pid"`
	P      int      `json:"p"`
	Uid    int      `json:"uid"`
	Title  string   `json:"title"`
	Author string   `json:"author"`
	Url    string   `json:"url"`
	R18    bool     `json:"r18"`
	Width  int      `json:"width"`
	Height int      `json:"height"`
	Tags   []string `json:"tags"`
}

func (s *Setu) Content() ([]byte, error) {
	req := requests.Requests()
	resp, err := req.Get(s.Url)
	if err != nil {
		return nil, err
	}
	return resp.Content(), nil
}

type Response struct {
	Code        int     `json:"code"`
	Msg         string  `json:"msg"`
	Quota       int     `json:"quota"`
	QuotaMinTTL int     `json:"quota_min_ttl"`
	Count       int     `json:"count"`
	Data        []*Setu `json:"data"`
}

func LoliconAppSetu(apikey string, R18 R18Type, keyword string, num int) (*Response, error) {
	params, err := utils.ToParams(&Request{
		Apikey:   apikey,
		R18:      int(R18),
		Keyword:  keyword,
		Num:      num,
		Proxy:    "i.pixiv.cat",
		Size1200: true,
	})
	if err != nil {
		return nil, err
	}
	req := requests.Requests()
	req.SetTimeout(5 * time.Second)
	resp, err := req.Get(Host, params)
	if err != nil {
		return nil, err
	}
	apiResp := new(Response)
	err = resp.Json(apiResp)
	if err != nil {
		return nil, err
	}
	if apiResp.Code != 0 {
		return nil, errors.New(apiResp.Msg)
	}
	return apiResp, nil
}
