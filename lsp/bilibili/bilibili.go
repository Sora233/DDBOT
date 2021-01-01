package bilibili

import (
	"strings"
)

const (
	Site            = "bilibili"
	BaseHost        = "https://api.bilibili.com"
	BaseLiveHost    = "https://api.live.bilibili.com"
	BaseDynamicHost = "https://api.vc.bilibili.com"
)

var BasePath = map[string]string{
	PathRoomInit:               BaseLiveHost,
	PathSpaceAccInfo:           BaseHost,
	PathDynamicSrvSpaceHistory: BaseDynamicHost,
	PathGetRoomInfoOld:         BaseLiveHost,
}

func BPath(path string) string {
	if strings.HasPrefix(path, "/") {
		return BasePath[path] + path
	} else {
		return BasePath[path] + "/" + path
	}
}
