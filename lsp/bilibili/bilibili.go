package bilibili

import (
	"errors"
	"fmt"
	"github.com/Sora233/DDBOT/proxy_pool/requests"
	"strings"
)

const (
	Site         = "bilibili"
	BaseHost     = "https://api.bilibili.com"
	BaseLiveHost = "https://api.live.bilibili.com"
	BaseVCHost   = "https://api.vc.bilibili.com"
	VideoView    = "https://www.bilibili.com/video"
	DynamicView  = "https://t.bilibili.com"
)

var BasePath = map[string]string{
	PathRoomInit:               BaseLiveHost,
	PathSpaceAccInfo:           BaseHost,
	PathDynamicSrvSpaceHistory: BaseVCHost,
	PathGetRoomInfoOld:         BaseLiveHost,
	PathDynamicSrvDynamicNew:   BaseVCHost,
	PathRelationModify:         BaseHost,
	PathRelationFeedList:       BaseLiveHost,
	PathGetAttentionList:       BaseVCHost,
}

var (
	ErrVerifyRequired = errors.New("verify required")
	SESSDATA          string
	biliJct           string
)

func BPath(path string) string {
	if strings.HasPrefix(path, "/") {
		return BasePath[path] + path
	} else {
		return BasePath[path] + "/" + path
	}
}

func BVIDUrl(bvid string) string {
	return fmt.Sprintf("%v/%v", VideoView, bvid)
}

func DynamicUrl(dynamicIdStr string) string {
	return fmt.Sprintf("%v/%v", DynamicView, dynamicIdStr)
}

func SetVerify(_SESSDATA string, _biliJct string) {
	SESSDATA = _SESSDATA
	biliJct = _biliJct
}

func AddCookiesOption() []requests.Option {
	return []requests.Option{requests.CookieOption("SESSDATA", SESSDATA), requests.CookieOption("bili_jct", biliJct)}
}

func AddUAOption() requests.Option {
	return requests.HeaderOption("user-agent", fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36'"))
}

func IsVerifyGiven() bool {
	if SESSDATA == "" || biliJct == "" {
		return false
	}
	return true
}
