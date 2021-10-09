package douyu

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"strings"
	"time"
)

const (
	PathBetard = "/betard"
)

func Betard(id int64) (*BetardResponse, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := DouyuPath(PathBetard) + fmt.Sprintf("/%v", id)
	var opts = []requests.Option{
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.TimeoutOption(time.Second * 10),
		requests.RetryOption(3),
	}
	var body = new(bytes.Buffer)
	err := requests.Get(url, nil, body, opts...)
	if err != nil {
		return nil, err
	}
	betardResp := new(BetardResponse)
	err = json.Unmarshal(body.Bytes(), betardResp)
	if err != nil {
		if strings.Contains(body.String(), "没有开放") {
			return nil, errors.New("房间不存在")
		}
		if strings.Contains(body.String(), "已被关闭") {
			return nil, ErrRoomBanned
		}
		return nil, err
	}
	return betardResp, nil
}
