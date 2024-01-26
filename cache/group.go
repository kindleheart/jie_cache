package cache

import (
	"fmt"
	"jie_cache/pb"
	"jie_cache/peer"
	"jie_cache/singleflight"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type Group struct {
	name               string
	getter             Getter
	mainCache          *Cache
	hotCache           *Cache
	peerPicker         peer.PeerPicker
	single             *singleflight.Group
	stats              map[string]*keyStats
	maxMinuteRemoteQPS int
}

type AtomicInt int64 // 封装一个原子类，用于进行原子操作，保证并发安全.

// Add 方法用于对 AtomicInt 中的值进行原子自增
func (i *AtomicInt) Add(n int64) { //原子自增
	atomic.AddInt64((*int64)(i), n)
}

// Get 方法用于获取 AtomicInt 中的值。
func (i *AtomicInt) Get() int64 {
	return atomic.LoadInt64((*int64)(i))
}

type keyStats struct {
	firstGetTime time.Time // 第一次请求的时间
	remoteCnt    AtomicInt // 远程调用的次数
}

const (
	MAX_MINUTE_REMOTE_QPS = 10
	MAX_BYTES             = 2 << 10
)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheType string, getter Getter, options ...Option) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:               name,
		getter:             getter,
		mainCache:          New(cacheType, int64(MAX_BYTES)),
		hotCache:           New(cacheType, int64(MAX_BYTES/8)),
		single:             new(singleflight.Group),
		stats:              make(map[string]*keyStats),
		maxMinuteRemoteQPS: MAX_MINUTE_REMOTE_QPS,
	}
	for _, option := range options {
		option(g)
	}
	groups[name] = g
	return g
}

// Functional Options 来初始化参数
type Option func(g *Group)

func MaxBytes(maxBytes int64) Option {
	return func(g *Group) {
		g.mainCache.maxBytes = maxBytes
		g.hotCache.maxBytes = maxBytes / 8
	}
}

func MaxMinuteRemoteQPS(qps int) Option {
	return func(g *Group) {
		g.maxMinuteRemoteQPS = qps
	}
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 函数用于获取缓存数据，获取顺序为：热点缓存、主缓存、数据源
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.hotCache.get(key); ok {
		log.Println("[JieCache] hit hotCache")
		return v, nil
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[JieCache] hit mainCache")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	val, err := g.single.Do(key, func() (interface{}, error) {
		if g.peerPicker != nil {
			if peerGetter, ok := g.peerPicker.PickPeer(key); ok {
				if view, err := g.getFromPeer(peerGetter, key); err == nil {
					return view, nil
				} else {
					log.Println("[JieCache] Failed to get from peer", err)
				}
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return val.(ByteView), nil
	}
	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)}
	g.mainCache.add(key, value)
	return value, nil
}

func (g *Group) RegisterPeerPicker(picker peer.PeerPicker) {
	if g.peerPicker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peerPicker = picker
}

func (g *Group) getFromPeer(peer peer.PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	resp := &pb.Response{}
	err := peer.Get(req, resp)
	if err != nil {
		return ByteView{}, err
	}
	// 更新查询其他节点key的统计数据
	if stat, ok := g.stats[key]; ok {
		stat.remoteCnt.Add(1)
		// 计算QPS，判断是否加入hotCache
		//计算QPS
		interval := float64(time.Now().Unix()-stat.firstGetTime.Unix()) / 60
		qps := stat.remoteCnt.Get() / int64(math.Max(1, math.Round(interval)))
		if qps >= int64(g.maxMinuteRemoteQPS) {
			//存入hotCache
			g.hotCache.add(key, ByteView{b: resp.Value})
			//删除映射关系,节省内存
			mu.Lock()
			delete(g.stats, key)
			mu.Unlock()
		}
	} else {
		// 第一次获取
		g.stats[key] = &keyStats{
			firstGetTime: time.Now(),
			remoteCnt:    1,
		}
	}

	return ByteView{b: resp.Value}, nil
}
