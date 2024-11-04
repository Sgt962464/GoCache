package LFU

import (
	"container/heap"
	"gocache/internal/service/policy/interfaces"
	"time"
)

type LFUCache struct {
	maxBytes  int64 //允许使用的最大内存
	usedBytes int64 //已经使用的内存
	cache     map[string]*lfuEntry
	pd        *priorityqueue
	OnEvicted func(key string, value interfaces.Value)
}

func (p *LFUCache) Len() int {
	return p.pd.Len()
}

func NewLFUCache(maxBytes int64, onEvicted func(string, interfaces.Value)) *LFUCache {
	queue := priorityqueue(make([]*lfuEntry, 0))
	return &LFUCache{
		maxBytes:  maxBytes,
		pd:        &queue,
		cache:     make(map[string]*lfuEntry),
		OnEvicted: onEvicted,
	}
}

func (p *LFUCache) Get(key string) (value interfaces.Value, updateAt *time.Time, ok bool) {
	if e, ok := p.cache[key]; ok {
		e.Referenced()
		heap.Fix(p.pd, e.index)
		return e.entry.Value, e.entry.UpdateAt, ok
	}
	return
}
