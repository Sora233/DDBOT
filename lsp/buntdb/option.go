package buntdb

import (
	"github.com/tidwall/buntdb"
	"strconv"
	"time"
)

type option struct {
	ignoreExpire   bool
	noOverWrite    bool
	expire         time.Duration
	keepLastExpire bool
	previous       interface{}
	ignoreNotFound bool
}

func (o *option) getIgnoreExpire() bool {
	if o == nil {
		return false
	}
	return o.ignoreExpire
}

func (o *option) getNoOverWrite() bool {
	if o == nil {
		return false
	}
	return o.noOverWrite
}

func (o *option) getExpire() time.Duration {
	if o == nil {
		return 0
	}
	return o.expire
}

func (o *option) getInnerExpire() *buntdb.SetOptions {
	if o == nil || o.expire == 0 {
		return nil
	}
	return &buntdb.SetOptions{
		Expires: true,
		TTL:     o.expire,
	}
}

func (o *option) getIgnoreNotFound() bool {
	if o == nil {
		return false
	}
	return o.ignoreNotFound
}

func (o *option) setPrevious(previous string) {
	if o.previous == nil {
		return
	}
	switch ptr := o.previous.(type) {
	case *int64:
		i, err := strconv.ParseInt(previous, 10, 64)
		if err == nil {
			*ptr = i
		} else {
			logger.Errorf("setPrevious int64 on value %v error %v", previous, err)
		}
	case *string:
		*ptr = previous
	default:
		_ = json.Unmarshal([]byte(previous), ptr)
	}
}

type OptionFunc func(o *option)

func emptyOptionFunc(o *option) {}

// SetExpireOpt 设置set时的过期时间
// 该设置与 SetKeepLastExpireOpt 同时设置时，本设置会生效
func SetExpireOpt(expire time.Duration) OptionFunc {
	return func(o *option) {
		o.expire = expire
	}
}

// SetKeepLastExpireOpt 设置set时保留上次的过期时间
// 该设置与 SetKeepLastExpireOpt 同时设置时， SetExpireOpt 设置会生效
func SetKeepLastExpireOpt() OptionFunc {
	return func(o *option) {
		o.keepLastExpire = true
	}
}

// SetNoOverWriteOpt Set配置，当key已经存在时不进行覆盖，而是返回 ErrRollback
func SetNoOverWriteOpt() OptionFunc {
	return func(o *option) {
		o.noOverWrite = true
	}
}

func SetGetPreviousValueStringOpt(previous *string) OptionFunc {
	if previous == nil {
		return emptyOptionFunc
	}
	return func(o *option) {
		o.previous = previous
	}
}

func SetGetPreviousValueInt64Opt(previous *int64) OptionFunc {
	if previous == nil {
		return emptyOptionFunc
	}
	return func(o *option) {
		o.previous = previous
	}
}

func SetGetPreviousValueJsonObjectOpt(previous interface{}) OptionFunc {
	if previous == nil {
		return emptyOptionFunc
	}
	return func(o *option) {
		o.previous = previous
	}
}

func GetIgnoreExpireOpt() OptionFunc {
	return func(o *option) {
		o.ignoreExpire = true
	}
}

func GetIgnoreNotFound() OptionFunc {
	return func(o *option) {
		o.ignoreNotFound = true
	}
}

func DeleteIgnoreNotFound() OptionFunc {
	return func(o *option) {
		o.ignoreNotFound = true
	}
}

func getOption(opts ...OptionFunc) *option {
	var s = new(option)
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	return s
}
