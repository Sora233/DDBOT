package concern

import (
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
)

type KeySet interface {
	GroupConcernStateKey(keys ...interface{}) string
	GroupConcernConfigKey(keys ...interface{}) string
	FreshKey(keys ...interface{}) string
	GroupAtAllMarkKey(keys ...interface{}) string
	ParseGroupConcernStateKey(key string) (groupCode int64, id interface{}, err error)
}

type PrefixKeySet struct {
	prefix string
	parser func(key string) (groupCode int64, id interface{}, err error)

	groupConcernStateKey  string
	groupConcernConfigKey string
	freshKey              string
	groupAtAllMarkKey     string
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

func (p *PrefixKeySet) ParseGroupConcernStateKey(key string) (groupCode int64, id interface{}, err error) {
	return p.parser(key)
}

func newPrefixKeySet(prefix string, parser func(key string) (groupCode int64, id interface{}, err error)) *PrefixKeySet {
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

func NewPrefixKeySetWithStringID(prefix string) *PrefixKeySet {
	return newPrefixKeySet(prefix, func(key string) (groupCode int64, id interface{}, err error) {
		return localdb.ParseConcernStateKeyWithString(key)
	})
}

func NewPrefixKeySetWithInt64ID(prefix string) *PrefixKeySet {
	return newPrefixKeySet(prefix, func(key string) (groupCode int64, id interface{}, err error) {
		return localdb.ParseConcernStateKeyWithInt64(key)
	})
}
