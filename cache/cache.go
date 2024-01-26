package cache

import (
	"jie_cache/strategy"
	"jie_cache/strategy/lfu"
	"jie_cache/strategy/lru"
	"sync"
)

// 增加锁来实现并发安全
type Cache struct {
	mu        sync.Mutex
	baseCache strategy.BaseCache
	maxBytes  int64
	kind      string
}

const (
	LRU = "LRU"
	LFU = "LFU"
)

func New(kind string, maxBytes int64) *Cache {
	if kind != LRU && kind != LFU {
		panic("don't have this strategy")
	}
	return &Cache{
		mu:       sync.Mutex{},
		maxBytes: maxBytes,
		kind:     kind,
	}
}

func (c *Cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 延迟创建，节省内存
	if c.baseCache == nil {
		if c.kind == LRU {
			c.baseCache = lru.New(c.maxBytes, nil)
		} else {
			c.baseCache = lfu.New(c.maxBytes, nil)
		}
	}
	c.baseCache.Add(key, value)
}

func (c *Cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.baseCache == nil {
		return
	}

	if v, ok := c.baseCache.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
