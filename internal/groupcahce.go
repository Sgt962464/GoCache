package service

import (
	"errors"
	"fmt"
	"gocache/utils/logger"
	"gorm.io/gorm"
	"sync"
	"time"
)

var (
	mu           sync.RWMutex
	GroupManager = make(map[string]*Group)
)

/*
Getter 缓存不存在时，调用Getter,得到源数据
*/
type Getter interface {
	Get(key string) ([]byte, error)
}
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

/*
Group
  - Group是缓存命名空间，相关数据加载至此,每个 Group 拥有一个唯一的名称 name
  - getter Getter，即缓存未命中时获取源数据的回调(callback)。
  - mainCache cache，并发缓存。
  - 节点
*/
type Group struct {
	name      string
	mainCache *cache
	retriever Retriever
	server    Picker
	flight    *SingleFlight
}

// RegisterServer 注册一个 server Picker  ,用以选择远程对等节点
func (g *Group) RegisterServer(peers Picker) {
	if g.server != nil {
		panic("group had been registered server")
	}
	g.server = peers
}

// NewGroup :创建Group实例

func NewGroup(name string, strategy string, maxBytes int64, retriever Retriever) *Group {
	if retriever == nil {
		panic("Group Retriver must be existed!")
	}
	if _, ok := GroupManager[name]; ok {
		return GroupManager[name]
	}
	g := &Group{
		name:      name,
		mainCache: newCache(strategy, maxBytes),
		retriever: retriever,
		flight:    NewSingleFlight(time.Second * 10),
	}

	mu.Lock()
	GroupManager[name] = g
	mu.Unlock()
	return g
}

// GetGroup :返回 NewGroup 创建的group，没有则返回nil
func GetGroup(name string) *Group {
	mu.RLock()
	g := GroupManager[name]
	mu.RUnlock()
	return g
}

//func DestroyGroup(name string) {
//	g := GetGroup(name)
//	if g != nil {
//		svr := g.server.(*Server)
//		svr.Stop()
//
//		delete(GroupManager, name)
//	}
//}

/*
Get
 - 从mainCache中查找缓存，存在则返回缓存值
 - 缓存不存在，调用load，load调用getLocally（分布式场景下调用getFromPeer从
   其他节点获取），getLocally调用回调函数getter.Get获取数据源，并将源数据添加
   到缓存mainCache中（通过populateCache方法）

1	                            是
2	接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
3	                |  否                         是
4	                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
5	                            |  否
6	                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶
*/

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required!")
	}
	if v, ok := g.mainCache.get(key); ok {
		logger.LogrusObj.Infof("[GoCache] Group %s cache hit....,key %s ...", g.name, key)
		return v, nil
	}

	// cache未命中
	return g.load(key)
}

// load 方法，使用 PickPeer 方法选择节点，若非本机节点，则调用 getFromPeer() 从远程获取。
// 若是本机节点或失败，则回退到 getLocally()。
func (g *Group) load(key string) (value ByteView, err error) {
	// 每个key仅被获取一次
	view, err := g.flight.Do(key, func() (interface{}, error) {
		if g.server != nil {
			if fetcher, ok := g.server.Pick(key); ok {
				//fmt.Println(66666666666666)
				bytes, err := fetcher.Fetch(g.name, key)
				if err == nil {
					return ByteView{b: cloneBytes(bytes)}, nil
				}
				logger.LogrusObj.Warnf("fetch key %s failed, error: %s\n", fetcher, err.Error())
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}
	return ByteView{}, err
}

// getFromPeer 访问远程节点，获取缓存
//func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
//	req := &pb.Request{
//		Group: g.name,
//		Key:   key,
//	}
//	res := &pb.Response{}
//	err := peer.Get(req, res)
//	if err != nil {
//		return ByteView{}, err
//	}
//	return ByteView{b: res.Value}, err
//}

// getLocally 调用回调函数getter.Get获取数据源，并将源数据添加到缓存mainCache中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.retriever.retrieve(key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.LogrusObj.Warnf("对于不存在的 key, 为了防止缓存穿透, 先存入缓存中并设置合理过期时间")
			g.mainCache.add(key, ByteView{})
		}
	}

	value := ByteView{b: cloneBytes(bytes)}

	g.populateCache(key, value)

	return value, nil
}

/*
populateCache 填充缓存使用从基础数据库查询的数据填充缓存
*/
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
