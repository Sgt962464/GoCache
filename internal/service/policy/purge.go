package policy

import (
	"gocache/internal/service/policy/FIFO"
	"gocache/internal/service/policy/LFU"
	"gocache/internal/service/policy/LRU"
	"gocache/internal/service/policy/interfaces"
	"strings"
)

func New(name string, maxBytes int64, onEvicted func(string, interfaces.Value)) interfaces.CacheStrategy {
	name = strings.ToLower(name)
	switch name {
	case "lru":
		return LRU.NewLRUCache(maxBytes, onEvicted)
	case "lfu":
		return LFU.NewLFUCache(maxBytes, onEvicted)
	case "fifo":
		return FIFO.NewFIFOCache(maxBytes, onEvicted)
	}
	return nil
}
