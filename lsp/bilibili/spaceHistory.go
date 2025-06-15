package bilibili

import (
	"fmt"
	"time"

	"github.com/Sora233/DDBOT/v2/proxy_pool"
	"github.com/Sora233/DDBOT/v2/requests"
	"github.com/Sora233/DDBOT/v2/utils"
)

const (
	PathDynamicSrvSpaceHistory = "/dynamic_svr/v1/dynamic_svr/space_history"
)

type DynamicSrvSpaceHistoryRequest struct {
	OffsetDynamicId int64 `json:"offset_dynamic_id"`
	HostUid         int64 `json:"host_uid"`
	NeedTop         int32 `json:"need_top"`
}

func DynamicSrvSpaceHistory(hostUid int64) (*DynamicSvrSpaceHistoryResponse, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathDynamicSrvSpaceHistory)
	params, err := utils.ToParams(&DynamicSrvSpaceHistoryRequest{
		HostUid: hostUid,
	})
	if err != nil {
		return nil, err
	}
	var opts = []requests.Option{
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.HeaderOption("Referer", fmt.Sprintf("https://space.bilibili.com/%v/", hostUid)),
		AddUAOption(),
		requests.TimeoutOption(time.Second * 15),
		delete412ProxyOption,
	}
	spaceHistoryResp := new(DynamicSvrSpaceHistoryResponse)
	err = requests.Get(url, params, spaceHistoryResp, opts...)
	if err != nil {
		return nil, err
	}
	return spaceHistoryResp, nil
}
