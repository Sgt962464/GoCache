package service

import (
	"context"
	"fmt"
	pb "gocache/api/groupcachepb"
	"gocache/internal/middleware/etcd/discovery/discovery3"
	"gocache/internal/service/consistenthash"
	"gocache/utils/logger"
	"gocache/utils/validate"
	"google.golang.org/grpc"
	"net"
	"strings"
	"sync"
	"time"
)

// 测试 Server 是否实现了 Picker 接口
var _Picker = (*Server)(nil)

/*
	服务器模块提供组缓存之间的通信功能。
	通过这种方式，部署在其他机器上的 groupcache 可以通过访问服务器来获得缓存。
	至于要找到哪一个主机，要靠一致性哈希。
*/

const (
	defaultBaseAddr = "127.0.0.1:9999"
	defaultReplicas = 50
)

type Server struct {
	pb.UnimplementedGroupCacheServer

	Addr       string     //format: ip:port
	Status     bool       //true:running    false:stop
	stopSignal chan error //通知register revoke服务
	mu         sync.Mutex
	consHash   *consistenthash.ConsistentHash
	clients    map[string]*Client
	update     chan bool
}

/*
	NewServer 将创建缓存服务器;如果addr为空，则使用默认的addr。
*/

func NewServer(update chan bool, addr string) (*Server, error) {
	if addr == "" {
		addr = defaultBaseAddr
	}
	if !validate.ValidPeerAddr(addr) {
		return nil, fmt.Errorf("invalid addr [ %s ],expect address format is x.x.x.x:port", addr)
	}
	return &Server{
		Addr:   addr,
		update: update,
	}, nil
}

/*
Get 处理来自客户端的 RPC 请求
  - 请求解析
  - 获取组实例
  - 从组中获取值，view, err := g.Get(key)
  - 构建响应
*/
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {

	group, key := req.GetGroup(), req.GetKey()
	resp := &pb.GetResponse{}
	logger.LogrusObj.Infof("[Groupcache server %s] Recv RPC Request - (%s)/(%s)", s.Addr, group, key)

	if key == "" || group == "" {
		return resp, fmt.Errorf("key and group name is reqiured")
	}
	g := GetGroup(group)
	if g == nil {
		return resp, fmt.Errorf("group %s not found", group)
	}

	view, err := g.Get(key)
	if err != nil {
		return resp, err
	}
	resp.Value = view.ByteSlice()
	return resp, nil
}

/*
SetPeers 将每个远程主机IP配置到服务器
  - 加锁并处理空的peer
  - 构建一致性哈希环
  - 创建客户端连接
*/
func (s *Server) SetPeers(peersAddr []string) {
	s.mu.Lock()

	if len(peersAddr) == 0 {
		peersAddr = []string{s.Addr}
	}

	s.consHash = consistenthash.NewConsistentHash(defaultReplicas, nil) //新的哈希环
	s.consHash.Add(peersAddr)                                           //添加对等节点（真实）

	s.clients = make(map[string]*Client)

	for _, addr := range peersAddr {
		if !validate.ValidPeerAddr(addr) {
			s.mu.Unlock()
			panic(fmt.Sprintf("[peer %s] invalid address format, it should be x.x.x.x:port", addr))
		}
		/*
			GroupCache/localhost:9999
			GroupCache/localhost:10000
			GroupCache/localhost:10001
			attention：服务发现原理建议看下 Endpoint 源码, key 是 service/addr value 是 addr
			服务解析时按照 service 进行前缀查询，找到所有服务节点
			而 clusters 前缀是为了拿到所有实例地址做一致性哈希使用的
			注意 service 要和在 protocol 文件中定义的服务名称一致
		*/

		// 使用固定的服务名称“GroupCache”创建客户端
		s.clients[addr] = NewClient("GroupCache")
	}
	s.mu.Unlock()

	/*
		<-s.update 重新构建服务器的一致性哈希环和客户端连接映射
		<-s.stopSignal 停止服务
		default:休眠2s
	*/
	go func() {
		for {
			select {
			case <-s.update:
				s.reconstruct()
			case <-s.stopSignal:
				s.Stop()
			default:
				time.Sleep(time.Second * 2)
			}
		}
	}()
}

