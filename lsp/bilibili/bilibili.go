package bilibili

import (
	"errors"
	"fmt"
	"github.com/Sora233/DDBOT/proxy_pool/requests"
	"strconv"
	"strings"
	"sync"
)

const (
	Site         = "bilibili"
	BaseHost     = "https://api.bilibili.com"
	BaseLiveHost = "https://api.live.bilibili.com"
	BaseVCHost   = "https://api.vc.bilibili.com"
	VideoView    = "https://www.bilibili.com/video"
	DynamicView  = "https://t.bilibili.com"
	PassportHost = "https://passport.bilibili.com"
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
	PathOAuth2GetKey:           PassportHost,
	PathV3OAuth2Login:          PassportHost,
}

var (
	ErrVerifyRequired = errors.New("verify required")
	SESSDATA          string
	biliJct           string

	mux      = new(sync.RWMutex)
	username string
	password string
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

func SetAccount(_username string, _password string) {
	username = _username
	password = _password
}

func AddVerifyOption() []requests.Option {
	if IsCookieGiven() {
		return []requests.Option{requests.CookieOption("SESSDATA", SESSDATA), requests.CookieOption("bili_jct", biliJct)}
	}

	mux.Lock()
	defer mux.Unlock()

	if IsCookieGiven() {
		return []requests.Option{requests.CookieOption("SESSDATA", SESSDATA), requests.CookieOption("bili_jct", biliJct)}
	}

	if !IsAccountGiven() {
		logger.Errorf("AddVerifyOption error - 未设置cookie和帐号")
		return nil
	} else {
		logger.Debug("AddVerifyOption 使用帐号刷新cookie")
		cookieInfo, err := freshAccountCookieInfo()
		if err != nil {
			logger.Errorf("b站登陆失败，请手动指定cookie配置 - freshAccountCookieInfo error %v", err)
		} else {
			logger.Debug("b站登陆成功 - freshAccountCookieInfo ok")
			for _, cookie := range cookieInfo.GetCookies() {
				if cookie.GetName() == "SESSDATA" {
					SESSDATA = cookie.GetValue()
					logger.Debug("使用cookieInfo设置 SESSDATA")
				}
				if cookie.GetName() == "bili_jct" {
					biliJct = cookie.GetValue()
					logger.Debug("使用cookieInfo设置 bili_jct")
				}
			}
			if !IsCookieGiven() {
				logger.Errorf("b站登陆成功，但是设置cookie失败，如果发现这个问题，请反馈给开发者。")
			}
		}
		if !IsCookieGiven() {
			// 设置错误cookie防止反复登陆
			SESSDATA = "wrong"
			biliJct = "wrong"
		}
	}
	return []requests.Option{requests.CookieOption("SESSDATA", SESSDATA), requests.CookieOption("bili_jct", biliJct)}
}

func AddUAOption() requests.Option {
	return requests.AddUAOption()
}

func AddReferOption(refer ...string) requests.Option {
	if len(refer) == 0 {
		return requests.HeaderOption("Referer", "https://www.bilibili.com/")
	}
	return requests.HeaderOption("Referer", refer[0])
}

func IsVerifyGiven() bool {
	if IsCookieGiven() || IsAccountGiven() {
		return true
	}
	return false
}

func IsCookieGiven() bool {
	if SESSDATA == "" || biliJct == "" {
		return false
	}
	return true
}

func IsAccountGiven() bool {
	if username == "" {
		return false
	}
	return true
}

func ParseUid(s string) (int64, error) {
	s = strings.TrimLeft(s, "UID:")
	return strconv.ParseInt(s, 10, 64)
}

func freshAccountCookieInfo() (*LoginResponse_Data_CookieInfo, error) {
	logger.Debug("freshAccountCookieInfo")
	if !IsAccountGiven() {
		return nil, errors.New("未设置帐号")
	}
	if ci, err := GetCookieInfo(username); err == nil {
		logger.Debug("GetCookieInfo from db ok")
		return ci, nil
	}
	logger.Debug("login to fresh cookie")
	resp, err := Login(username, password)
	if err != nil {
		logger.Errorf("Login error %v", err)
		return nil, err
	}
	if resp.GetCode() != 0 {
		logger.Errorf("Login code %v", resp.GetCode())
		return nil, fmt.Errorf("login code %v", resp.GetCode())
	}
	logger.Debug("login success")
	if err = SetCookieInfo(username, resp.GetData().GetCookieInfo()); err != nil {
		logger.Errorf("SetCookieInfo error %v", err)
	} else {
		logger.Debug("SetCookieInfo ok")
	}
	return resp.GetData().GetCookieInfo(), nil
}
