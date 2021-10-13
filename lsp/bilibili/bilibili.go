package bilibili

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Sora233/DDBOT/requests"
	jsoniter "github.com/json-iterator/go"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var BasePath = map[string]string{
	PathRoomInit:                 BaseLiveHost,
	PathXSpaceAccInfo:            BaseHost,
	PathDynamicSrvSpaceHistory:   BaseVCHost,
	PathGetRoomInfoOld:           BaseLiveHost,
	PathDynamicSrvDynamicNew:     BaseVCHost,
	PathRelationModify:           BaseHost,
	PathRelationFeedList:         BaseLiveHost,
	PathGetAttentionList:         BaseVCHost,
	PathPassportLoginWebKey:      PassportHost,
	PathPassportLoginOAuth2Login: PassportHost,
	PathXRelationStat:            BaseHost,
}

type VerifyInfo struct {
	SESSDATA   string
	BiliJct    string
	VerifyOpts []requests.Option
}

var (
	ErrVerifyRequired = errors.New("verify required")
	// atomicVerifyInfo is a *VerifyInfo
	atomicVerifyInfo atomic.Value

	mux      = new(sync.Mutex)
	username string
	password string
)

func Init() {
	var (
		SESSDATA = config.GlobalConfig.GetString("bilibili.SESSDATA")
		biliJct  = config.GlobalConfig.GetString("bilibili.bili_jct")
	)
	if len(SESSDATA) != 0 && len(biliJct) != 0 {
		SetVerify(SESSDATA, biliJct)
	}
	SetAccount(config.GlobalConfig.GetString("bilibili.account"), config.GlobalConfig.GetString("bilibili.password"))

}

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
	atomicVerifyInfo.Store(&VerifyInfo{
		SESSDATA:   _SESSDATA,
		BiliJct:    _biliJct,
		VerifyOpts: []requests.Option{requests.CookieOption("SESSDATA", _SESSDATA), requests.CookieOption("bili_jct", _biliJct)},
	})
}

func getVerify() *VerifyInfo {
	return atomicVerifyInfo.Load().(*VerifyInfo)
}

func SetAccount(_username string, _password string) {
	username = _username
	password = _password
}

func GetVerifyOption() []requests.Option {
	info := GetVerifyInfo()
	if info == nil {
		return nil
	}
	return info.VerifyOpts
}

func GetVerifyBiliJct() string {
	info := GetVerifyInfo()
	if info == nil {
		return ""
	}
	return info.BiliJct
}

func GetVerifyInfo() *VerifyInfo {
	if IsCookieGiven() {
		return getVerify()
	}

	mux.Lock()
	defer mux.Unlock()

	if IsCookieGiven() {
		return getVerify()
	}

	if !IsAccountGiven() {
		logger.Errorf("GetVerifyInfo error - 未设置cookie和帐号")
		return nil
	} else {
		var (
			SESSDATA string
			biliJct  string
			expire   int64 = -1
			ok       bool
		)
		logger.Debug("GetVerifyInfo 使用帐号刷新cookie")
		cookieInfo, err := freshAccountCookieInfo()
		if err != nil {
			logger.Errorf("b站登陆失败，请手动指定cookie配置 - freshAccountCookieInfo error %v", err)
		} else {
			logger.Debug("b站登陆成功 - freshAccountCookieInfo ok")
			for _, cookie := range cookieInfo.GetCookies() {
				if expire == -1 || expire > cookie.GetExpires() {
					expire = cookie.GetExpires()
				}
				if cookie.GetName() == "SESSDATA" {
					SESSDATA = cookie.GetValue()
					logger.Debug("使用cookieInfo设置 SESSDATA")
				}
				if cookie.GetName() == "bili_jct" {
					biliJct = cookie.GetValue()
					logger.Debug("使用cookieInfo设置 bili_jct")
				}
			}
			if len(SESSDATA) == 0 || len(biliJct) == 0 {
				logger.Errorf("b站登陆成功，但是设置cookie失败，如果发现这个问题，请反馈给开发者。")
			} else {
				ok = true
				SetVerify(SESSDATA, biliJct)
				if expire > 0 {
					// 尝试cookie过期之后自动刷新
					// 但是登陆方法有效期应该比cookie有效期短
					// 这样做真的值得吗
					t := time.Until(time.Unix(expire, 0))
					if t <= 0 {
						logger.Info("当前Cookie已过期，即将自动刷新")
						atomicVerifyInfo.Store(new(VerifyInfo))
					} else {
						logger.Infof("当前Cookie还有 %v 时间过期，过期后会自动刷新", t)
						time.AfterFunc(t, func() {
							atomicVerifyInfo.Store(new(VerifyInfo))
						})
					}
				}
			}
		}
		if !ok {
			SetVerify("wrong", "wrong")
		}
	}
	return getVerify()
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
	v := atomicVerifyInfo.Load()
	if v == nil {
		return false
	}
	info, ok := v.(*VerifyInfo)
	if !ok {
		return false
	}
	return len(info.VerifyOpts) > 0
}

func IsAccountGiven() bool {
	if username == "" {
		return false
	}
	return true
}

func ParseUid(s string) (int64, error) {
	s = strings.TrimPrefix(strings.ToUpper(s), "UID:")
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
		logger.Errorf("Login error %v - %v", resp.GetCode(), resp.GetMessage())
		return nil, fmt.Errorf("login error %v - %v", resp.GetCode(), resp.GetMessage())
	}
	logger.Debug("login success")
	if err = SetCookieInfo(username, resp.GetData().GetCookieInfo()); err != nil {
		logger.Errorf("SetCookieInfo error %v", err)
	} else {
		logger.Debug("SetCookieInfo ok")
	}
	return resp.GetData().GetCookieInfo(), nil
}
