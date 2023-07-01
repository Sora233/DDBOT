package bilibili

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"github.com/samber/lo"
	"sort"
	"strconv"
	"strings"
	"time"
)

const PathXWebInterfaceNav = "/x/web-interface/nav"

func refreshNavWbi() {
	resp, err := XWebInterfaceNav(false)
	if err != nil {
		logger.Errorf("bilibili: refreshNavWbi error %v", err)
		return
	}
	wbiImg := resp.GetData().GetWbiImg()
	if wbiImg != nil {
		wbi.Store(wbiImg)
	}
	logger.Trace("bilibili: refreshNavWbi ok")
}

func getWbi() (imgKey string, subKey string) {
	wbi := wbi.Load()
	getKey := func(url string) string {
		path, _ := lo.Last(strings.Split(url, "/"))
		key := strings.Split(path, ".")[0]
		return key
	}
	imgKey = getKey(wbi.ImgUrl)
	subKey = getKey(wbi.SubUrl)
	return
}
func getMixinKey(orig string) string {
	var str strings.Builder
	for _, v := range mixinKeyEncTab {
		if v < len(orig) {
			str.WriteByte(orig[v])
		}
	}
	return str.String()[:32]
}

func signWbi(params map[string]string) map[string]string {
	imgKey, subKey := getWbi()
	mixinKey := getMixinKey(imgKey + subKey)
	currTime := strconv.FormatInt(time.Now().Unix(), 10)
	params["wts"] = currTime
	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// Build URL parameters
	var str strings.Builder
	for _, k := range keys {
		str.WriteString(fmt.Sprintf("%s=%s&", k, params[k]))
	}
	query := strings.TrimSuffix(str.String(), "&")
	hash := md5.Sum([]byte(query + mixinKey))
	params["w_rid"] = hex.EncodeToString(hash[:])
	return params
}

func XWebInterfaceNav(login bool) (*WebInterfaceNavResponse, error) {
	if login && !IsVerifyGiven() {
		return nil, ErrVerifyRequired
	}
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	path := BPath(PathXWebInterfaceNav)
	var opts = []requests.Option{
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.TimeoutOption(time.Second * 15),
		AddUAOption(),
		delete412ProxyOption,
	}
	if login && getVerify() != nil {
		opts = append(opts, getVerify().VerifyOpts...)
	}
	xwin := new(WebInterfaceNavResponse)
	err := requests.Get(path, nil, xwin, opts...)
	if err != nil {
		return nil, err
	}
	return xwin, nil
}
