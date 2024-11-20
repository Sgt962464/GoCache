package LFU

import (
	"fmt"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}
func TestPriorityQueue_Get(t *testing.T) {
	lfu := NewLFUCache(10, nil)
	lfu.Add("k1", String("1234"))
	if v, _, ok := lfu.Get("k1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	} else {
		fmt.Printf("cache hit key1=%v\n", v)
	}

	if _, _, ok := lfu.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}
func TestPriorityQueue_Remove(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	curCap := len(k1 + k2 + v1 + v2)

	lfu := NewLFUCache(int64(curCap), nil)
	lfu.Add(k1, String(v1))
	lfu.Add(k1, String(v1))
	lfu.Add(k2, String(v2))
	lfu.Add(k3, String(v3))

	if _, _, ok := lfu.Get("key2"); ok || lfu.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}
