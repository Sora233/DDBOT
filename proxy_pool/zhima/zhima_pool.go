package zhima

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/asmcos/requests"
	"github.com/tidwall/buntdb"
	"sync"
	"time"
)

var (
	logger = utils.GetModuleLogger("zhima_proxy_pool")
)

const (
	BackUpCap = 30
	ActiveCap = 6
	TimeLimit = time.Minute * 170
)

type proxy struct {
	Ip               string `json:"ip"`
	Port             int    `json:"port"`
	ExpireTimeString string `json:"expire_time"`
}

func (p *proxy) ProxyString() string {
	return fmt.Sprintf("%v:%v", p.Ip, p.Port)
}

// ExpireTime return a time.Time parsed from ExpireTimeString
func (p *proxy) ExpireTime() (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", p.ExpireTimeString, time.Local)
}

func (p *proxy) Expired() bool {
	t, err := p.ExpireTime()
	if err != nil {
		return true
	}
	return t.Before(time.Now().Add(time.Second * 10))
}

type Response struct {
	Code    int      `json:"code"`
	Data    []*proxy `json:"data"`
	Msg     string   `json:"msg"`
	Success bool     `json:"success"`
}

// http://h.zhimaruanjian.com/
type zhimaPool struct {
	api         string
	backupProxy *list.List
	activeProxy []*proxy
	*sync.Cond
	activeMutex *sync.RWMutex

	index int
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
	pool.activeMutex.Lock()
	defer pool.activeMutex.Unlock()
	for len(pool.activeProxy) < ActiveCap {
		backup, err := pool.popBackup()
		if err != nil {
			logger.Errorf("fill active proxy failed %v", err)
		} else {
			pool.activeProxy = append(pool.activeProxy, backup)
		}
	}
}

func (pool *zhimaPool) Clear() {
	// zhima proxy timeout
	pool.L.Lock()
	defer pool.L.Unlock()
	logger.Debug("backup cleared")
	pool.backupProxy = list.New()
}

func (pool *zhimaPool) FillBackup() {
	for {
		pool.L.Lock()

		for pool.checkBackup() {
			pool.Broadcast()
			pool.Wait()
		}
		logger.WithField("backup size", pool.backupProxy.Len()).Debug("backup proxy not enough... fresh")

		var loopCount = 0

		for pool.backupProxy.Len() < BackUpCap {
			if loopCount >= 5 {
				logger.WithField("backup size", pool.backupProxy.Len()).
					Errorf("can not get enough backup proxy after fetch 5 times, check your timeLimit or backupCap")
				break
			}
			loopCount += 1
			resp, err := requests.Get(pool.api)
			if err != nil {
				logger.Errorf("fresh failed %v", err)
				pool.L.Unlock()
				break
			}
			zhimaResp := new(Response)
			err = resp.Json(zhimaResp)
			if err != nil {
				logger.Errorf("parse zhima response failed %v", err)
				pool.L.Unlock()
				break
			}
			if zhimaResp.Code != 0 {
				log := logger.WithField("code", zhimaResp.Code).
					WithField("msg", zhimaResp.Msg)
				switch zhimaResp.Code {
				case 111:
					time.Sleep(time.Second * 5)
				default:
					log.Errorf("fresh failed")
				}
			} else {
				now := time.Now()
				for _, proxy := range zhimaResp.Data {
					t, err := proxy.ExpireTime()
					if err != nil {
						continue
					}
					if t.Sub(now) >= TimeLimit {
						pool.backupProxy.PushBack(proxy)
					}
				}
			}
		}
		if pool.checkBackup() {
			pool.Broadcast()
		}
		logger.WithField("backup size", pool.backupProxy.Len()).Debug("backup freshed")
		pool.L.Unlock()
	}
}

func (pool *zhimaPool) Get() (proxy_pool.IProxy, error) {
	var result *proxy
	pool.activeMutex.RLock()

	pos := pool.index
	pool.index = (pool.index + 1) % ActiveCap

	result = pool.activeProxy[pos]
	if result.Expired() {
		pool.activeMutex.RUnlock()
		pool.activeMutex.Lock()
		if result.ProxyString() == pool.activeProxy[pos].ProxyString() {
			err := pool.replaceActive(pos)
			if err != nil {
				pool.activeMutex.Unlock()
				return nil, err
			}
		}
		pool.activeMutex.Unlock()
		pool.activeMutex.RLock()
	}
	result = pool.activeProxy[pos]
	pool.activeMutex.RUnlock()

	//logger.WithField("return proxy", result).WithField("all", pool.activeProxy).Debug("proxy")
	return result, nil
}

func (pool *zhimaPool) Delete(iProxy proxy_pool.IProxy) bool {
	pool.L.Lock()
	defer pool.L.Unlock()

	var result = false

	for index, curProxy := range pool.activeProxy {
		if curProxy.ProxyString() == iProxy.ProxyString() {
			err := pool.replaceActive(index)
			if err == nil {
				result = true
			}
			break
		}
	}
	return result
}

func (pool *zhimaPool) Stop() error {
	pool.L.Lock()
	defer pool.L.Unlock()

	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		key := localdb.Key("zhimaproxy", "active")
		bproxy, err := json.Marshal(pool.activeProxy)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(key, string(bproxy), nil)
		return err
	})
	return err
}

func (pool *zhimaPool) replaceActive(index int) (err error) {
	log := logger.WithField("deleted_proxy", pool.activeProxy[index].ProxyString()).WithField("old_expire", pool.activeProxy[index].ExpireTimeString)
	oldProxy := pool.activeProxy[index]
	newProxy, err := pool.popBackup()
	if err != nil {
		return err
	}
	if oldProxy.ProxyString() == pool.activeProxy[index].ProxyString() {
		pool.activeProxy[index] = newProxy
		log.WithField("new_proxy", pool.activeProxy[index].ProxyString()).WithField("new_expire", pool.activeProxy[index].ExpireTimeString).Debug("deleted")
	} else {
		log.Errorf("delete proxy changed")
	}
	return nil
}

func (pool *zhimaPool) loadActive() error {
	pool.L.Lock()
	defer pool.L.Unlock()

	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		key := localdb.Key("zhimaproxy", "active")
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		var proxies []*proxy = make([]*proxy, 0)
		err = json.Unmarshal([]byte(val), &proxies)
		if err != nil {
			return err
		}
		for _, proxy := range proxies {
			if !proxy.Expired() {
				pool.activeProxy = append(pool.activeProxy, proxy)
			}
		}
		return nil
	})
	return err
}

func (pool *zhimaPool) popBackup() (*proxy, error) {
	// caller must hold the lock
	for !pool.checkBackup() {
		pool.Signal()
		pool.Wait()
	}
	backup := pool.backupProxy.Front()
	pool.backupProxy.Remove(backup)
	return backup.Value.(*proxy), nil
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
	if err := pool.loadActive(); err != nil {
		logger.WithField("active size", len(pool.activeProxy)).Debug("load err %v", err)
	} else {
		logger.WithField("active size", len(pool.activeProxy)).Debug("load ok")
	}
	pool.Start()
	return pool
}
