package bilibili

import (
	"errors"
	"fmt"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/MiraiGo-Template/config"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/atomic"
	"strconv"
	"strings"
	"sync"
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

	CompactExpireTime = time.Minute * 60
	// followerNotifyCap 提示粉丝数过少的阈值
	followerNotifyCap = 50
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var BasePath = map[string]string{
	PathXSpaceAccInfo:            BaseHost,
	PathDynamicSrvSpaceHistory:   BaseVCHost,
	PathDynamicSrvDynamicNew:     BaseVCHost,
	PathRelationModify:           BaseHost,
	PathRelationFeedList:         BaseLiveHost,
	PathGetAttentionList:         BaseVCHost,
	PathPassportLoginWebKey:      PassportHost,
	PathPassportLoginOAuth2Login: PassportHost,
	PathXRelationStat:            BaseHost,
	PathXWebInterfaceNav:         BaseHost,
	PathDynamicSrvDynamicHistory: BaseVCHost,
}

type VerifyInfo struct {
	SESSDATA   string
	BiliJct    string
	VerifyOpts []requests.Option
}

type ICode interface {
	GetCode() int32
}

var (
	ErrVerifyRequired = errors.New("账号信息缺失")
	atomicVerifyInfo  atomic.Pointer[VerifyInfo]

	mux                  = new(sync.Mutex)
	username             string
	password             string
	accountUid           atomic.Int64
	wbi                  atomic.Pointer[WebInterfaceNavResponse_Data_WbiImg]
	delete412ProxyOption = func() requests.Option {
		return requests.ProxyCallbackOption(func(out interface{}, proxy string) {
			if out == nil {
				return
			}
			if c, ok := out.(ICode); ok && (c.GetCode() == -412 || c.GetCode() == 412) {
				proxy_pool.Delete(proxy)
			}
		})
	}()
	mixinKeyEncTab = []int{
		46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
		33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
		61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
		36, 20, 34, 44, 52,
	}
)

func Init() {
	var (
		SESSDATA = config.GlobalConfig.GetString("bilibili.SESSDATA")
		biliJct  = config.GlobalConfig.GetString("bilibili.bili_jct")
	)
	if len(SESSDATA) != 0 && len(biliJct) != 0 {
		SetVerify(SESSDATA, biliJct)
		FreshSelfInfo()
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
	return atomicVerifyInfo.Load()
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
		logger.Trace("GetVerifyInfo error - 未设置cookie和帐号")
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
				FreshSelfInfo()
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
			FreshSelfInfo()
		}
	}
	return getVerify()
}

func FreshSelfInfo() {
	navResp, err := XWebInterfaceNav(true)
	if err != nil {
		logger.Errorf("获取个人信息失败 - %v，b站功能可能无法使用", err)
	} else {
		if navResp.GetCode() != 0 {
			logger.Errorf("获取个人信息失败 - %v %v", navResp.GetCode(), navResp.GetMessage())
		} else {
			if navResp.GetData().GetIsLogin() {
				logger.Infof("B站启动成功，当前使用账号：UID:%v %v LV%v %v",
					navResp.GetData().GetMid(),
					navResp.GetData().GetVipLabel().GetText(),
					navResp.GetData().GetLevelInfo().GetCurrentLevel(),
					navResp.GetData().GetUname())
				if navResp.GetData().GetLevelInfo().GetCurrentLevel() >= 5 {
					logger.Warnf("注意：当前使用的B站账号为5级或以上，请注意使用b站订阅时，该账号会自动关注订阅的目标用户！" +
						"如果不想新增关注，请使用小号。")
				}
				accountUid.Store(navResp.GetData().GetMid())
				return
			} else {
				logger.Errorf("账号未登陆")
			}
		}
	}
	accountUid.Store(0)
}

func AddUAOption() requests.Option {
	return requests.AddRandomUAOption(requests.Computer)
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
	return len(v.VerifyOpts) > 0
}

func IsAccountGiven() bool {
	if username == "" {
		return false
	}
	return true
}

func ParseUid(s string) (int64, error) {
	// 手机端复制的时候会带上UID:前缀，所以支持这个格式
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
	if resp.GetData().GetStatus() != 0 {
		logger.Errorf("Login status error %v - %v", resp.GetData().GetStatus(), resp.GetData().GetMessage())
		return nil, fmt.Errorf("login status error %v - %v", resp.GetData().GetStatus(), resp.GetData().GetMessage())
	}
	logger.Debug("login success")
	if err = SetCookieInfo(username, resp.GetData().GetCookieInfo()); err != nil {
		logger.Errorf("SetCookieInfo error %v", err)
	} else {
		logger.Debug("SetCookieInfo ok")
	}
	return resp.GetData().GetCookieInfo(), nil
}
