package bilibili

import (
	"errors"
	"github.com/asmcos/requests"
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
	url := BPath(PathDynamicSrvSpaceHistory)
	params, err := BGetRequestToParams(&DynamicSrvSpaceHistoryRequest{
		HostUid:         hostUid,
		OffsetDynamicId: 0,
		NeedTop:         0,
	})
	if err != nil {
		return nil, err
	}
	resp, err := requests.Get(url, params)
	if err != nil {
		return nil, err
	}
	spaceHistoryResp := new(DynamicSvrSpaceHistoryResponse)
	err = resp.Json(spaceHistoryResp)
	if err != nil {
		return nil, err
	}
	if spaceHistoryResp.GetCode() != 0 {
		return nil, errors.New(spaceHistoryResp.Message)
	}
	return spaceHistoryResp, nil
}
