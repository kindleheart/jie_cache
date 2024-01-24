package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map constains all hashed keys
type Map struct {
	hash     Hash
	replicas int             // 虚拟节点倍数
	keys     map[string]bool // 存储key
	hashRing []int           // Sorted 哈希环
	hashMap  map[int]string
}

// New creates a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
		keys:     make(map[string]bool),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		if m.keys[key] {
			continue
		}
		m.keys[key] = true
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.hashRing = append(m.hashRing, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.hashRing)
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica.
	idx := sort.Search(len(m.hashRing), func(i int) bool {
		return m.hashRing[i] >= hash
	})

	return m.hashMap[m.hashRing[idx%len(m.hashRing)]]
}

func (m *Map) Remove(key string) {
	if !m.keys[key] {
		return
	}
	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
		idx := sort.SearchInts(m.hashRing, hash)
		m.hashRing = append(m.hashRing[:idx], m.hashRing[idx+1:]...)
		delete(m.hashMap, hash)
	}
}
