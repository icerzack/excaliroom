package inmemory

import (
	"sync"

	"go.uber.org/zap"
)

type Cache struct {
	mu     sync.RWMutex
	items  map[string]interface{}
	logger *zap.Logger
}

func NewCache(logger *zap.Logger) *Cache {
	return &Cache{
		items:  make(map[string]interface{}),
		logger: logger,
	}
}

func (c *Cache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = value
	c.logger.Debug("user added to cache", zap.String("key", key))
	return nil
}

func (c *Cache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		c.logger.Debug("user not found in cache", zap.String("key", key))
		return nil, nil
	}

	c.logger.Debug("user retrieved from cache", zap.String("key", key))
	return item, nil
}
