package lolicon_pool

import (
	"context"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const Host = "https://api.lolicon.app/setu"

type R18Type int

const (
	R18Off R18Type = iota
	R18On
	//R18Mix
)

func (r R18Type) String() string {
	switch r {
	case R18Off:
		return "R18Off"
	case R18On:
		return "R18On"
	//case R18Mix:
	//	return "R18Mix"
	default:
		return "Unknown"
	}
}

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
	return utils.ImageGet(s.Url, proxy_pool.PreferOversea, requests.HeaderOption("referer", "https://www.pixiv.net"))
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
		Proxy:    "disable",
		Size1200: true,
	})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	resp, err := requests.Get(ctx, Host, params, 3, requests.ProxyOption(proxy_pool.PreferOversea))
	if err != nil {
		return nil, err
	}
	apiResp := new(Response)
	err = resp.Json(apiResp)
	if err != nil {
		return nil, err
	}
	return apiResp, nil
}
