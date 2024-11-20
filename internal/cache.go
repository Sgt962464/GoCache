package service

import (
	"gocache/internal/policy"
	"gocache/internal/policy/interfaces"
	"gocache/utils/logger"
	"sync"
)

// lru上层并发上一层锁
type cache struct {
	mu         sync.Mutex
	strategy   interfaces.CacheStrategy
	cacheBytes int64
}

func newCache(strategy string, cacheSize int64) *cache {
	onEvicted := func(key string, value interfaces.Value) {
		logger.LogrusObj.Infof("缓存条目 [%s:%s] 被淘汰", key, value)
	}

	return &cache{
		cacheBytes: cacheSize,
		strategy:   policy.New(strategy, cacheSize, onEvicted),
	}
}
func (c *cache) set(key string, value ByteView) {
	c.mu.Lock()
	c.strategy.Add(key, value)
	c.mu.Unlock()
}
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.LogrusObj.Infof("存入数据库之后压入缓存, (key, value)=(%s, %s)", key, value)
	c.strategy.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, _, ok := c.strategy.Get(key); ok {
		return v.(ByteView), true
	} else {
		return ByteView{}, false
	}
}
