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

// OptFunc 预留的扩展字段，暂无用处
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

// RegisterConcern 向DDBOT注册一个 Concern。
func RegisterConcern(c Concern, opts ...OptFunc) {
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

// StartAll 启动所有 Concern，正常情况下框架会负责启动。
func StartAll() error {
	all := ListConcern()
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

// StopAll 停止所有Concern模块，正常情况下框架会负责停止。
// 会关闭notifyChan，所以停止后禁止再向notifyChan中写入数据。
func StopAll() {
	all := ListConcern()
	for _, c := range all {
		c.Stop()
	}
	close(notifyChan)
}

// ListConcern 返回所有注册过的 Concern。
func ListConcern() []Concern {
	return globalCenter.concernList
}

// GetConcernBySite 根据site返回 Concern。
// 如果site没有注册过，则会返回 ErrSiteNotSupported。
func GetConcernBySite(site string) (Concern, error) {
	if err := checkSite(site); err != nil {
		return nil, err
	}
	return globalCenter.concernMap[site], nil
}

// GetConcernBySiteAndType 根据site和ctype返回 Concern。
// 如果site没有注册过，则会返回 ErrSiteNotSupported；
// 如果site注册过，但不支持指定的ctype，则会返回 ErrTypeNotSupported。
func GetConcernBySiteAndType(site string, ctype concern_type.Type) (Concern, error) {
	if err := checkSiteAndType(site, ctype); err != nil {
		return nil, err
	}
	return globalCenter.concernMap[site], nil
}

// ListSite 返回所有注册的 Concern 支持的site。
func ListSite() []string {
	return globalCenter.concernSites
}

// GetNotifyChan 推送 channel，所有推送需要发送到这个channel中。
func GetNotifyChan() chan<- Notify {
	return notifyChan
}

// ReadNotifyChan 读取推送 channel，应该只由框架负责调用。
func ReadNotifyChan() <-chan Notify {
	return notifyChan
}

// GetConcernTypes 根据site查询 Concern，返回 Concern.Types。
// 如果site没有注册过，则返回 ErrSiteNotSupported。
func GetConcernTypes(site string) (concern_type.Type, error) {
	if err := checkSite(site); err != nil {
		return concern_type.Empty, err
	}
	return globalCenter.concernTypeMap[site], nil
}

// ParseRawSite 解析string格式的site，可以安全处理用户输入的site。
// 返回匹配的site，如果没有匹配上，返回 ErrSiteNotSupported。
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

// ParseRawSiteAndType 尝试解析string格式的 site 和 ctype，可以安全处理用户输入的site和ctype。
// 如果site合法，rawType为空，则默认返回注册时的第一个type。
// rawSite 和 rawType 默认为前缀匹配模式，即bi可以匹配bilibili，n可以匹配news。
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
	ctypes, _ := GetConcernTypes(site)
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

// GetConcernByParseSite 尝试解析string格式的 site，默认为前缀匹配模式。
func GetConcernByParseSite(rawSite string) (Concern, error) {
	site, err := ParseRawSite(rawSite)
	if err != nil {
		return nil, err
	}
	return GetConcernBySite(site)
}

// GetConcernByParseSiteAndType 尝试解析string格式的 site 和 ctype，可以安全处理用户输入的site和ctype，
// 并返回 Concern，site，ctype。
// 默认为前缀匹配模式
func GetConcernByParseSiteAndType(rawSite, rawType string) (Concern, string, concern_type.Type, error) {
	site, ctypes, err := ParseRawSiteAndType(rawSite, rawType)
	if err != nil {
		return nil, "", concern_type.Empty, err
	}
	cm, _ := GetConcernBySiteAndType(site, ctypes)
	return cm, site, ctypes, nil
}
