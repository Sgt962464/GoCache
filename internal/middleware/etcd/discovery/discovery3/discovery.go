package discovery3

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"gocache/config"
	"gocache/utils/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"math/rand"
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

	return grpc.Dial(
		"etcd:///"+service,
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
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

	endPointsManager, err := endpoints.NewManager(cli, serviceName)
	if err != nil {
		logger.LogrusObj.Errorf("create endpoints manager failed, %v", err)
		return []string{}, err
	}

	Key2EndPointMap, err := endPointsManager.List(context.Background())
	if err != nil {
		logger.LogrusObj.Errorf("enpoint manager list op failed, %v", err)
		return []string{}, err
	}

	var peers []string
	for key, endpoint := range Key2EndPointMap {
		peers = append(peers, endpoint.Addr)
		logger.LogrusObj.Infof("found endpoint %s (%s):(%s)", key, endpoint.Addr, endpoint.Metadata)
	}

	return peers, nil
}

/*
DynamicServices 动态监视etcd中特定服务名称键空间变化的函数。
  - 连接etcd
  - 使用了WithPrefix()选项，watch会监视具有指定前缀的所有键。
  - 处理watch事件
    --PUT事件，可能表示一个新的服务实例被添加或现有实例的信息被更新，向update通道发送一个true值，并记录一条警告日志
    --DELETE事件，向update通道发送一个true值，并记录一条警告日志，记录的是被删除键的键名
*/
func DynamicServices(update chan bool, service string) {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		logger.LogrusObj.Errorf("failed to connected to etcd, error: %v", err)
		return
	}
	defer cli.Close()

	watchChan := cli.Watch(context.Background(), service, clientv3.WithPrefix())

	// 每次用户往指定的服务中添加或者删除新的实例地址时，watchChan 后台都能通过 WithPrefix() 扫描到实例数量的变化并以  watchResp.Events 事件的方式返回
	// 当发生变更时，往 update channel 发送一个信号，告知 endpoint manager 重新构建哈希映射
	for watchResp := range watchChan {
		for _, ev := range watchResp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				update <- true // 通知 endpoint manager 重新构建节点视图
				logger.LogrusObj.Warnf("Service endpoint added or updated: %s", string(ev.Kv.Value))
			case clientv3.EventTypeDelete:
				update <- true // 通知 endpoint manager 重新构建节点视图
				logger.LogrusObj.Warnf("Service endpoint removed: %s", string(ev.Kv.Key))
			}
		}
	}
}

func shuffle(peers []string) string {
	rand.Shuffle(len(peers), func(i, j int) {
		peers[i], peers[j] = peers[j], peers[i]
	})
	return peers[len(peers)/2]
}
