package buntdb

import (
	"encoding/json"
	"github.com/modern-go/gls"
	"github.com/tidwall/buntdb"
	"strings"
	"time"
)

type ShortCut struct{}

var shortCut = new(ShortCut)

var TxKey = new(struct{})

// RWCoverTx 在一个RW事务中执行f，注意f的返回值不一定是RWCoverTx的返回值
// 有可能f返回nil，但RWTxCover返回non-nil
// 可以忽略error，但不要简单地用f返回值替代RWTxCover返回值，ref: bilibili/MarkDynamicId
func (*ShortCut) RWCoverTx(f func(tx *buntdb.Tx) error) error {
	if itx := gls.Get(TxKey); itx != nil {
		return f(itx.(*buntdb.Tx))
	}
	db, err := GetClient()
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		var err error
		gls.WithEmptyGls(func() {
			gls.Set(TxKey, tx)
			err = f(tx)
		})()
		return err
	})
}

func (*ShortCut) RCoverTx(f func(tx *buntdb.Tx) error) error {
	if itx := gls.Get(TxKey); itx != nil {
		return f(itx.(*buntdb.Tx))
	}
	db, err := GetClient()
	if err != nil {
		return err
	}
	return db.View(func(tx *buntdb.Tx) error {
		var err error
		gls.WithEmptyGls(func() {
			gls.Set(TxKey, tx)
			err = f(tx)
		})()
		return err
	})
}

func (*ShortCut) RCover(f func() error) error {
	if itx := gls.Get(TxKey); itx != nil {
		return f()
	}
	db, err := GetClient()
	if err != nil {
		return err
	}
	return db.View(func(tx *buntdb.Tx) error {
		var err error
		gls.WithEmptyGls(func() {
			gls.Set(TxKey, tx)
			err = f()
		})()
		return err
	})
}

func (s *ShortCut) JsonSave(key string, obj interface{}, opt ...*buntdb.SetOptions) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		b, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		if len(opt) == 0 {
			_, _, err = tx.Set(key, string(b), nil)
		} else {
			_, _, err = tx.Set(key, string(b), opt[0])
		}
		return err
	})
}

func (s *ShortCut) JsonGet(key string, obj interface{}) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), obj)
		return err
	})
}

func RWCoverTx(f func(tx *buntdb.Tx) error) error {
	return shortCut.RWCoverTx(f)
}

func RCoverTx(f func(tx *buntdb.Tx) error) error {
	return shortCut.RCoverTx(f)
}

func RCover(f func() error) error {
	return shortCut.RCover(f)
}

func JsonGet(key string, obj interface{}) error {
	return shortCut.JsonGet(key, obj)
}

func JsonSave(key string, obj interface{}, opt ...*buntdb.SetOptions) error {
	return shortCut.JsonSave(key, obj, opt...)
}

func ExpireOption(duration time.Duration) *buntdb.SetOptions {
	if duration == 0 {
		return nil
	}
	return &buntdb.SetOptions{
		Expires: true,
		TTL:     duration,
	}
}

// RemoveByPrefixAndIndex 遍历每个index，如果一个key满足任意prefix，则删掉
func RemoveByPrefixAndIndex(prefixKey []string, indexKey []string) error {
	return RWCoverTx(func(tx *buntdb.Tx) error {
		var removeKey = make(map[string]interface{})
		var iterErr error
		for _, index := range indexKey {
			iterErr = tx.Ascend(index, func(key, value string) bool {
				for _, prefix := range prefixKey {
					if strings.HasPrefix(key, prefix) {
						removeKey[key] = struct{}{}
						return true
					}
				}
				return true
			})
			if iterErr != nil {
				return iterErr
			}
		}
		for key := range removeKey {
			tx.Delete(key)
		}
		return nil
	})
}
