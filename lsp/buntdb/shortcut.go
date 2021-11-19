package buntdb

import (
	"errors"
	"github.com/Sora233/MiraiGo-Template/utils"
	"github.com/modern-go/gls"
	"github.com/tidwall/buntdb"
	"strconv"
	"strings"
	"time"
)

// ShortCut 包含了许多数据库的读写helper，只需嵌入即可使用，如果不想嵌入，也可以通过包名调用
type ShortCut struct{}

var shortCut = new(ShortCut)

var txKey = new(struct{})

var logger = utils.GetModuleLogger("localdb")

// RWCoverTx 在一个可读可写事务中执行f，注意f的返回值不一定是RWCoverTx的返回值
// 有可能f返回nil，但RWTxCover返回non-nil
// 可以忽略error，但不要简单地用f返回值替代RWTxCover返回值，ref: bilibili/MarkDynamicId
// 需要注意可写事务是唯一的，同一时间只会存在一个可写事务，所有耗时操作禁止放在可写事务中执行
// 在同一Goroutine中，可写事务可以嵌套
func (*ShortCut) RWCoverTx(f func(tx *buntdb.Tx) error) error {
	if itx := gls.Get(txKey); itx != nil {
		return f(itx.(*buntdb.Tx))
	}
	db, err := GetClient()
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		var err error
		gls.WithEmptyGls(func() {
			gls.Set(txKey, tx)
			err = f(tx)
		})()
		return err
	})
}

// RWCover 在一个可读可写事务中执行f，不同的是它不获取 buntdb.Tx ，而由 f 自己控制。
// 需要注意可写事务是唯一的，同一时间只会存在一个可写事务，所有耗时操作禁止放在可写事务中执行
// 在同一Goroutine中，可写事务可以嵌套
func (*ShortCut) RWCover(f func() error) error {
	if itx := gls.Get(txKey); itx != nil {
		return f()
	}
	db, err := GetClient()
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		var err error
		gls.WithEmptyGls(func() {
			gls.Set(txKey, tx)
			err = f()
		})()
		return err
	})
}

// RCoverTx 在一个只读事务中执行f。
// 所有写操作会失败或者回滚。
func (*ShortCut) RCoverTx(f func(tx *buntdb.Tx) error) error {
	if itx := gls.Get(txKey); itx != nil {
		return f(itx.(*buntdb.Tx))
	}
	db, err := GetClient()
	if err != nil {
		return err
	}
	return db.View(func(tx *buntdb.Tx) error {
		var err error
		gls.WithEmptyGls(func() {
			gls.Set(txKey, tx)
			err = f(tx)
		})()
		return err
	})
}

// RCover 在一个只读事务中执行f，不同的是它不获取 buntdb.Tx ，而由 f 自己控制。
// 所有写操作会失败，或者回滚。
func (*ShortCut) RCover(f func() error) error {
	if itx := gls.Get(txKey); itx != nil {
		return f()
	}
	db, err := GetClient()
	if err != nil {
		return err
	}
	return db.View(func(tx *buntdb.Tx) error {
		var err error
		gls.WithEmptyGls(func() {
			gls.Set(txKey, tx)
			err = f()
		})()
		return err
	})
}

// SeqNext 将key上的int64值加上1并保存，返回保存后的值。
// 如果key不存在，则会默认其为0，返回值为1
// 等价于 s.IncInt64(key, 1)
func (s *ShortCut) SeqNext(key string) (int64, error) {
	return s.IncInt64(key, 1)
}

// IncInt64 将key上的int64值加上 value 并保存，返回保存后的值。
// 如果key不存在，则会默认其为0，返回值为1
// 如果key上的value不是一个int64，则会返回错误
func (s *ShortCut) IncInt64(key string, value int64) (int64, error) {
	var result int64
	err := s.RWCover(func() error {
		oldVal, err := s.GetInt64(key, IgnoreNotFoundOpt())
		if err != nil {
			return err
		}
		result = oldVal + value
		return s.SetInt64(key, result)
	})
	if err != nil {
		result = 0
	}
	return result, err
}

