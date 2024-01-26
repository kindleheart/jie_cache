package consistenthash

import (
	"hash/crc32"
	"log"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Consistent constains all hashed nodes
type Consistent struct {
	hash     Hash            // 哈希函数
	replicas int             // 虚拟节点倍数
	nodes    map[string]bool // 存储node
	hashRing []int           // Sorted 哈希环, 存储哈希值
	hashMap  map[int]string  // 存储哈希值和key的对应关系
}

// New creates a Consistent instance
func New(replicas int, fn Hash) *Consistent {
	m := &Consistent{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
		nodes:    make(map[string]bool),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some nodes to the hash.
func (m *Consistent) Add(nodes ...string) {
	for _, node := range nodes {
		if m.nodes[node] {
			continue
		}
		m.nodes[node] = true
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + node)))
			m.hashRing = append(m.hashRing, hash)
			m.hashMap[hash] = node
		}
	}
	sort.Ints(m.hashRing)
}

// Get gets the closest node in the hash to the provided key.
func (m *Consistent) Get(key string) string {
	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica.
	idx := sort.Search(len(m.hashRing), func(i int) bool {
		return m.hashRing[i] >= hash
	})

	return m.hashMap[m.hashRing[idx%len(m.hashRing)]]
}

// Remove delete node
func (m *Consistent) Remove(key string) {
	if !m.nodes[key] {
		log.Println("this key does not exist")
		return
	}
	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
		idx := sort.SearchInts(m.hashRing, hash)
		m.hashRing = append(m.hashRing[:idx], m.hashRing[idx+1:]...)
		delete(m.hashMap, hash)
	}
}
