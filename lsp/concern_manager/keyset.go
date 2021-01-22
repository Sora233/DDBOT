package concern_manager

type KeySet interface {
	GroupConcernStateKey(keys ...interface{}) string
	ConcernStateKey(keys ...interface{}) string
	FreshKey(keys ...interface{}) string
	ParseGroupConcernStateKey(key string) (groupCode int64, id interface{}, err error)
}
