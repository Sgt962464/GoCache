package LFU

import (
	"container/heap"
	"gocache/internal/policy/interfaces"
	"time"
)

type LFUCache struct {
	maxBytes  int64 //允许使用的最大内存
	usedBytes int64 //已经使用的内存
	cache     map[string]*lfuEntry
	pq        *priorityqueue
	OnEvicted func(key string, value interfaces.Value)
}

func NewLFUCache(maxBytes int64, onEvicted func(string, interfaces.Value)) *LFUCache {
	queue := priorityqueue(make([]*lfuEntry, 0))
	return &LFUCache{
		maxBytes:  maxBytes,
		pq:        &queue,
		cache:     make(map[string]*lfuEntry),
		OnEvicted: onEvicted,
	}
}

func (p *LFUCache) Get(key string) (value interfaces.Value, updateAt *time.Time, ok bool) {
	if e, ok := p.cache[key]; ok {
		e.Referenced()
		heap.Fix(p.pq, e.index)
		return e.entry.Value, e.entry.UpdateAt, ok
	}
	return
}

func (p *LFUCache) Add(key string, value interfaces.Value) {
	if e, ok := p.cache[key]; ok {
		p.usedBytes += int64(value.Len()) - int64(e.entry.Value.Len())
		e.entry.Value = value
		e.Referenced()
		heap.Fix(p.pq, e.index)
	} else {
		e := &lfuEntry{0, interfaces.Entry{Key: key, Value: value, UpdateAt: nil}, 0}
		e.Referenced()
		heap.Push(p.pq, e)
		p.cache[key] = e
		p.usedBytes += int64(len(e.entry.Key)) + int64(e.entry.Value.Len())
	}

	for p.maxBytes != 0 && p.maxBytes < p.usedBytes {
		p.Remove()
	}
}

func (p *LFUCache) CleanUp(ttl time.Duration) {
	for _, e := range *p.pq {
		if e.entry.Expired(ttl) {
			kv := heap.Remove(p.pq, e.index).(*lfuEntry).entry
			delete(p.cache, kv.Key)
			p.usedBytes -= int64(len(kv.Key)) + int64(kv.Value.Len())
			if p.OnEvicted != nil {
				p.OnEvicted(kv.Key, kv.Value)
			}
		}
	}
}

func (p *LFUCache) Remove() {
	e := heap.Pop(p.pq).(*lfuEntry)
	delete(p.cache, e.entry.Key)
	p.usedBytes -= int64(len(e.entry.Key)) + int64(e.entry.Value.Len())
	if p.OnEvicted != nil {
		p.OnEvicted(e.entry.Key, e.entry.Value)
	}
}

func (p *LFUCache) Len() int {
	return p.pq.Len()
}
