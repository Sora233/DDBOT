package bilibili

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const (
	PathRelationModify = "/x/relation/modify"
)

const (
	ActSub            = 1
	ActUnsub          = 2
	ActHiddenSub      = 3
	ActUnhiddenSub    = 4
	ActBlock          = 5
	ActUnblock        = 6
	ActRemoveFollower = 7
)

type RelationModifyRequest struct {
	Fid  int64  `json:"fid"`
	Act  int    `json:"act"`
	Csrf string `json:"csrf"`
}

func RelationModify(fid int64, act int) (*RelationModifyResponse, error) {
	if !IsVerifyGiven() {
		return nil, ErrVerifyRequired
	}
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	var err error
	url := BPath(PathRelationModify)
	formRequest := &RelationModifyRequest{
		Fid:  fid,
		Act:  act,
		Csrf: GetVerifyBiliJct(),
	}
	form, err := utils.ToParams(formRequest)
	if err != nil {
		return nil, err
	}
	var opts []requests.Option
	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.TimeoutOption(time.Second*10),
		AddUAOption(),
		delete412ProxyOption,
	)
	opts = append(opts, GetVerifyOption()...)
	rmr := new(RelationModifyResponse)
	err = requests.PostWWWForm(url, form, rmr, opts...)
	if err != nil {
		return nil, err
	}
	return rmr, nil
}
