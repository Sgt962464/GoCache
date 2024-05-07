package LRU

import (
	"container/list"
	"gocache/config"
	"gocache/internal/service/policy/interfaces"
	"gocache/utils/logger"
	"log"
	"sync"
	"time"
)

/*
LRUCache
  - map:存储键值映射关系，根据key查找value
  - 队列：双向链表实现，将所有值放入双向链表，访问某值，将其移动到队尾
*/
type LRUCache struct {
	maxBytes  int64      //允许使用的最大内存
	usedBytes int64      //已经使用的内存
	ll        *list.List //双向链表
	mu        sync.RWMutex
	cache     map[string]*list.Element

	// 回调函数，采用依赖注入的方式，该函数用于处理从缓存中淘汰的数据
	OnEvicted func(key string, value interfaces.Value)
}

func (c *LRUCache) usedLen(kv *interfaces.Entry) int64 {
	return int64(len(kv.Key)) + int64(kv.Value.Len())
}

/*
NewLRUCache
  - Cache的构造函数
*/
func NewLRUCache(maxBytes int64, onEvicted func(string, interfaces.Value)) *LRUCache {
	lru := &LRUCache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
	ttl := time.Duration(config.Conf.Services["groupcache"].TTL)
	//ttl := time.Duration(30)
	log.Printf("set ttl is %d s", ttl)

	go func() {
		ticker := time.NewTicker(time.Minute * 1)
		defer ticker.Stop()

		for {
			<-ticker.C
			lru.CleanUp(ttl)
			logger.LogrusObj.Warnf("触发过期缓存，清理后台任务......")
		}
	}()
	return lru
}

/*
Get

	*从map中找到节点
	*将其移动到队尾  (头Back  尾Front)
*/
func (c *LRUCache) Get(key string) (value interfaces.Value, updateAt *time.Time, ok bool) {
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element)
		kv := element.Value.(*interfaces.Entry)
		kv.Touch()
		return kv.Value, kv.UpdateAt, ok
	}
	return
}

/*
RemoveOldest
  移除最近、最少被访问的节点（队首）
*/

func (c *LRUCache) RemoveOldest() {
	element := c.ll.Back()
	if element != nil {
		c.mu.Lock()
		//还有 CleanUp 并发 goroutine 的过期淘汰策略，
		//因此需要进行并发安全双检，否则对 nil interface 进行断言直接触发
		//panic
		if element == nil {
			c.mu.Unlock()
			return
		}
		//c.ll.Remove(element)
		if kv, ok := element.Value.(*interfaces.Entry); ok {
			delete(c.cache, kv.Key)
			//c.usedBytes -= int64(len(kv.key)) + int64(kv.value.Len())
			c.usedBytes -= c.usedLen(kv)
			if c.OnEvicted != nil {
				c.OnEvicted(kv.Key, kv.Value)
			}
		}
		c.mu.Unlock()
	}
}

/*
Add

	向Cache中添加value
*/
func (c *LRUCache) Add(key string, value interfaces.Value) {
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element)
		if kv, isOK := element.Value.(*interfaces.Entry); isOK {
			kv.Touch()
			c.usedBytes += int64(value.Len()) - int64(kv.Value.Len())
			//c.usedBytes += c.usedLen(kv)
			kv.Value = value
		}
	} else {
		kv := &interfaces.Entry{Key: key, Value: value}
		kv.Touch()
		element := c.ll.PushFront(kv)
		c.cache[key] = element
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.usedBytes {
		c.RemoveOldest()
	}
}

func (c *LRUCache) Len() int {
	return c.ll.Len()
}

// TODO
func (c *LRUCache) CleanUp(ttl time.Duration) {
	for ele := c.ll.Front(); ele != nil; ele = ele.Next() {
		c.mu.Lock()
		if ele.Value == nil {
			c.mu.Unlock()
			return
		}
		if ele.Value.(*interfaces.Entry).Expired(ttl) {
			kv := c.ll.Remove(ele).(*interfaces.Entry)
			delete(c.cache, kv.Key)
			c.usedBytes -= c.usedLen(kv)
			if c.OnEvicted != nil {
				c.OnEvicted(kv.Key, kv.Value)
			}
		}
	}
	c.mu.Unlock()
}
