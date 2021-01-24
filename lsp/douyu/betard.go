package douyu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	"strings"
	"time"
)

const (
	PathBetard = "/betard"
)

func Betard(id int64) (*BetardResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := DouyuPath(PathBetard) + fmt.Sprintf("/%v", id)
	resp, err := requests.Get(ctx, url, nil, 3, requests.ProxyOption(proxy_pool.PreferNone))
	if err != nil {
		return nil, err
	}
	betardResp := new(BetardResponse)
	content := resp.Content()
	err = json.Unmarshal(content, betardResp)
	if err != nil {
		if strings.Contains(string(content), "没有开放") {
			return nil, errors.New("房间不存在")
		}
		if resp.Proxy != "" {
			proxy_pool.Delete(resp.Proxy)
		}
		return nil, err
	}
	return betardResp, nil
}
