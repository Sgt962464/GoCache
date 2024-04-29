package gocache

import (
	"fmt"
	pb "gocache/gocachepb"
	"gocache/singlefilght"
	"log"
	"sync"
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
	getter    Getter
	mainCache cache
	peers     PeerPicker
	loader    *singlefilght.Group //保证每个key仅被获取一次
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// RegisterPeers 注册一个 PeerPicker ,用以选择远程对等节点
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// NewGroup :创建Group实例

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singlefilght.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup :返回 NewGroup 创建的group，没有则返回nil
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

/*
Get
 - 从mainCache中查找缓存，存在则返回缓存值
 - 缓存不存在，调用load，load调用getLocally（分布式场景下调用getFromPeer从
   其他节点获取），getLocally调用回调函数getter.Get获取数据源，并将源数据添加
   到缓存mainCache中（通过populateCache方法）
*/

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required!")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GoCache] hit")
		return v, nil
	}
	return g.load(key)
}

// load 方法，使用 PickPeer 方法选择节点，若非本机节点，则调用 getFromPeer() 从远程获取。
// 若是本机节点或失败，则回退到 getLocally()。
func (g *Group) load(key string) (value ByteView, err error) {
	// 每个key仅被获取一次
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				//fmt.Println(66666666666666)
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GoCache] Failed to get from peer ", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// getFromPeer 访问远程节点，获取缓存
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, err
}

// getLocally 调用回调函数getter.Get获取数据源，并将源数据添加到缓存mainCache中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
