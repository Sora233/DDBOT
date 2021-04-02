package bilibili

import (
	"context"
	"fmt"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	"time"
)

const (
	PathDynamicSrvDynamicNew = "/dynamic_svr/v1/dynamic_svr/dynamic_new"
)

type DynamicSrvDynamicNewRequest struct {
	Platform string `json:"platform"`
	From     string `json:"from"`
	TypeList string `json:"type_list"`
}

func DynamicSrvDynamicNew() (*DynamicSvrDynamicNewResponse, error) {
	if !IsVerifyGiven() {
		return nil, ErrVerifyRequired
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathDynamicSrvDynamicNew)
	params, err := utils.ToParams(&DynamicSrvDynamicNewRequest{
		Platform: "web",
		From:     "weball",
		TypeList: "268435455", // 会变吗？
	})
	if err != nil {
		return nil, err
	}
	var opts []requests.Option
	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferAny),
		requests.HeaderOption("origin", fmt.Sprintf("https://t.bilibili.com")),
		requests.HeaderOption("referer", fmt.Sprintf("https://t.bilibili.com")),
		AddUAOption(),
		requests.TimeoutOption(time.Second*5),
	)
	opts = append(opts, AddCookiesOption()...)
	resp, err := requests.Get(ctx, url, params, 1, opts...)
	if err != nil {
		return nil, err
	}
	dynamicNewResp := new(DynamicSvrDynamicNewResponse)
	err = resp.Json(dynamicNewResp)
	if err != nil {
		logger.WithField("content", string(resp.Content())).Errorf("DynamicSrvDynamicNew response json failed")
		return nil, err
	}
	if dynamicNewResp.Code == -412 && resp.Proxy != "" {
		proxy_pool.Delete(resp.Proxy)
	}
	return dynamicNewResp, nil
}
