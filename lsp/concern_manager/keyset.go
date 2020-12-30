package concern_manager

type KeySet interface {
	ConcernStateKey(keys ...interface{}) string
	ParseConcernStateKey(key string) (int64, int64, error)
	FreshKey(keys ...interface{}) string
	ParseCurrentLiveKey(key string) (int64, error)
	CurrentLiveKey(keys ...interface{}) string
}
