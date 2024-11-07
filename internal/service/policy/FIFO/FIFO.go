package FIFO

import (
	"container/list"
	"gocache/internal/service/policy/interfaces"
	"time"
)

type fifoCahce struct {
	maxBytes  int64
	usedBytes int64
	ll        *list.List
	cache     map[string]*list.Element
	// optional and executed when an entry is purged.
	// 回调函数，采用依赖注入的方式，该函数用于处理从缓存中淘汰的数据
	OnEvicted func(key string, value interfaces.Value)
}

func NewFIFOCache(maxBytes int64, onEvicted func(key string, value interfaces.Value)) *fifoCahce {
	return &fifoCahce{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (f *fifoCahce) Get(key string) (value interfaces.Value, updateAt *time.Time, ok bool) {
	if elem, ok := f.cache[key]; ok {
		e := elem.Value.(*interfaces.Entry)
		return e.Value, e.UpdateAt, ok
	}
	return
}

func (f *fifoCahce) Add(key string, value interfaces.Value) {
	if elem, ok := f.cache[key]; ok {
		//更新cache
		f.usedBytes += int64(value.Len()) - int64(elem.Value.(*interfaces.Entry).Value.Len())
		elem.Value.(*interfaces.Entry).Value = value
	} else {
		kv := &interfaces.Entry{Key: key, Value: value, UpdateAt: nil}
		kv.Touch()
		elem := f.ll.PushBack(kv)
		f.cache[key] = elem
		f.usedBytes += int64(value.Len()) + int64(kv.Value.Len())
	}

	for f.maxBytes != 0 && f.usedBytes > f.maxBytes {
		f.RemoveFront()
	}
}

func (f *fifoCahce) CleanUp(ttl time.Duration) {
	for e := f.ll.Front(); e != nil; e = e.Next() {
		if e.Value.(*interfaces.Entry).Expired(ttl) {
			kv := f.ll.Remove(e).(*interfaces.Entry)
			delete(f.cache, kv.Key)
			f.usedBytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
			if f.OnEvicted != nil {
				f.OnEvicted(kv.Key, kv.Value)
			}
		} else {
			break
		}

	}
}
func (f *fifoCahce) RemoveFront() {
	elem := f.ll.Front()
	if elem != nil {
		kv := f.ll.Remove(elem).(*interfaces.Entry)
		delete(f.cache, kv.Key)
		f.usedBytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
		if f.OnEvicted != nil {
			f.OnEvicted(kv.Key, kv.Value)
		}
	}
}

func (f *fifoCahce) Len() int {
	return f.ll.Len()
}
