package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 哈希值要映射到2^32空间中
type Hash func(data []byte) uint32

/*
Consistence 包含所有经过hash的key
*/
type Consistence struct {
	hash     Hash
	replicas int            //虚拟节点倍数
	keys     []int          //哈希环
	hashMap  map[int]string //虚拟节点与真实节点的映射 map[vnode]rnode
}

func New(replicas int, f Hash) *Consistence {
	m := &Consistence{
		replicas: replicas,
		hash:     f,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE //TODO 默认使用ChecksumIEEE算法
	}

	return m
}

/*
Add 添加真实节点

  - 允许传入0个或多个真实节点名称
  - 对每一个真实节点，创建replicas个虚拟节点，虚拟节点名称是strconv.Itoa(i) + key
  - 计算虚拟节点的hash值并添加至环
  - 在hashMap中添加虚拟节点与真实节点的映射值
  - 环上的hash value排序
*/
func (m *Consistence) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}

	}
	sort.Ints(m.keys)
}

/*
Get 选择节点
  - 计算key的hash value
  - 顺时针找到第一个匹配的虚拟节点下标idx，从m.keys中获取对应的hash value
  - 通过hashMap映射得到真实节点/‘
*/
func (m *Consistence) Get(key string) string {
	if len(key) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	//sort.Search的结果取值范围是 [0,len(m.keys)]
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	//idx==len(m.keys)  应该选择m.keys[0]
	return m.hashMap[m.keys[idx%len(m.keys)]]
}

func (m *Consistence) Remove(key string) {
	if len(key) == 0 {
		return
	}

	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(key + strconv.Itoa(i))))
		idx := sort.SearchInts(m.keys, hash)
		if m.keys[idx] != idx {
			return
		}
		m.keys = append(m.keys[:idx], m.keys[idx+1:]...)
		delete(m.hashMap, hash)
	}
}
