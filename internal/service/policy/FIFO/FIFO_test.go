package FIFO

import (
	"container/list"
	"fmt"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func Test_fifoCacheGet(t *testing.T) {
	cache := fifoCahce{
		maxBytes:  15,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: nil,
	}
	cache.Add("key1", String("1234"))
	if v, _, ok := cache.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	} else {
		fmt.Printf("cache hit key1=1234 ok,value=%s\n", v)
	}
	if _, _, ok := cache.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func Test_fifoCahce_RemoveFront(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	curcap := len(k1 + k2 + v1 + v2)
	f := NewFIFOCache(int64(curcap), nil)
	f.Add(k1, String(v1))
	f.Add(k2, String(v2))
	f.Add(k3, String(v3))
	if _, _, ok := f.Get("key1"); ok || f.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}
