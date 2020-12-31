package zhima

import (
	"container/list"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/asmcos/requests"
	"math/rand"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("zhima_proxy_pool")

const (
	BackUpCap = 20
	TimeLimit = time.Minute * 10
)

type proxy struct {
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	ExpireTime string `json:"expire_time"`
}

func (p *proxy) ProxyString() string {
	return fmt.Sprintf("%v:%v", p.Ip, p.Port)
}

func (p *proxy) Expired() bool {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", p.ExpireTime, time.Local)
	if err != nil {
		return true
	}
	return t.Before(time.Now())
}

type Response struct {
	Code    int      `json:"code"`
	Data    []*proxy `json:"data"`
	Msg     string   `json:"msg"`
	Success bool     `json:"success"`
}

type zhimaPool struct {
	api         string
	backupProxy *list.List
	activeProxy []*proxy
	*sync.Cond
	activeMutex *sync.RWMutex
}

func (pool *zhimaPool) Start() {
	go pool.FillBackup()
	go func() {
		ticker := time.NewTicker(time.Second * 90)
		for {
			select {
			case <-ticker.C:
				go pool.Clear()
			}
		}
	}()
}

func (pool *zhimaPool) Clear() {
	// zhima proxy timeout
	pool.L.Lock()
	defer pool.L.Unlock()
	logger.Debug("zhima pool clear")
	pool.backupProxy = list.New()
}

func (pool *zhimaPool) FillBackup() {
	for {
		pool.L.Lock()

		for pool.checkBackup() {
			pool.Wait()
		}
		logger.WithField("backup size", pool.backupProxy.Len()).Debug("backup proxy not enough")

		resp, err := requests.Get(pool.api)
		if err != nil {
			logger.Errorf("fresh failed %v", err)
			pool.L.Unlock()
			continue
		}
		zhimaResp := new(Response)
		err = resp.Json(zhimaResp)
		if err != nil {
			logger.Errorf("parse zhima response failed %v", err)
		}
		if zhimaResp.Code != 0 {
			logger.WithField("code", zhimaResp.Code).
				WithField("msg", zhimaResp.Msg).
				Errorf("fresh failed")
			pool.backupProxy = nil
		} else {
			now := time.Now()
			for _, proxy := range zhimaResp.Data {
				t, err := time.ParseInLocation("2006-01-02 15:04:05", proxy.ExpireTime, time.Local)
				if err != nil {
					continue
				}
				if t.Sub(now) > TimeLimit+time.Second*20 {
					pool.backupProxy.PushBack(proxy)
				}
			}
			if pool.backupProxy.Len() >= BackUpCap {
				pool.Signal()
			}
		}
		logger.WithField("backup size", pool.backupProxy.Len()).Debug("backup freshed")
		pool.L.Unlock()
	}
}

func (pool *zhimaPool) Get() (proxy_pool.IProxy, error) {
	var result *proxy
	pool.L.Lock()
	for !pool.checkBackup() {
		pool.Signal()
		pool.Wait()
	}

	for len(pool.activeProxy) < 6 && pool.backupProxy.Len() != 0 {
		backProxy := pool.backupProxy.Front()
		pool.activeProxy = append(pool.activeProxy, backProxy.Value.(*proxy))
		pool.backupProxy.Remove(backProxy)
	}

	pos := rand.Intn(len(pool.activeProxy))
	result = pool.activeProxy[pos]
	//logger.WithField("return proxy", result).WithField("all", pool.activeProxy).Debug("proxy")
	pool.L.Unlock()
	return result, nil
}

func (pool *zhimaPool) Delete(iProxy proxy_pool.IProxy) bool {
	pool.L.Lock()
	defer pool.L.Unlock()
	for index, curProxy := range pool.activeProxy {
		if curProxy.ProxyString() == iProxy.ProxyString() {
			for !pool.checkBackup() {
				pool.Signal()
				pool.Wait()
			}
			backup := pool.backupProxy.Front()
			pool.activeProxy[index] = backup.Value.(*proxy)
			//logger.WithField("addr", iProxy).
			//	WithField("proxy", iProxy.ProxyString()).
			//	WithField("new_proxy", pool.activeProxy[index].ProxyString()).
			//	WithField("result", result).
			//	Debug("delete")
		}
	}
	return true
}

func (pool *zhimaPool) checkBackup() bool {
	return pool.backupProxy.Len() != 0
}

func NewZhimaPool(api string) *zhimaPool {
	activeMutex := new(sync.RWMutex)
	pool := &zhimaPool{
		api:         api,
		activeProxy: make([]*proxy, 0),
		backupProxy: list.New(),
		Cond:        sync.NewCond(activeMutex),
		activeMutex: activeMutex,
	}
	pool.Start()
	return pool
}
