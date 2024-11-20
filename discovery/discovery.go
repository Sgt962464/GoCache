package discovery

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"gocache/config"
	"gocache/utils/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
Discovery 通过 etcd 服务发现机制来建立 gRPC 客户端连接

  - 创建etcd解析器，使用 etcd 客户端实例 c 来创建一个 gRPC 解析器（resolver）
    将服务名（如 service）解析为实际的网络地址列表

  - 建立gRPC客户端连接

    -- grpc.WithResolvers(etcdResolver)：指定使用之前创建的 etcd 解析器来解析目标地址。

    -- grpc.WithTransportCredentials(insecure.NewCredentials())：配置 gRPC 客户端以使用不安全的传输凭据。这通常用于测试或开发环境，不建议在生产环境中使用。

    -- grpc.WithBlock()：使 Dial 函数阻塞，直到建立连接或发生错误。
*/
func Discovery(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		return nil, err
	}

	//return grpc.Dial(
	//	"etcd:///"+service,
	//	grpc.WithResolvers(etcdResolver),
	//	grpc.WithTransportCredentials(insecure.NewCredentials()),
	//	grpc.WithBlock(),
	//)
	return grpc.NewClient("etcd:///"+service, grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
}

/*
ListServicePeers 用来从 etcd 中检索并列出指定服务名的所有端点（peers）地址
  - 连接etcd
  - 创建端点管理器
  - 列出所有端点，endPointsManager 的 List 方法来获取服务名对应的所有端点映射
  - 遍历并收集端点地址
*/
func ListServicePeers(serviceName string) ([]string, error) {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Errorf("failed to connected to etcd, error: %v", err)
		return []string{}, err
	}

	// Endpoints are actually ip:port combinations, which can also be regarded as socket in Unix.
	// An endpoint manager stores both an etcd client object and the name of the requested service.
	endpointsManager, err := endpoints.NewManager(cli, serviceName)
	if err != nil {
		logger.LogrusObj.Errorf("create endpoints manager failed, %v", err)
		return []string{}, err
	}

	// List returns all endpoints of the current service in the form of a map.
	Key2EndpointMap, err := endpointsManager.List(context.Background())
	if err != nil {
		logger.LogrusObj.Errorf("list endpoint nodes for target service failed, error: %s", err.Error())
		return []string{}, err
	}

	var peersAddr []string
	for key, endpoint := range Key2EndpointMap {
		peersAddr = append(peersAddr, endpoint.Addr) // Addr is the server address on which a connection will be established.
		logger.LogrusObj.Infof("found endpoint addr: %s (%s):(%v)", key, endpoint.Addr, endpoint.Metadata)
	}

	return peersAddr, nil
}

/*
DynamicServices 动态监视etcd中特定服务名称键空间变化的函数。
  - 连接etcd
  - 使用了WithPrefix()选项，watch会监视具有指定前缀的所有键。
  - 处理watch事件
    --PUT事件，可能表示一个新的服务实例被添加或现有实例的信息被更新，向update通道发送一个true值，并记录一条警告日志
    --DELETE事件，向update通道发送一个true值，并记录一条警告日志，记录的是被删除键的键名
*/
func DynamicServices(update chan struct{}, service string) {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Errorf("failed to connected to etcd, error: %v", err)
		return
	}
	defer cli.Close()

	// Subscription and publishing mechanism.
	// Can also be seen as an observer pattern.
	// Monitor the changes of the {service} key or KV pairs prefixed with {service},
	// and return the corresponding events, notify through the returned channel.
	watchChan := cli.Watch(context.Background(), service, clientv3.WithPrefix())

	// Each time a user adds or removes a new instance address to a given service, the watchChan backend daemon
	// can scan for changes in the number of instances via WithPrefix() and return them as watchResp.Events events.
	for watchResp := range watchChan {
		for _, ev := range watchResp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				update <- struct{}{} //When a change occurs, send a signal to update channel telling endpoint manager to rebuild the hash map.
				logger.LogrusObj.Warnf("Service endpoint added or updated: %s", string(ev.Kv.Value))
			case clientv3.EventTypeDelete:
				update <- struct{}{} //When a change occurs, send a signal to update channel telling endpoint manager to rebuild the hash map.
				logger.LogrusObj.Warnf("Service endpoint removed: %s", string(ev.Kv.Key))
			}
		}
	}
}
