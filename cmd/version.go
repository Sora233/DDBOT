package main

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

var (
	CommitId  = "UNKNOWN"
	BuildTime = "UNKNOWN"
	Tags      = "UNKNOWN"
)

func CheckUpdate() {
	defer func() {
		if e := recover(); e != nil {
			logrus.Errorf("更新检测失败：%v", e)
		}
	}()
	if Tags == "UNKNOWN" {
		logrus.Debug("自编译版本，跳过更新检测")
		return
	}
	var opts = []requests.Option{
		requests.TimeoutOption(time.Second * 3),
		requests.ProxyOption(proxy_pool.PreferOversea),
		requests.RetryOption(2),
	}
	var m map[string]interface{}
	err := requests.Get("https://api.github.com/repos/Sora233/DDBOT/releases/latest", nil, &m, opts...)
	if err != nil {
		logrus.Errorf("更新检测失败：%v", err)
		return
	}
	if msg := m["message"]; msg != nil {
		if s, ok := msg.(string); ok {
			logrus.Errorf("更新检测失败：%v", s)
			return
		}
	}
	latestTagName := m["tag_name"].(string)

	if compareVersion(Tags, latestTagName) {
		logrus.Infof("更新检测完成：DDBOT有可用更新版本【%v】，请前往 https://github.com/Sora233/DDBOT/releases 查看详细信息", latestTagName)
	} else {
		logrus.Debug("更新检测完成：当前为DDBOT最新版本")
	}

}

// compareVersion return true if a < b
func compareVersion(a, b string) bool {
	splitVersion := func(a string) []int {
		a = strings.TrimPrefix(a, "v")
		var result []int
		sp := strings.Split(a, ".")
		for _, i := range sp {
			x, err := strconv.ParseInt(i, 10, 0)
			if err != nil {
				return nil
			}
			result = append(result, int(x))
		}
		return result
	}
	sa, sb := splitVersion(a), splitVersion(b)
	if sa == nil || sb == nil {
		return false
	}
	for idx := range sa {
		if idx >= len(sb) {
			return false
		}
		if sa[idx] > sb[idx] {
			return false
		} else if sa[idx] < sb[idx] {
			return true
		}
	}
	if len(sa) == len(sb) {
		return false
	}
	return true
}
