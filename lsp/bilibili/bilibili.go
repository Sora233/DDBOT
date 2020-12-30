package bilibili

import (
	"github.com/asmcos/requests"
	"math/rand"
	"strings"
)

const DBKey = "bilibili"
const BaseHost = "https://api.bilibili.com"
const BaseLiveHost = "https://api.live.bilibili.com"
const BaseDynamicHost = "https://api.vc.bilibili.com"

var BasePath = map[string]string{
	PathRoomInit:               BaseLiveHost,
	PathSpaceAccInfo:           BaseHost,
	PathDynamicSrvSpaceHistory: BaseDynamicHost,
	PathGetRoomInfoOld:         BaseLiveHost,
}

var proxy []string

func BPath(path string) string {
	if strings.HasPrefix(path, "/") {
		return BasePath[path] + path
	} else {
		return BasePath[path] + "/" + path
	}
}

func SetProxy(p []string) {
	if p != nil && len(p) != 0 {
		proxy = p
	}
}

func GetBilibiliRequest() (*requests.Request, error) {
	req := requests.Requests()
	if len(proxy) > 0 {
		index := rand.Intn(len(proxy) + 1)
		if index != len(proxy) {
			req.Proxy(proxy[index])
		}
	}
	return req, nil
}
