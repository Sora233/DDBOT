package bilibili

import (
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
}

func BPath(path string) string {
	if strings.HasPrefix(path, "/") {
		return BasePath[path] + path
	} else {
		return BasePath[path] + "/" + path
	}
}
