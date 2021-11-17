package buntdb

import (
	"github.com/tidwall/buntdb"
	"strconv"
	"time"
)

type option struct {
	ignoreExpire   bool
	noOverWrite    bool
	isOverWrite    *bool
	expire         time.Duration
	keepLastExpire bool
	previous       interface{}
	ignoreNotFound bool
	ttl            *time.Duration
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
	if o == nil || o.expire <= 0 {
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

func (o *option) getTTL() *time.Duration {
	if o == nil {
		return nil
	}
	return o.ttl
}

func (o *option) setPrevious(previous string) {
	if o == nil || o.previous == nil || len(previous) == 0 {
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

func (o *option) setIsOverWrite(replaced bool) {
	if o == nil || o.isOverWrite == nil {
		return
	}
	*o.isOverWrite = replaced
}

func (o *option) setTTL(ttl time.Duration) {
	if o == nil || o.ttl == nil || ttl == 0 {
		return
	}
	if ttl < 0 {
		ttl = 0
	}
	*o.ttl = ttl
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

// SetGetPreviousValueStringOpt Set配置，存储key上一个值到previous中
func SetGetPreviousValueStringOpt(previous *string) OptionFunc {
	if previous == nil {
		return emptyOptionFunc
	}
	return func(o *option) {
		o.previous = previous
	}
}

// SetGetPreviousValueInt64Opt Set配置，将key的上一个值解析到int64并放到previous中
func SetGetPreviousValueInt64Opt(previous *int64) OptionFunc {
	if previous == nil {
		return emptyOptionFunc
	}
	return func(o *option) {
		o.previous = previous
	}
}

// SetGetPreviousValueJsonObjectOpt Set配置，将key的上一个值用json解析并放到previous中
func SetGetPreviousValueJsonObjectOpt(previous interface{}) OptionFunc {
	if previous == nil {
		return emptyOptionFunc
	}
	return func(o *option) {
		o.previous = previous
	}
}

// SetGetIsOverwriteOpt Set配置，获取此次set是否将覆盖一个旧值，可以与 SetNoOverWriteOpt 同时设置
func SetGetIsOverwriteOpt(isOverWrite *bool) OptionFunc {
	if isOverWrite == nil {
		return emptyOptionFunc
	}
	return func(o *option) {
		o.isOverWrite = isOverWrite
	}
}

// GetIgnoreExpireOpt Get配置，忽略key的过期时间，如果key曾经设置过但已经过期，依然可以获取到
func GetIgnoreExpireOpt() OptionFunc {
	return func(o *option) {
		o.ignoreExpire = true
	}
}

// IgnoreNotFoundOpt 获取值时不返回 buntdb.ErrNotFound ，而是返回nil
func IgnoreNotFoundOpt() OptionFunc {
	return func(o *option) {
		o.ignoreNotFound = true
	}
}

// GetTTLOpt 获取key上的ttl
func GetTTLOpt(ttl *time.Duration) OptionFunc {
	return func(o *option) {
		o.ttl = ttl
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
