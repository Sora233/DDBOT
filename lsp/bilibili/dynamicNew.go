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

func DynamicSrvDynamicNew(SESSDATA string, biliJct string) (*DynamicSvrDynamicNewResponse, error) {
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
	resp, err := requests.Get(ctx, url, params, 1,
		requests.ProxyOption(proxy_pool.PreferAny),
		requests.HeaderOption("origin", fmt.Sprintf("https://t.bilibili.com")),
		requests.HeaderOption("referer", fmt.Sprintf("https://t.bilibili.com")),
		requests.HeaderOption("user-agent", fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36'")),
		requests.TimeoutOption(time.Second*5),
	)
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
