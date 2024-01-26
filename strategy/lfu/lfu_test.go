package lfu

import (
	"fmt"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	lfu := New(int64(9), nil)
	lfu.Add("key1", String("1"))
	fmt.Println(lfu.nBytes)
	lfu.Add("key2", String("2"))
	if _, ok := lfu.Get("key1"); ok {
		t.Fatal("key1 hasn't not been phased out")
	}
	if v, ok := lfu.Get("key2"); !ok || string(v.(String)) != "2" {
		t.Fatal("nodeMap hit key2=2 failed")
	}
	if _, ok := lfu.Get("key3"); ok {
		t.Fatal("nodeMap miss key3 failed")
	}
}
