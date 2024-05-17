package inmemory

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type Item struct {
	Value      interface{}
	Expiration int64
}

type Cache struct {
	mu     sync.RWMutex
	items  map[string]*Item
	logger *zap.Logger
}

func NewCache(logger *zap.Logger) *Cache {
	return &Cache{
		items:  make(map[string]*Item),
		logger: logger,
	}
}

func (c *Cache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &Item{
		Value: value,
	}
	c.logger.Debug("User added to cache", zap.String("key", key))
	return nil
}

func (c *Cache) SetWithTTL(key string, value interface{}, ttl int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &Item{
		Value:      value,
		Expiration: time.Now().UnixNano() + ttl*int64(time.Second),
	}
	c.logger.Debug("User added to cache", zap.String("key", key))
	return nil
}

func (c *Cache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok || item.Expiration < time.Now().UnixNano() {
		c.logger.Debug("User not found in cache", zap.String("key", key))
		return nil, nil
	}

	return item.Value, nil
}
