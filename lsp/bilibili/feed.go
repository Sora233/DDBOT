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

func FeedList(page int, pageSize int) (*FeedListResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathRoomInit)
	params, err := utils.ToParams(&RelationFeedRequest{
		Page:     page,
		PageSize: pageSize,
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
