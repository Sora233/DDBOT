package concern_manager

type KeySet interface {
	GroupConcernStateKey(keys ...interface{}) string
	ConcernStateKey(keys ...interface{}) string
	FreshKey(keys ...interface{}) string
	CurrentLiveKey(keys ...interface{}) string
	ParseGroupConcernStateKey(key string) (int64, int64, error)
	ParseCurrentLiveKey(key string) (int64, error)
}
