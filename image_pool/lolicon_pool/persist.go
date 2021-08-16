package lolicon_pool

import (
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/buntdb"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func (pool *LoliconPool) load() {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()

	logger.Debug("image pool load cache start")

	for k, v := range pool.cache {
		var img []*Setu
		err := localdb.RCoverTx(func(tx *buntdb.Tx) error {
			key := localdb.LoliconPoolStoreKey(k.String())
			val, err := tx.Get(key)
			if err == buntdb.ErrNotFound {
				return nil
			} else if err != nil {
				return err
			}
			return json.Unmarshal([]byte(val), &img)
		})
		if err != nil {
			logger.WithField("r18", k.String()).
				Errorf("image pool cache load failed %v", err)
			continue
		}
		for _, i := range img {
			v.PushBack(i)
		}
		logger.WithField("r18", k.String()).
			WithField("image_count", v.Len()).
			Debug("image cache loaded")
	}
}

func (pool *LoliconPool) store() {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()

	if !pool.changed {
		return
	}

	defer func() {
		pool.changed = false
	}()

	logger.Debug("image pool store cache start")

	for k, v := range pool.cache {
		var img []*Setu
		root := v.Front()
		if root == nil {
			continue
		}
		for {
			img = append(img, root.Value.(*Setu))
			if root == v.Back() {
				break
			}
			root = root.Next()
			if root == nil {
				break
			}
		}
		err := localdb.RWCoverTx(func(tx *buntdb.Tx) error {
			key := localdb.LoliconPoolStoreKey(k.String())
			b, err := json.Marshal(img)
			if err != nil {
				return err
			}
			_, _, err = tx.Set(key, string(b), nil)
			return err
		})
		if err != nil {
			logger.WithField("r18", k.String()).Errorf("image pool cache store failed %v", err)
			continue
		}
		logger.WithField("r18", k.String()).
			WithField("image_count", v.Len()).
			Debug("image pool cache stored")
	}
}
