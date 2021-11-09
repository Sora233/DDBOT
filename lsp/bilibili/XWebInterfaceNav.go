package bilibili

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const PathXWebInterfaceNav = "/x/web-interface/nav"

func XWebInterfaceNav() (*WebInterfaceNavResponse, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathXWebInterfaceNav)
	var opts = []requests.Option{
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.TimeoutOption(time.Second * 15),
		AddUAOption(),
		delete412ProxyOption,
	}
	if getVerify() != nil {
		opts = append(opts, getVerify().VerifyOpts...)
	}
	xwin := new(WebInterfaceNavResponse)
	err := requests.Get(url, nil, xwin, opts...)
	if err != nil {
		return nil, err
	}
	return xwin, nil
}
