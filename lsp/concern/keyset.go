package concern

type KeySet interface {
	GroupConcernStateKey(keys ...interface{}) string
	GroupConcernConfigKey(keys ...interface{}) string
	FreshKey(keys ...interface{}) string
	GroupAtAllMarkKey(keys ...interface{}) string
	ParseGroupConcernStateKey(key string) (groupCode int64, id interface{}, err error)
}
