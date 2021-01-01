package concern_manager

type KeySet interface {
	ConcernStateKey(keys ...interface{}) string
	FreshKey(keys ...interface{}) string
	CurrentLiveKey(keys ...interface{}) string
	ParseConcernStateKey(key string) (int64, int64, error)
	ParseCurrentLiveKey(key string) (int64, error)
}
