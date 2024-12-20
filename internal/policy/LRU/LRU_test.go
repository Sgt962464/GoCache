package LRU

import (
	"fmt"
	"gocache/config"
	"gocache/internal/policy/interfaces"
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	config.InitConfig()
	lru := NewLRUCache(int64(15), nil)
	lru.Add("key1", String("1234"))
	if v, _, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveoldest(t *testing.T) {
	config.InitConfig()
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	curcap := len(k1 + k2 + v1 + v2)
	lru := NewLRUCache(int64(curcap), nil)
	fmt.Println(lru.maxBytes)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))
	if _, _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		fmt.Println(lru.cache["key2"])
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	config.InitConfig()
	keys := make([]string, 0)
	callback := func(key string, value interfaces.Value) {
		keys = append(keys, key)
	}
	lru := NewLRUCache(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}

}
