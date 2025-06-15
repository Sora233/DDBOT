package acfun

import (
	"bytes"
	"errors"
	"regexp"
	"time"

	"github.com/Sora233/DDBOT/v2/proxy_pool"
	"github.com/Sora233/DDBOT/v2/requests"
	"github.com/Sora233/DDBOT/v2/utils"
)

var livePageRegex = regexp.MustCompile("<script>window.__INITIAL_STATE__=(?P<json>.*?);\\(")

func LivePage(uid int64) (*LivePageResponse, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := LiveUrl(uid)
	var opts []requests.Option
	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.AddUAOption(),
		requests.TimeoutOption(time.Second*10),
	)
	var body = new(bytes.Buffer)
	err := requests.Get(url, nil, body, opts...)
	if err != nil {
		return nil, err
	}
	var b = body.Bytes()

	match := livePageRegex.FindSubmatch(b)
	if len(match) <= livePageRegex.SubexpIndex("json") {
		return nil, errors.New("no json data matched")
	}
	var result = new(LivePageResponse)
	if err = json.Unmarshal(match[livePageRegex.SubexpIndex("json")], result); err != nil {
		return nil, err
	}
	return result, nil
}
