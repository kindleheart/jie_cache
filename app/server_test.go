package app

import (
	"fmt"
	"jie_cache/cache"
	"log"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestNewServer(t *testing.T) {
	cache.NewGroup("school", cache.LRU, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	server := NewServer("", "localhost:8080")
	server.Start()
}
