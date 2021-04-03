package bilibili

import (
	"context"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"github.com/Sora233/Sora233-MiraiGo/utils"
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
		requests.ProxyOption(proxy_pool.PreferAny),
		AddUAOption(),
		requests.TimeoutOption(time.Second*3),
	)
	opts = append(opts, AddCookiesOption()...)
	resp, err := requests.Get(ctx, url, nil, 3, opts...)
	if err != nil {
		return nil, err
	}
	getAttentionListResp := new(GetAttentionListResponse)
	err = resp.Json(getAttentionListResp)
	if err != nil {
		logger.WithField("content", string(resp.Content())).Errorf("GetAttentionList response json failed")
		return nil, err
	}
	if getAttentionListResp.Code == -412 && resp.Proxy != "" {
		proxy_pool.Delete(resp.Proxy)
	}
	return getAttentionListResp, nil
}
