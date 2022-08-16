package bilibili

import (
	"fmt"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const (
	PathDynamicSrvDynamicHistory = "/dynamic_svr/v1/dynamic_svr/dynamic_history"
)

type DynamicSrvDynamicHistoryRequest struct {
	OffsetDynamicId string `json:"offset_dynamic_id"`
	TypeList        string `json:"type_list"`
}

func DynamicSvrDynamicHistory(offsetDynamicId string) (*DynamicSvrDynamicHistoryResponse, error) {
	if !IsVerifyGiven() {
		return nil, ErrVerifyRequired
	}
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathDynamicSrvDynamicHistory)
	params, err := utils.ToParams(&DynamicSrvDynamicHistoryRequest{
		OffsetDynamicId: offsetDynamicId,
		TypeList:        "268435455", // 会变吗？
	})
	if err != nil {
		return nil, err
	}
	var opts []requests.Option
	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.HeaderOption("origin", fmt.Sprintf("https://t.bilibili.com")),
		requests.HeaderOption("referer", fmt.Sprintf("https://t.bilibili.com")),
		AddUAOption(),
		requests.TimeoutOption(time.Second*10),
		delete412ProxyOption,
	)
	opts = append(opts, GetVerifyOption()...)
	dynamicHistoryResp := new(DynamicSvrDynamicHistoryResponse)
	err = requests.Get(url, params, dynamicHistoryResp, opts...)
	if err != nil {
		return nil, err
	}
	return dynamicHistoryResp, nil
}
