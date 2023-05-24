package bilibili

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"github.com/guonaihong/gout"
	"github.com/samber/lo"
	"sort"
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

func signWbi(params gout.H) {
	imgKey, subKey := getWbi()
	var sb strings.Builder
	var orig = imgKey + subKey
	for _, r := range mixinKeyEncTab {
		sb.WriteRune(rune(orig[r]))
	}
	salt := sb.String()[:32]
	params["wts"] = time.Now().Unix()
	p := lo.MapToSlice(params, func(key string, value any) lo.Tuple2[string, any] {
		return lo.Tuple2[string, any]{A: key, B: value}
	})
	sort.Slice(p, func(i, j int) bool {
		return p[i].A < p[j].A
	})

	query := strings.Join(lo.Map(p, func(item lo.Tuple2[string, any], _ int) string {
		return fmt.Sprintf("%v=%v", item.A, item.B)
	}), "&")
	hash := md5.New()
	hash.Write([]byte(query + salt))
	wbiSign := hex.EncodeToString(hash.Sum(nil))
	params["w_rid"] = wbiSign
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