// GetJson 获取key对应的value，并通过 json.Unmarshal 到obj上
// 支持 GetIgnoreExpireOpt IgnoreNotFoundOpt GetTTLOpt
func (s *ShortCut) GetJson(key string, obj interface{}, opt ...OptionFunc) error {
	if obj == nil {
		return errors.New("<nil obj>")
	}
	opts := getOption(opt...)
	var value string
	err := s.RCoverTx(func(tx *buntdb.Tx) error {
		var err error
		value, err = s.getWithOpts(tx, key, opts)
		return err
	})
	if err != nil {
		return err
	}
	if len(value) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(value), obj)
}

// SetJson 将obj通过 json.Marshal 转成json字符串，并设置到key上。
// 支持 SetExpireOpt SetKeepLastExpireOpt SetNoOverWriteOpt SetGetIsOverwriteOpt
// SetGetPreviousValueStringOpt SetGetPreviousValueInt64Opt SetGetPreviousValueJsonObjectOpt
func (s *ShortCut) SetJson(key string, obj interface{}, opt ...OptionFunc) error {
	if obj == nil {
		return errors.New("<nil obj>")
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	opts := getOption(opt...)
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		return s.setWithOpts(tx, key, string(b), opts)
	})
}

// DeleteInt64 删除key，解析key上的值到int64并返回
// 支持 IgnoreNotFoundOpt
func (s *ShortCut) DeleteInt64(key string, opt ...OptionFunc) (int64, error) {
	return s.int64Wrapper(s.Delete(key, opt...))
}

// GetInt64 通过key获取value，并将value解析成int64
// 支持 GetIgnoreExpireOpt IgnoreNotFoundOpt GetTTLOpt
// 当设置了 IgnoreNotFoundOpt 时，key不存在时会直接返回0，不会返回错误
func (s *ShortCut) GetInt64(key string, opt ...OptionFunc) (int64, error) {
	return s.int64Wrapper(s.Get(key, opt...))
}

// SetInt64 通过key设置int64格式的value
// 支持 SetExpireOpt SetKeepLastExpireOpt SetNoOverWriteOpt SetGetIsOverwriteOpt
// SetGetPreviousValueStringOpt SetGetPreviousValueInt64Opt SetGetPreviousValueJsonObjectOpt
func (s *ShortCut) SetInt64(key string, value int64, opt ...OptionFunc) error {
	return s.Set(key, strconv.FormatInt(value, 10), opt...)
}

// Delete 删除key，并返回key上的值
// 支持 IgnoreNotFoundOpt
func (s *ShortCut) Delete(key string, opt ...OptionFunc) (string, error) {
	opts := getOption(opt...)
	var previous string
	err := s.RWCoverTx(func(tx *buntdb.Tx) error {
		var err error
		previous, err = s.deleteWithOpts(tx, key, opts)
		return err
	})
	return previous, err
}

// Get 通过key获取value
// 支持 GetIgnoreExpireOpt IgnoreNotFoundOpt GetTTLOpt
func (s *ShortCut) Get(key string, opt ...OptionFunc) (string, error) {
	var result string
	opts := getOption(opt...)
	err := s.RCoverTx(func(tx *buntdb.Tx) error {
		var err error
		result, err = s.getWithOpts(tx, key, opts)
		return err
	})
	return result, err
}

// Set 通过key设置value
// 支持 SetExpireOpt SetKeepLastExpireOpt SetNoOverWriteOpt SetGetIsOverwriteOpt
// SetGetPreviousValueStringOpt SetGetPreviousValueInt64Opt SetGetPreviousValueJsonObjectOpt
func (s *ShortCut) Set(key, value string, opt ...OptionFunc) error {
	opts := getOption(opt...)
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		return s.setWithOpts(tx, key, value, opts)
	})
}

// Exist 查询key是否存在，key不存在或者发生任何错误时返回 false
// 支持 GetTTLOpt GetIgnoreExpireOpt
func (s *ShortCut) Exist(key string, opt ...OptionFunc) bool {
	var result bool
	opts := getOption(opt...)
	err := s.RWCoverTx(func(tx *buntdb.Tx) error {
		result = s.existWithOpts(tx, key, opts)
		return nil
	})
	if err != nil {
		if !IsNotFound(err) {
			logger.Errorf("Exist key %v error %v", key, err)
		}
		result = false
	}
	return result
}

