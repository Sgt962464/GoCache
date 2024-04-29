package LRU

import (
	"container/list"
)

/*
Cache
  - map:存储键值映射关系，根据key查找value
  - 队列：双向链表实现，将所有值放入双向链表，访问某值，将其移动到队尾
*/
type Cache struct {
	maxBytes  int64      //允许使用的最大内存
	usedBytes int64      //已经使用的内存
	ll        *list.List //双向链表
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}
type entry struct {
	key   string
	value Value
}

// 用于计算每个value所占字节数
type Value interface {
	Len() int
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
func (c *Cache) usedLen(kv *entry) int64 {
	return int64(len(kv.key)) + int64(kv.value.Len())
}

/*
New
  - Cache的构造函数
*/
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

/*
Get

	*从map中找到节点
	*将其移动到队尾
*/
func (c *Cache) Get(key string) (value Value, ok bool) {
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element)
		kv := element.Value.(*entry)
		return kv.value, ok
	}
	return
}

/*
RemoveOldest
  移除最近、最少被访问的节点（队首）
*/

func (c *Cache) RemoveOldest() {
	element := c.ll.Back()
	if element != nil {
		c.ll.Remove(element)
		if kv, ok := element.Value.(*entry); ok {
			delete(c.cache, kv.key)
			//c.usedBytes -= int64(len(kv.key)) + int64(kv.value.Len())
			c.usedBytes -= c.usedLen(kv)
			if c.OnEvicted != nil {
				c.OnEvicted(kv.key, kv.value)
			}
		}

	}
}

/*
Add

	向Cache中添加value
*/
func (c *Cache) Add(key string, value Value) {
	if element, ok := c.cache[key]; ok {
		c.ll.Remove(element)
		if kv, isOK := element.Value.(*entry); isOK {
			c.usedBytes += int64(value.Len()) - int64(kv.value.Len())
			//c.usedBytes += c.usedLen(kv)
			kv.value = value
		}
	} else {
		element := c.ll.PushFront(&entry{key, value})
		c.cache[key] = element
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.usedBytes {
		c.RemoveOldest()
	}
}
