package douyu

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"strings"
)

const (
	PathBetard = "/betard"
)

func Betard(id int64) (*BetardResponse, error) {
	url := DouyuPath(PathBetard) + fmt.Sprintf("/%v", id)
	resp, err := requests.Get(url, nil)
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
		proxy_pool.Delete(resp.Proxy)
		return nil, err
	}
	return betardResp, nil
}
