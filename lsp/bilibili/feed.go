package bilibili

import (
	"context"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	"time"
)

const (
	PathRelationFeed = "/relation/v1/feed/feed_list"
)

type RelationFeedRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pagesize"`
}

type FeedOpt func(map[string]int)

func FeedPageOpt(page int) FeedOpt {
	return func(m map[string]int) {
		m["page"] = page
	}
}

func FeedPageSizeOpt(pageSize int) FeedOpt {
	return func(m map[string]int) {
		m["pageSize"] = pageSize
	}
}

func FeedList(opt ...FeedOpt) (*FeedListResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	var p = map[string]int{
		"page":     1,
		"pageSize": 30,
	}
	for _, o := range opt {
		o(p)
	}

	url := BPath(PathRoomInit)
	params, err := utils.ToParams(&RelationFeedRequest{
		Page:     p["page"],
		PageSize: p["pageSize"],
	})
	if err != nil {
		return nil, err
	}
	resp, err := requests.Get(ctx, url, params, 1, requests.ProxyOption(proxy_pool.PreferAny))
	if err != nil {
		return nil, err
	}
	rir := new(FeedListResponse)
	err = resp.Json(rir)
	if err != nil {
		return nil, err
	}
	if rir.Code == -412 && resp.Proxy != "" {
		proxy_pool.Delete(resp.Proxy)
	}
	return rir, nil
}
