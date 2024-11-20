package service

import (
	"gocache/utils/logger"
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 哈希值映射到2^32的空间中
type Hash func(data []byte) uint32

/*
ConsistentHash 包含所有经过hash的key
*/
type ConsistentHash struct {
	hash         Hash
	replicas     int            //虚拟节点倍数
	virtualNodes []int          //哈希环
	hashMap      map[int]string //虚拟节点与真实节点的映射
}

func NewConsistentHash(replicas int, f Hash) *ConsistentHash {

	if f == nil {
		f = crc32.ChecksumIEEE //TODO 默认使用这个算法，后续可自行实现其他算法
	}
	return &ConsistentHash{
		hash:     f,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
}

/*
Add 添加真实节点

  - 允许传入0个或多个真实节点名称
  - 对每一个真实节点，创建replicas个虚拟节点，虚拟节点名称是strconv.Itoa(i) + key
  - 计算虚拟节点的hash值并添加至环
  - 在hashMap中添加虚拟节点与真实节点的映射值
  - 环上的hash value排序
*/
func (m *ConsistentHash) AddTruthNode(keys []string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.virtualNodes = append(m.virtualNodes, hash)
			m.hashMap[hash] = key
		}

	}
	sort.Ints(m.virtualNodes)
}

/*
Get 选择节点
  - 计算key的hash value
  - 顺时针找到第一个匹配的虚拟节点下标idx，从m.virtualNodes中获取对应的hash value
  - 通过hashMap映射得到真实节点/‘
*/
func (m *ConsistentHash) GetTruthNode(key string) string {
	if len(m.virtualNodes) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	//sort.Search的结果取值范围是 [0,len(m.virtualNodes)]
	idx := sort.Search(len(m.virtualNodes), func(i int) bool {
		return m.virtualNodes[i] >= hash
	})
	logger.LogrusObj.Infof("计算出 key 的 hash: %d, 顺时针选择的虚拟节点下标 idx: %d", hash, idx)
	//logger.LogrusObj.Infof("2322626082 2871910706 3693793700 4252452532")
	logger.LogrusObj.Infof("选择的真实节点：%s", m.hashMap[m.virtualNodes[idx%len(m.virtualNodes)]])
	//idx==len(m.virtualNodes)  应该选择m.virtualNodes[0]
	return m.hashMap[m.virtualNodes[idx%len(m.virtualNodes)]]
}

func (m *ConsistentHash) RemovePeer(key string) {
	// 将真实节点从 hash 环中删除
	// logger.LogrusObj.Warn("peers:", v)
	virtualHash := []int{}
	for key, v := range m.virtualNodes {
		logger.LogrusObj.Warn("peers: ", v)
		if v == key {
			delete(m.hashMap, key)
			virtualHash = append(virtualHash, key)
		}
	}

	for i := 0; i < len(virtualHash); i++ {
		for index, value := range virtualHash {
			if value == virtualHash[i] {
				m.virtualNodes = append(m.virtualNodes[:index], m.virtualNodes[index+1:]...)
			}
		}
	}
}