// setWithOpts 统一在有option的情况下的set行为，考虑到性能需要手动传 buntdb.Tx
func (s *ShortCut) setWithOpts(tx *buntdb.Tx, key string, value string, opt *option) error {
	var (
		prev     string
		replaced bool
		err      error
		setOpt   *buntdb.SetOptions
	)
	if innerOpt := opt.getInnerExpire(); innerOpt != nil {
		setOpt = innerOpt
	} else if opt.keepLastExpire {
		lastTTL, _ := tx.TTL(key)
		if lastTTL > 0 {
			setOpt = ExpireOption(lastTTL)
		}
	}
	prev, replaced, err = tx.Set(key, value, setOpt)
	if err != nil {
		return err
	}
	opt.setIsOverWrite(replaced)
	opt.setPrevious(prev)
	if replaced && opt.getNoOverWrite() {
		return ErrRollback
	}
	return nil
}

// getWithOpts 统一在有option的情况下的get行为，考虑到性能需要手动传 buntdb.Tx
func (s *ShortCut) getWithOpts(tx *buntdb.Tx, key string, opt *option) (string, error) {
	result, err := tx.Get(key, opt.getIgnoreExpire())
	if opt.getTTL() != nil {
		ttl, _ := tx.TTL(key)
		opt.setTTL(ttl)
	}
	if opt.getIgnoreNotFound() && IsNotFound(err) {
		err = nil
	}
	return result, err
}

// deleteWithOpts 统一在有option的情况下的delete行为，考虑到性能需要手动传 buntdb.Tx
func (s *ShortCut) deleteWithOpts(tx *buntdb.Tx, key string, opt *option) (string, error) {
	result, err := tx.Delete(key)
	if opt.getIgnoreNotFound() && IsNotFound(err) {
		err = nil
	}
	return result, err
}

// existWithOpts 统一在有option的情况下的exist行为，考虑到性能需要手动传 buntdb.Tx
func (s *ShortCut) existWithOpts(tx *buntdb.Tx, key string, opt *option) bool {
	_, err := tx.Get(key, opt.getIgnoreExpire())
	if opt.getTTL() != nil {
		ttl, _ := tx.TTL(key)
		opt.setTTL(ttl)
	}
	return err == nil
}

func (s *ShortCut) int64Wrapper(result string, err error) (int64, error) {
	if err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, nil
	}
	return strconv.ParseInt(result, 10, 64)
}

func (s *ShortCut) CreatePatternIndex(patternFunc KeyPatternFunc, suffix []interface{}, less ...func(a, b string) bool) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		var err error
		if len(less) == 0 {
			err = tx.CreateIndex(patternFunc(suffix...), patternFunc(append(suffix[:], "*")...), buntdb.IndexString)
		}
		err = tx.CreateIndex(patternFunc(suffix...), patternFunc(append(suffix[:], "*")...), less...)
		if err == buntdb.ErrIndexExists {
			err = nil
		}
		return err
	})
}

// RWCoverTx 在一个可读可写事务中执行f，注意f的返回值不一定是RWCoverTx的返回值
// 有可能f返回nil，但RWTxCover返回non-nil
// 可以忽略error，但不要简单地用f返回值替代RWTxCover返回值，ref: bilibili/MarkDynamicId
// 需要注意可写事务是唯一的，同一时间只会存在一个可写事务，所有耗时操作禁止放在可写事务中执行
// 在同一Goroutine中，可写事务可以嵌套
func RWCoverTx(f func(tx *buntdb.Tx) error) error {
	return shortCut.RWCoverTx(f)
}

// RWCover 在一个可读可写事务中执行f，不同的是它不获取 buntdb.Tx ，而由 f 自己控制。
// 需要注意可写事务是唯一的，同一时间只会存在一个可写事务，所有耗时操作禁止放在可写事务中执行
// 在同一Goroutine中，可写事务可以嵌套
func RWCover(f func() error) error {
	return shortCut.RWCover(f)
}

// RCoverTx 在一个只读事务中执行f。
// 所有写操作会失败或者回滚。
func RCoverTx(f func(tx *buntdb.Tx) error) error {
	return shortCut.RCoverTx(f)
}

