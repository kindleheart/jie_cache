package lfu

import (
	"container/list"
	"jie_cache/cache"
)

type Cache struct {
	maxBytes  int64 // 缓存的最大内存, 0代表没有限制
	nBytes    int64 // 当前使用的内存
	minFreq   int   // 最少访问频率
	listMap   map[int]*list.List
	nodeMap   map[string]*list.Element
	OnEvicted func(key string, value cache.Value) // key被删除时的回调函数
}

type entry struct {
	key   string
	value cache.Value
	freq  int
}

func New(maxBytes int64, onEvicted func(key string, value cache.Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		listMap:   make(map[int]*list.List),
		nodeMap:   make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (cache.Value, bool) {
	if node, ok := c.nodeMap[key]; ok {
		kv := node.Value.(*entry)
		c.MoveNodeToNextLevel(node)
		return kv.value, true
	}
	return nil, false
}

func (c *Cache) getList(freq int) *list.List {
	if c.listMap[freq] == nil {
		c.listMap[freq] = list.New()
	}
	return c.listMap[freq]
}

func (c *Cache) Add(key string, value cache.Value) {
	if node, ok := c.nodeMap[key]; ok {
		kv := node.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
		c.MoveNodeToNextLevel(node)
	} else {
		ll := c.getList(1)
		c.minFreq = 1
		c.nBytes += int64(len(key)) + int64(value.Len())
		c.nodeMap[key] = ll.PushFront(&entry{
			key:   key,
			value: value,
			freq:  1,
		})
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.removeOldest()
	}
}

func (c *Cache) MoveNodeToNextLevel(node *list.Element) {
	kv := node.Value.(*entry)
	c.listMap[kv.freq].Remove(node)
	if c.listMap[c.minFreq].Len() == 0 {
		c.minFreq++
	}
	kv.freq++
	ll := c.getList(kv.freq)
	c.nodeMap[kv.key] = ll.PushFront(kv)
}

func (c *Cache) removeOldest() {
	ll := c.listMap[c.minFreq]
	if ll == nil {
		return
	}
	node := ll.Back()
	if node != nil {
		kv := node.Value.(*entry)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		delete(c.nodeMap, kv.key)
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}
