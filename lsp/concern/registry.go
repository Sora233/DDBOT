package concern

import (
	"fmt"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/utils"
	"golang.org/x/sync/errgroup"
	"strings"
)

var globalCenter = newConcernCenter()
var notifyChan = make(chan Notify, 500)

type option struct {
}

type OptFunc func(opt *option) *option

type center struct {
	concernMap     map[string]Concern
	concernTypeMap map[string]concern_type.Type
	concernSites   []string
	concernList    []Concern
}

func checkSite(site string) error {
	if _, found := globalCenter.concernMap[site]; found {
		return nil
	}
	return ErrSiteNotSupported
}

func checkSiteAndType(site string, ctype concern_type.Type) error {
	var err error
	if err = checkSite(site); err != nil {
		return err
	}
	if combineType, found := globalCenter.concernTypeMap[site]; found {
		if combineType.ContainAll(ctype) {
			return nil
		}
		return ErrTypeNotSupported
	}
	return ErrSiteNotSupported
}

func newConcernCenter() *center {
	return &center{
		concernMap:     make(map[string]Concern),
		concernTypeMap: make(map[string]concern_type.Type),
		concernList:    nil,
		concernSites:   nil,
	}
}

func RegisterConcernManager(c Concern, opts ...OptFunc) {
	if c == nil {
		panic("Concern: Register <nil> concern")
	}
	site := c.Site()
	if _, found := globalCenter.concernMap[site]; found {
		panic(fmt.Sprintf("Concern %v: is already registered", site))
	}
	for _, ctype := range c.Types() {
		if !ctype.IsTrivial() {
			panic(fmt.Sprintf("Concern %v: Type %v IsTrivial() must be True", site, ctype.String()))
		}
	}
	if concern_type.Empty.Add(c.Types()...).Empty() {
		panic(fmt.Sprintf("Concern %v: register with empty types", site))
	}
	globalCenter.concernMap[site] = c
	globalCenter.concernList = append(globalCenter.concernList, c)
	globalCenter.concernSites = append(globalCenter.concernSites, c.Site())
	globalCenter.concernTypeMap[site] = concern_type.Empty.Add(c.Types()...)
}

// ClearConcern 现阶段仅用于测试，如果用于其他目的将导致不可预料的后果。
func ClearConcern() {
	globalCenter = newConcernCenter()
}

func StartAll() error {
	all := ListConcernManager()
	errG := errgroup.Group{}
	for _, c := range all {
		c := c
		errG.Go(func() error {
			c.FreshIndex()
			logger.Debugf("启动Concern %v模块", c.Site())
			err := c.Start()
			if err != nil {
				logger.Errorf("启动Concern %v 失败 - %v", c.Site(), err)
			}
			return err
		})
	}
	return errG.Wait()
}

// StopAll 停止所有Concern模块，会关闭notifyChan，所以停止后禁止再向notifyChan中写入数据
func StopAll() {
	all := ListConcernManager()
	for _, c := range all {
		c.Stop()
	}
	close(notifyChan)
}

func ListConcernManager() []Concern {
	return globalCenter.concernList
}

func GetConcernManagerBySite(site string) (Concern, error) {
	if err := checkSite(site); err != nil {
		return nil, err
	}
	return globalCenter.concernMap[site], nil
}

func GetConcernManagerBySiteAndType(site string, ctype concern_type.Type) (Concern, error) {
	if err := checkSiteAndType(site, ctype); err != nil {
		return nil, err
	}
	return globalCenter.concernMap[site], nil
}

func ListSite() []string {
	return globalCenter.concernSites
}

func GetNotifyChan() chan<- Notify {
	return notifyChan
}

func ReadNotifyChan() <-chan Notify {
	return notifyChan
}

func GetConcernManagerTypes(site string) (concern_type.Type, error) {
	if err := checkSite(site); err != nil {
		return concern_type.Empty, err
	}
	return globalCenter.concernTypeMap[site], nil
}

func ParseRawSite(rawSite string) (string, error) {
	var (
		found bool
		site  string
	)

	rawSite = strings.Trim(rawSite, `"`)
	site, found = utils.PrefixMatch(ListSite(), rawSite)
	if !found {
		return "", ErrSiteNotSupported
	}
	return site, nil
}

// ParseRawSiteAndType 尝试解析string格式的 site 和 concern_type.Type，可以安全处理用户输入的site和type
// 如果site合法，rawType为空，则默认返回注册时的第一个type
func ParseRawSiteAndType(rawSite string, rawType string) (string, concern_type.Type, error) {
	var (
		site  string
		_type string
		found bool
		err   error
	)
	rawSite = strings.Trim(rawSite, `"`)
	rawType = strings.Trim(rawType, `"`)
	site, err = ParseRawSite(rawSite)
	if err != nil {
		return "", concern_type.Empty, err
	}
	var sTypes []string
	ctypes, _ := GetConcernManagerTypes(site)
	// 如果没有指定type，则默认注册时的第一个type
	if rawType == "" {
		return site, ctypes.Split()[0], nil
	}
	for _, t := range ctypes.Split() {
		sTypes = append(sTypes, t.String())
	}
	_type, found = utils.PrefixMatch(sTypes, rawType)
	if !found {
		return "", concern_type.Empty, ErrTypeNotSupported
	}
	return site, concern_type.Type(_type), nil
}

func GetConcernManagerByParseSite(rawSite string) (Concern, error) {
	site, err := ParseRawSite(rawSite)
	if err != nil {
		return nil, err
	}
	return GetConcernManagerBySite(site)
}

func GetConcernManagerByParseSiteAndType(rawSite, rawType string) (Concern, string, concern_type.Type, error) {
	site, ctypes, err := ParseRawSiteAndType(rawSite, rawType)
	if err != nil {
		return nil, "", concern_type.Empty, err
	}
	cm, _ := GetConcernManagerBySiteAndType(site, ctypes)
	return cm, site, ctypes, nil
}
