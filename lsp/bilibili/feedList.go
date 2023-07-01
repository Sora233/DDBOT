package bilibili

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const (
	PathRelationFeedList = "/xlive/web-ucenter/v1/xfetter/FeedList"
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
	if !IsVerifyGiven() {
		return nil, ErrVerifyRequired
	}
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

	url := BPath(PathRelationFeedList)
	params, err := utils.ToDatas(&RelationFeedRequest{
		Page:     p["page"],
		PageSize: p["pageSize"],
	})
	if err != nil {
		return nil, err
	}
	signWbi(params)
	var opts []requests.Option
	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferNone),
		AddUAOption(),
		requests.TimeoutOption(time.Second*10),
	)
	opts = append(opts, GetVerifyOption()...)
	flr := new(FeedListResponse)
	err = requests.Get(url, params, flr, opts...)
	if err != nil {
		return nil, err
	}
	return flr, nil
}
