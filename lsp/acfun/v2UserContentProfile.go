package acfun

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const (
	PathV2UserContentProfile = "/v2/user/content/profile"
)

type V2UserContentProfileRequest struct {
	UserId int64 `json:"userId"`
}

func V2UserContentProfile(userId int64) (*V2UserContentProfileResponse, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := APath(PathV2UserContentProfile)
	params, err := utils.ToParams(&V2UserContentProfileRequest{UserId: userId})
	if err != nil {
		return nil, err
	}
	var opts []requests.Option
	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.AddUAOption(),
		requests.TimeoutOption(time.Second*10),
	)
	opts = append(opts, AcfunHeaderOption()...)
	v2UserContentProfileResp := new(V2UserContentProfileResponse)
	err = requests.Get(url, params, &v2UserContentProfileResp, opts...)
	if err != nil {
		return nil, err
	}
	return v2UserContentProfileResp, nil
}