// RCover 在一个只读事务中执行f，不同的是它不获取 buntdb.Tx ，而由 f 自己控制。
// 所有写操作会失败，或者回滚。
func RCover(f func() error) error {
	return shortCut.RCover(f)
}

// SeqNext 将key上的int64值加上1并保存，返回保存后的值。
// 如果key不存在，则会默认其为0，返回值为1
// 等价于 IncInt64(key, 1)
func SeqNext(key string) (int64, error) {
	return shortCut.SeqNext(key)
}

// IncInt64 将key上的int64值加上 value 并保存，返回保存后的值。
// 如果key不存在，则会默认其为0，返回值为1
// 如果key上的value不是一个int64，则会返回错误
func IncInt64(key string, value int64) (int64, error) {
	return shortCut.IncInt64(key, value)
}

// GetJson 获取key对应的value，并通过 json.Unmarshal 到obj上
// 支持 GetIgnoreExpireOpt IgnoreNotFoundOpt GetTTLOpt
func GetJson(key string, obj interface{}, opt ...OptionFunc) error {
	return shortCut.GetJson(key, obj, opt...)
}

// SetJson 将obj通过 json.Marshal 转成json字符串，并设置到key上。
// 支持 SetExpireOpt SetKeepLastExpireOpt SetNoOverWriteOpt SetGetIsOverwriteOpt
// SetGetPreviousValueStringOpt SetGetPreviousValueInt64Opt SetGetPreviousValueJsonObjectOpt
func SetJson(key string, obj interface{}, opt ...OptionFunc) error {
	return shortCut.SetJson(key, obj, opt...)
}

// DeleteInt64 删除key，解析key上的值到int64并返回
// 支持 IgnoreNotFoundOpt
func DeleteInt64(key string, opt ...OptionFunc) (int64, error) {
	return shortCut.DeleteInt64(key, opt...)
}

// GetInt64 通过key获取value，并将value解析成int64
// 支持 GetIgnoreExpireOpt IgnoreNotFoundOpt GetTTLOpt
// 当设置了 IgnoreNotFoundOpt 时，key不存在时会直接返回0
func GetInt64(key string, opt ...OptionFunc) (int64, error) {
	return shortCut.GetInt64(key, opt...)
}

// SetInt64 通过key设置int64格式的value
// 支持 SetExpireOpt SetKeepLastExpireOpt SetNoOverWriteOpt SetGetIsOverwriteOpt
// SetGetPreviousValueStringOpt SetGetPreviousValueInt64Opt SetGetPreviousValueJsonObjectOpt
func SetInt64(key string, value int64, opt ...OptionFunc) error {
	return shortCut.SetInt64(key, value, opt...)
}

// Delete 删除key，并返回key上的值
// 支持 IgnoreNotFoundOpt
func Delete(key string, opt ...OptionFunc) (string, error) {
	return shortCut.Delete(key, opt...)
}

// Get 通过key获取value
// 支持 GetIgnoreExpireOpt IgnoreNotFoundOpt GetTTLOpt
func Get(key string, opt ...OptionFunc) (string, error) {
	return shortCut.Get(key, opt...)
}

// Set 通过key设置value
// 支持 SetExpireOpt SetKeepLastExpireOpt SetNoOverWriteOpt SetGetIsOverwriteOpt
// SetGetPreviousValueStringOpt SetGetPreviousValueInt64Opt SetGetPreviousValueJsonObjectOpt
func Set(key, value string, opt ...OptionFunc) error {
	return shortCut.Set(key, value, opt...)
}

// Exist 查询key是否存在，key不存在或者发生任何错误时返回 false
// 支持 GetTTLOpt GetIgnoreExpireOpt
func Exist(key string, opt ...OptionFunc) bool {
	return shortCut.Exist(key, opt...)
}

// ExpireOption 是一个创建 buntdb.SetOptions 的函数糖，当直接操作底层buntdb的时候可以使用。
// 使用本package的时候请使用 SetExpireOpt
func ExpireOption(duration time.Duration) *buntdb.SetOptions {
	if duration <= 0 {
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
