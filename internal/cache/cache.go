package cache

type Cache interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
	SetWithTTL(key string, value interface{}, ttl int64) error
}
