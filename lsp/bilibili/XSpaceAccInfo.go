package bilibili

import (
	"fmt"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const (
	PathXSpaceAccInfo = "/x/space/acc/info"
)

type XSpaceAccInfoRequest struct {
	Mid      int64  `json:"mid"`
	Platform string `json:"platform"`
	Jsonp    string `json:"jsonp"`
	Token    string `json:"token"`
}

func XSpaceAccInfo(mid int64) (*XSpaceAccInfoResponse, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathXSpaceAccInfo)
	params, err := utils.ToParams(&XSpaceAccInfoRequest{
		Mid:      mid,
		Platform: "web",
		Jsonp:    "jsonp",
	})
	if err != nil {
		return nil, err
	}
	var opts = []requests.Option{
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.TimeoutOption(time.Second * 15),
		AddUAOption(),
		requests.HeaderOption("accept", "application/json"),
		requests.HeaderOption("accept-language", "zh-CN,zh;q=0.9"),
		requests.HeaderOption("origin", "https://space.bilibili.com"),
		requests.HeaderOption("referer", fmt.Sprintf("https://space.bilibili.com/%v", mid)),
		requests.RequestAutoHostOption(),
		requests.NotIgnoreEmptyOption(),
		delete412ProxyOption,
	}
	opts = append(opts, GetVerifyOption()...)
	xsai := new(XSpaceAccInfoResponse)
	err = requests.Get(url, params, xsai, opts...)
	if err != nil {
		return nil, err
	}
	return xsai, nil
}
