package bilibili

import (
	"context"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const (
	PathGetAttentionList = "/feed/v1/feed/get_attention_list"
)

func GetAttentionList() (*GetAttentionListResponse, error) {
	if !IsVerifyGiven() {
		return nil, ErrVerifyRequired
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathGetAttentionList)
	var opts []requests.Option
	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferNone),
		AddUAOption(),
		requests.TimeoutOption(time.Second*3),
	)
	opts = append(opts, GetVerifyOption()...)
	resp, err := requests.Get(ctx, url, nil, 3, opts...)
	if err != nil {
		return nil, err
	}
	getAttentionListResp := new(GetAttentionListResponse)
	err = resp.Json(getAttentionListResp)
	if err != nil {
		content, _ := resp.Content()
		logger.WithField("content", string(content)).Errorf("GetAttentionList response json failed")
		return nil, err
	}
	if getAttentionListResp.Code == -412 && resp.Proxy != "" {
		proxy_pool.Delete(resp.Proxy)
	}
	return getAttentionListResp, nil
}
