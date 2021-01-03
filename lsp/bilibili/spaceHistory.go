package bilibili

import (
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	"time"
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
	resp, err := requests.Get(url, params, 3)
	if err != nil {
		return nil, err
	}
	spaceHistoryResp := new(DynamicSvrSpaceHistoryResponse)
	err = resp.Json(spaceHistoryResp)
	if err != nil {
		return nil, err
	}
	if spaceHistoryResp.Code == -412 {
		proxy_pool.Delete(resp.Proxy)
	}
	return spaceHistoryResp, nil
}
