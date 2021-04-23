package bilibili

import (
	"context"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/proxy_pool/requests"
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
	ActBlock          = 5
	ActUnblock        = 6
	ActRemoveFollower = 7
)

type RelationModifyRequest struct {
	Fid   int64  `json:"fid"`
	Act   int    `json:"act"`
	ReSrc int    `json:"re_src"`
	Csrf  string `json:"csrf"`
}

func RelationModify(fid int64, act int) (*RelationModifyResponse, error) {
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
	var err error
	url := BPath(PathRelationModify)
	formRequest := &RelationModifyRequest{
		Fid:   fid,
		Act:   act,
		ReSrc: 11,
		Csrf:  biliJct,
	}
	form, err := utils.ToDatas(formRequest)
	if err != nil {
		return nil, err
	}
	var opts []requests.Option
	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.TimeoutOption(time.Second*5),
		AddUAOption(),
	)
	opts = append(opts, AddCookiesOption()...)
	resp, err := requests.Post(ctx, url, form, 1,
		opts...,
	)
	if err != nil {
		return nil, err
	}
	rmr := new(RelationModifyResponse)
	err = resp.Json(rmr)
	if err != nil {
		return nil, err
	}
	if rmr.Code == -412 && resp.Proxy != "" {
		proxy_pool.Delete(resp.Proxy)
	}
	return rmr, nil

}
