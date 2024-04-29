package gocache

import "gocache/gocachepb"

/*
PeerPicker 是必须实现的接口，用于定位拥有特定密钥的对等方。
  - PeerPicker 的 PickPeer() 方法用于根据传入的 key 选择相应节点 PeerGetter。
  - PeerGetter 的 Get() 方法用于从对应 group 查找缓存值。PeerGetter 就对应于上述流程中的 HTTP 客户端。
*/

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}
type PeerGetter interface {
	Get(in *gocachepb.Request, out *gocachepb.Response) error
}
