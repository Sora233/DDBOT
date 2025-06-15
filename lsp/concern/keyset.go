package concern

import (
	localdb "github.com/Sora233/DDBOT/v2/lsp/buntdb"
)

// KeySet 是不同 StateManager 之间用来彼此隔离的一个接口。
// 通常 StateManager 会创建多个，用于不同的 Concern 模块，所以创建 StateManager 的时候需要指定 KeySet。
// 大多数情况下可以方便的使用内置实现 PrefixKeySet。
type KeySet interface {
	GroupConcernStateKey(keys ...interface{}) string
	GroupConcernConfigKey(keys ...interface{}) string
	FreshKey(keys ...interface{}) string
	GroupAtAllMarkKey(keys ...interface{}) string
	ParseGroupConcernStateKey(key string) (groupCode uint32, id interface{}, err error)
}

// PrefixKeySet 是 KeySet 的一个默认实现，它使用一个唯一的前缀彼此区分。
type PrefixKeySet struct {
	prefix                string
	groupConcernStateKey  string
	groupConcernConfigKey string
	freshKey              string
	groupAtAllMarkKey     string
	parser                func(key string) (groupCode uint32, id interface{}, err error)
}

func (p *PrefixKeySet) GroupConcernStateKey(keys ...interface{}) string {
	return localdb.NamedKey(p.groupConcernStateKey, keys)
}

func (p *PrefixKeySet) GroupConcernConfigKey(keys ...interface{}) string {
	return localdb.NamedKey(p.groupConcernConfigKey, keys)
}

func (p *PrefixKeySet) FreshKey(keys ...interface{}) string {
	return localdb.NamedKey(p.freshKey, keys)
}

func (p *PrefixKeySet) GroupAtAllMarkKey(keys ...interface{}) string {
	return localdb.NamedKey(p.groupAtAllMarkKey, keys)
}

func (p *PrefixKeySet) ParseGroupConcernStateKey(key string) (groupCode uint32, id interface{}, err error) {
	return p.parser(key)
}

func newPrefixKeySet(prefix string, parser func(key string) (groupCode uint32, id interface{}, err error)) *PrefixKeySet {
	p := &PrefixKeySet{
		prefix: prefix,
		parser: parser,
	}
	p.groupConcernStateKey = p.prefix + "GroupConcernState"
	p.groupConcernConfigKey = p.prefix + "GroupConcernConfig"
	p.freshKey = p.prefix + "FreshKey"
	p.groupAtAllMarkKey = p.prefix + "GroupAtAllMark"
	return p
}

// NewPrefixKeySetWithStringID 根据 prefix 创建一个使用 string 格式的id的 PrefixKeySet
// id的格式需要与 Concern.ParseId 返回的格式一致
// prefix 可以简单地使用 Concern.Site
func NewPrefixKeySetWithStringID(prefix string) *PrefixKeySet {
	return newPrefixKeySet(prefix, func(key string) (groupCode uint32, id interface{}, err error) {
		return localdb.ParseConcernStateKeyWithString(key)
	})
}

// NewPrefixKeySetWithInt64ID 根据 prefix 创建一个使用 int64 格式的id的 PrefixKeySet
// id的格式需要与 Concern.ParseId 返回的格式一致
// prefix 可以简单地使用 Concern.Site
func NewPrefixKeySetWithInt64ID(prefix string) *PrefixKeySet {
	return newPrefixKeySet(prefix, func(key string) (groupCode uint32, id interface{}, err error) {
		return localdb.ParseConcernStateKeyWithInt64(key)
	})
}
