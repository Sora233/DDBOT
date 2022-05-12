package concern

import (
	"errors"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"strconv"
	"strings"
)

// KeySet 是不同 StateManager 之间用来彼此隔离的一个接口。
// 通常 StateManager 会创建多个，用于不同的 Concern 模块，所以创建 StateManager 的时候需要指定 KeySet。
// 大多数情况下可以方便的使用内置实现 PrefixKeySet。
type KeySet interface {
	ConcernStateKey(keys ...interface{}) string
	ConcernConfigKey(keys ...interface{}) string
	FreshKey(keys ...interface{}) string
	AtAllMarkKey(keys ...interface{}) string
	ParseConcernStateKey(key string) (target mt.Target, id interface{}, err error)
}

// PrefixKeySet 是 KeySet 的一个默认实现，它使用一个唯一的前缀彼此区分。
type PrefixKeySet struct {
	prefix                string
	concernStateKey       string
	groupConcernConfigKey string
	freshKey              string
	groupAtAllMarkKey     string
	parser                func(key string) (target mt.Target, id interface{}, err error)
}

func (p *PrefixKeySet) ConcernStateKey(keys ...interface{}) string {
	return localdb.NamedKey(p.concernStateKey, keys)
}

func (p *PrefixKeySet) ConcernConfigKey(keys ...interface{}) string {
	return localdb.NamedKey(p.groupConcernConfigKey, keys)
}

func (p *PrefixKeySet) FreshKey(keys ...interface{}) string {
	return localdb.NamedKey(p.freshKey, keys)
}

func (p *PrefixKeySet) AtAllMarkKey(keys ...interface{}) string {
	return localdb.NamedKey(p.groupAtAllMarkKey, keys)
}

func (p *PrefixKeySet) ParseConcernStateKey(key string) (target mt.Target, id interface{}, err error) {
	return p.parser(key)
}

func newPrefixKeySet(prefix string, parser func(key string) (target mt.Target, id interface{}, err error)) *PrefixKeySet {
	p := &PrefixKeySet{
		prefix: prefix,
		parser: parser,
	}
	p.concernStateKey = p.prefix + "ConcernState"
	p.groupConcernConfigKey = p.prefix + "ConcernConfig"
	p.freshKey = p.prefix + "FreshKey"
	p.groupAtAllMarkKey = p.prefix + "AtAllMark"
	return p
}

// NewPrefixKeySetWithStringID 根据 prefix 创建一个使用 string 格式的id的 PrefixKeySet
// id的格式需要与 Concern.ParseId 返回的格式一致
// prefix 可以简单地使用 Concern.Site
func NewPrefixKeySetWithStringID(prefix string) *PrefixKeySet {
	return newPrefixKeySet(prefix, func(key string) (target mt.Target, id interface{}, err error) {
		return ParseConcernStateKeyWithString(key)
	})
}

// NewPrefixKeySetWithInt64ID 根据 prefix 创建一个使用 int64 格式的id的 PrefixKeySet
// id的格式需要与 Concern.ParseId 返回的格式一致
// prefix 可以简单地使用 Concern.Site
func NewPrefixKeySetWithInt64ID(prefix string) *PrefixKeySet {
	return newPrefixKeySet(prefix, func(key string) (target mt.Target, id interface{}, err error) {
		return ParseConcernStateKeyWithInt64(key)
	})
}

var errInvalidKey = errors.New("invalid key")

func ParseConcernStateKeyWithInt64(key string) (target mt.Target, id int64, err error) {
	keys := strings.Split(key, ":")
	if len(keys) != 3 {
		return nil, 0, errInvalidKey
	}
	target = mt.ParseTargetHash(keys[1])
	if target == nil {
		return nil, 0, errInvalidKey
	}
	id, err = strconv.ParseInt(keys[2], 10, 64)
	if err != nil {
		return nil, 0, err
	}
	return target, id, nil
}
func ParseConcernStateKeyWithString(key string) (target mt.Target, id string, err error) {
	keys := strings.Split(key, ":")
	if len(keys) != 3 {
		return nil, "", errInvalidKey
	}
	target = mt.ParseTargetHash(keys[1])
	if target == nil {
		return nil, "", errInvalidKey
	}
	return target, keys[2], nil
}
