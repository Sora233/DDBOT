package weibo

import (
	"fmt"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"github.com/guonaihong/gout"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	pathPassportGenvisitor = "https://passport.weibo.com/visitor/genvisitor"
	pathPassportVisitor    = "https://passport.weibo.com/visitor/visitor"
	pathLoginVisitor       = "https://login.sina.com.cn/visitor/visitor"
)

var (
	genvisitorRegex = regexp.MustCompile(`\((.*)\)`)
)

func genvisitor(externalOpts ...requests.Option) (*GenVisitorResponse, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	params := gout.H{
		"cb": "gen_callback",
		"fp": `{"os":"1","browser":"Gecko60,0,0,0","fonts":"undefined","screenInfo":"1536*864*24","plugins":""}`,
	}
	var opts = []requests.Option{
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.AddUAOption(),
		requests.TimeoutOption(time.Second * 10),
	}
	opts = append(opts, externalOpts...)
	var result string
	err := requests.Get(pathPassportGenvisitor, params, &result, opts...)
	if err != nil {
		return nil, err
	}
	submatch := genvisitorRegex.FindStringSubmatch(result)
	if len(submatch) < 2 {
		logger.Errorf("genvisitorRegex submatch not found")
		return nil, fmt.Errorf("genvisitor response regex extract failed")
	}
	var resp = new(GenVisitorResponse)
	err = json.Unmarshal([]byte(submatch[1]), resp)
	if err != nil {
		logger.WithField("Content", submatch[1]).Errorf("genvisitor data unmarshal error %v", err)
		resp = nil
	}
	return resp, err
}

func passportVisitor(params gout.H, externalOpts ...requests.Option) (*VisitorIncarnateResponse, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	var opts = []requests.Option{
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.AddUAOption(),
		requests.TimeoutOption(time.Second * 10),
	}
	opts = append(opts, externalOpts...)
	var result string
	err := requests.Get(pathPassportVisitor, params, &result, opts...)
	if err != nil {
		return nil, err
	}
	submatch := genvisitorRegex.FindStringSubmatch(result)
	if len(submatch) < 2 {
		logger.Errorf("passportVisitor submatch not found")
		return nil, fmt.Errorf("passportVisitor response regex extract failed")
	}
	var resp = new(VisitorIncarnateResponse)
	err = json.Unmarshal([]byte(submatch[1]), resp)
	if err != nil {
		logger.WithField("Content", submatch[1]).Errorf("passportVisitor data unmarshal error %v", err)
		resp = nil
	}
	return resp, err
}

func FreshCookie() ([]*http.Cookie, error) {
	jar, _ := cookiejar.New(nil)
	genVisitorResp, err := genvisitor(requests.WithCookieJar(jar))
	if err != nil {
		logger.Errorf("genvisitor error %v", err)
		return nil, err
	}
	if genVisitorResp.GetRetcode() != 20000000 || !strings.Contains(genVisitorResp.GetMsg(), "succ") {
		logger.WithFields(logrus.Fields{
			"Msg":     genVisitorResp.GetMsg(),
			"Retcode": genVisitorResp.GetRetcode(),
		}).Errorf("incarnateResp error")
		return nil, fmt.Errorf("genvisitor response error %v - %v",
			genVisitorResp.GetRetcode(), genVisitorResp.GetMsg())
	}

	incarnateParams := gout.H{
		"a":    "incarnate",
		"t":    url.QueryEscape(genVisitorResp.GetData().GetTid()),
		"gc":   "",
		"cb":   "cross_domain",
		"from": "weibo",
	}
	if genVisitorResp.GetData().GetNewTid() {
		incarnateParams["w"] = "3"
	} else {
		incarnateParams["w"] = "2"
	}
	incarnateParams["c"] = fmt.Sprintf("%03d", genVisitorResp.GetData().GetConfidence())
	incarnateResp, err := passportVisitor(incarnateParams, requests.WithCookieJar(jar))
	if err != nil {
		logger.Errorf("passportVisitor error %v", err)
		return nil, err
	}
	if incarnateResp.GetRetcode() != 20000000 || !strings.Contains(incarnateResp.GetMsg(), "succ") {
		logger.WithFields(logrus.Fields{
			"Msg":     incarnateResp.GetMsg(),
			"Retcode": incarnateResp.GetRetcode(),
		}).Errorf("incarnateResp error")
		return nil, fmt.Errorf("passportVisitor response error %v - %v",
			incarnateResp.GetRetcode(), incarnateResp.GetMsg())
	}

	crossdomainParams := gout.H{
		"a":    "crossdomain",
		"cb":   "return_back",
		"s":    incarnateResp.GetData().GetSub(),
		"sp":   incarnateResp.GetData().GetSubp(),
		"from": "weibo",
	}
	err = requests.Get(pathLoginVisitor, crossdomainParams, nil, requests.WithCookieJar(jar))
	if err != nil {
		logger.Errorf("loginVisitor error %v", err)
		return nil, err
	}

	cookieUrl, err := url.Parse(pathPassportVisitor)
	if err != nil {
		panic(fmt.Sprintf("path %v url parse error", pathPassportVisitor))
	}
	return jar.Cookies(cookieUrl), nil
}
