package buntdb

import (
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/modern-go/gls"
	"github.com/tidwall/buntdb"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ShortCut struct{}

var shortCut = new(ShortCut)

var TxKey = new(struct{})

var logger = utils.GetModuleLogger("localdb")

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

func (*ShortCut) RWCover(f func() error) error {
	if itx := gls.Get(TxKey); itx != nil {
		return f()
	}
	db, err := GetClient()
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		var err error
		gls.WithEmptyGls(func() {
			gls.Set(TxKey, tx)
			err = f()
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

func (s *ShortCut) JsonSave(key string, obj interface{}, overwrite bool, opts ...*buntdb.SetOptions) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		b, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		var opt *buntdb.SetOptions
		var replaced bool
		if len(opts) > 0 {
			opt = opts[0]
		}
		_, replaced, err = tx.Set(key, string(b), opt)
		if replaced && !overwrite {
			return ErrRollback
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
		if err != nil {
			logger.Errorf("JsonGet %v failed %v", reflect.TypeOf(obj).Name(), err)
		}
		return err
	})
}

func (s *ShortCut) SeqNext(key string) (int64, error) {
	var result int64
	err := s.RWCoverTx(func(tx *buntdb.Tx) error {
		oldV, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			oldV = "0"
		} else if err != nil {
			return err
		}
		old, err := strconv.ParseInt(oldV, 10, 64)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(key, strconv.FormatInt(old+1, 10), nil)
		if err == nil {
			result = old + 1
		}
		return err
	})
	return result, err
}

func (s *ShortCut) SeqClear(key string) error {
	err := s.RWCoverTx(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(key)
		if err == buntdb.ErrNotFound {
			err = nil
		}
		return err
	})
	return err
}

// SetIfNotExist 使用opt设置key value，如果key已经存在，则回滚并返回 ErrRollback
func (s *ShortCut) SetIfNotExist(key, value string, opt ...*buntdb.SetOptions) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		var (
			replaced bool
			err      error
		)
		if len(opt) == 0 {
			_, replaced, err = tx.Set(key, value, nil)
		} else {
			_, replaced, err = tx.Set(key, value, opt[0])
		}
		if err != nil {
			return err
		}
		if replaced {
			return ErrRollback
		}
		return nil
	})
}

func (s *ShortCut) CreatePatternIndex(patternFunc KeyPatternFunc, suffix []interface{}, less ...func(a, b string) bool) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		if len(less) == 0 {
			return tx.CreateIndex(patternFunc(suffix...), patternFunc(append(suffix[:], "*")...), buntdb.IndexString)
		}
		return tx.CreateIndex(patternFunc(suffix...), patternFunc(append(suffix[:], "*")...), less...)
	})
}

func (s *ShortCut) GetInt64(key string) (int64, error) {
	var result int64
	err := s.RCoverTx(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		r, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		result = r
		return nil
	})
	return result, err
}

func (s *ShortCut) SetInt64(key string, value int64, opt ...*buntdb.SetOptions) (int64, error) {
	var prev int64
	err := s.RWCoverTx(func(tx *buntdb.Tx) error {
		var err error
		var s string
		if len(opt) == 0 {
			s, _, err = tx.Set(key, strconv.FormatInt(value, 10), nil)
		} else {
			s, _, err = tx.Set(key, strconv.FormatInt(value, 10), opt[0])
		}
		if err != nil {
			return err
		}
		if len(s) == 0 {
			return nil
		}
		r, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		prev = r
		return nil
	})
	return prev, err
}

func RWCoverTx(f func(tx *buntdb.Tx) error) error {
	return shortCut.RWCoverTx(f)
}

func RCoverTx(f func(tx *buntdb.Tx) error) error {
	return shortCut.RCoverTx(f)
}

func RWCover(f func() error) error {
	return shortCut.RWCover(f)
}

func RCover(f func() error) error {
	return shortCut.RCover(f)
}

func JsonGet(key string, obj interface{}) error {
	return shortCut.JsonGet(key, obj)
}

func JsonSave(key string, obj interface{}, overwrite bool, opt ...*buntdb.SetOptions) error {
	return shortCut.JsonSave(key, obj, overwrite, opt...)
}

func SeqNext(key string) (int64, error) {
	return shortCut.SeqNext(key)
}

func SeqClear(key string) error {
	return shortCut.SeqClear(key)
}

func SetIfNotExist(key, value string, opt ...*buntdb.SetOptions) error {
	return shortCut.SetIfNotExist(key, value, opt...)
}

func SetInt64(key string, value int64, opt ...*buntdb.SetOptions) (int64, error) {
	return shortCut.SetInt64(key, value, opt...)
}

func GetInt64(key string) (int64, error) {
	return shortCut.GetInt64(key)
}

// ExpireOption 是一个创建expire的函数糖
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
func RemoveByPrefixAndIndex(prefixKey []string, indexKey []string) ([]string, error) {
	var deletedKey []string
	err := RWCoverTx(func(tx *buntdb.Tx) error {
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
			_, err := tx.Delete(key)
			if err == nil {
				deletedKey = append(deletedKey, key)
			}
		}
		return nil
	})
	return deletedKey, err
}

func CreatePatternIndex(patternFunc KeyPatternFunc, suffix []interface{}, less ...func(a, b string) bool) error {
	return shortCut.CreatePatternIndex(patternFunc, suffix, less...)
}