/*
reconstruct 重新构建服务器的一致性哈希环和客户端连接映射
  - 获取服务实例列表
  - 加锁后重新构建哈希环并将从服务发现机制获取的服务实例列表 serviceList 添加到新的一致性哈希环中
  - 重新构建客户端映射连接
*/
func (s *Server) reconstruct() {
	serviceList, err := discovery3.ListServicePeers("GroupCache")
	if err != nil { // 如果没有拿到服务实例列表，暂时先维持当前视图
		return
	}

	s.mu.Lock()

	s.consHash = consistenthash.NewConsistentHash(defaultReplicas, nil)
	s.consHash.Add(serviceList)

	s.clients = make(map[string]*Client)

	for _, peerAddr := range serviceList {
		if !validate.ValidPeerAddr(peerAddr) {
			panic(fmt.Sprintf("[peer %s] invalid address format, expect x.x.x.x:port", peerAddr))
		}

		// demo: GroupCache/127.0.0.1:9999
		//地址有效，使用固定服务名称GroupCache和地址创建新的客户端实例，并添加到s.clients
		s.clients[peerAddr] = NewClient("GroupCache")
	}
	s.mu.Unlock()
	logger.LogrusObj.Infof("hash ring reconstruct, contain service peer %v", serviceList)

}

/*
Pick 根据给定的键 key 从一致性哈希环中选择一个对等节点（peer）
*/
func (s *Server) Pick(key string) (Fetcher, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 获取对等节点地址
	peerAddr := s.consHash.Get(key)

	if peerAddr == s.Addr || peerAddr == "" {
		logger.LogrusObj.Infof("oohhh! pick myself, i am %s", s.Addr)
		return nil, false
	}
	logger.LogrusObj.Infof("[current peer %s] pick remote peer: %s", s.Addr, peerAddr)
	// 返回选定的对等节点的客户端实例
	return s.clients[peerAddr], true
}

/*
Start 启动服务器并处理相关的初始化任务

	------------启动服务----------------
	1.设置服务器运行状态
	2.初始化停止通道以通知注册表停止保活租约
	3.初始化tcp套接字并开始侦听
	4.将自定义rpc服务注册到grpc，以便grpc可以将请求分发到服务器进行处理
	5.使用etcd服务注册表，它可以通过服务名称，即grpc通道，直接获得与给定服务的客户端连接，然后创建一个客户端Stub，它实现与服务器相同的方法并直接调用它
*/
func (s *Server) Start() {
	s.mu.Lock()
	if s.Status == true {
		s.mu.Unlock()
		fmt.Printf("server %s is already started", s.Addr)
		return
	}
	s.Status = true
	s.stopSignal = make(chan error)

	//监听端口
	port := strings.Split(s.Addr, ":")[1]
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("failed to listen %s, error: %v", s.Addr, err)
		return
	}

	//设置gRPC服务器
	grpcServer := grpc.NewServer()
	//将服务及其实现注册到实现GroupCacheServer接口的具体类型
	pb.RegisterGroupCacheServer(grpcServer, s)
	defer s.Stop()

	//服务注册
	go func() {
		//注册当前服务器实例
		err := discovery3.Register("GroupCache", s.Addr, s.stopSignal)
		if err != nil {
			logger.LogrusObj.Error(err.Error())
		}
		close(s.stopSignal)
		// 关闭tcp监听器
		err = lis.Close()
		if err != nil {
			logger.LogrusObj.Error(err.Error())
		}
		logger.LogrusObj.Warnf("[%s] Revoke service and close tcp socket", s.Addr)
	}()

	//解锁 启动gRPC服务
	//Serve接受侦听器列表上的传入连接，为每个连接创建一个新的服务器传输和服务Goroutine。
	//服务goroutines读取gRPC请求，然后调用注册的处理程序来回复它们。
	s.mu.Unlock()
	if err := grpcServer.Serve(lis); s.Status && err != nil {
		logger.LogrusObj.Fatalf("failed to serve %s, error: %v", s.Addr, err)
		return
	}
}
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.Status {
		return
	}

	// 通知registry停止释放keepalive
	s.stopSignal <- nil

	s.Status = false
	//清理资源，释放内存，可以帮助GC
	s.clients = nil
	s.consHash = nil
}
