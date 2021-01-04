package zhima

import (
	"encoding/json"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	zhima_proxy_pool "github.com/Sora233/zhima-proxy-pool"
	"github.com/tidwall/buntdb"
)

type BuntdbPersister struct {
}

func (b *BuntdbPersister) Save(proxies []*zhima_proxy_pool.Proxy) error {
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		key := localdb.Key("zhimaproxy", "active")
		bproxy, err := json.Marshal(proxies)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(key, string(bproxy), nil)
		return err
	})
	return err
}

func (b *BuntdbPersister) Load() ([]*zhima_proxy_pool.Proxy, error) {
	db, err := localdb.GetClient()
	if err != nil {
		return nil, err
	}
	var proxies = make([]*zhima_proxy_pool.Proxy, 0)
	err = db.Update(func(tx *buntdb.Tx) error {
		key := localdb.Key("zhimaproxy", "active")
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), &proxies)
		if err != nil {
			return err
		}
		return nil
	})
	return proxies, err
}

func NewBuntdbPersister() *BuntdbPersister {
	return &BuntdbPersister{}
}
