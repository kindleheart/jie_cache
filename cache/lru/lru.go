package lru

import (
	"container/list"
	"jie_cache/cache"
)

type Cache struct {
	maxBytes  int64 // 缓存的最大内存, 0代表没有限制
	nBytes    int64 // 当前使用的内存
	ll        *list.List
	nodeMap   map[string]*list.Element
	OnEvicted func(key string, value cache.Value) // key被删除时的回调函数
}

type entry struct {
	key   string
	value cache.Value
}

func New(maxBytes int64, onEvicted func(string, cache.Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		nodeMap:   make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value cache.Value, ok bool) {
	if node, ok := c.nodeMap[key]; ok {
		c.ll.MoveToFront(node)
		kv := node.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) Add(key string, value cache.Value) {
	if node, ok := c.nodeMap[key]; ok {
		c.ll.MoveToFront(node)
		kv := node.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		c.nodeMap[key] = c.ll.PushFront(&entry{key, value})
		c.nBytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.removeOldest()
	}
}

func (c *Cache) removeOldest() {
	node := c.ll.Back()
	if node != nil {
		kv := c.ll.Remove(node).(*entry)
		delete(c.nodeMap, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
