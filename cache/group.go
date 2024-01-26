package cache

import (
	"fmt"
	"jie_cache/pb"
	"jie_cache/peer"
	"jie_cache/singleflight"
	"log"
	"sync"
)

type Group struct {
	name       string
	getter     Getter
	mainCache  *Cache
	peerPicker peer.PeerPicker
	single     *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, kind string, maxBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: New(kind, maxBytes),
		single:    new(singleflight.Group),
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[JieCache] hit")
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
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}
