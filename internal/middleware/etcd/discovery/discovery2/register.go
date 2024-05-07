package discovery2

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"gocache/config"
	"gocache/utils/logger"
)

/*
 Register 负责将服务注册到etcd中，并通过keepalive机制保持租约活动状态。通过监听多个通道，它可以响应停止信号、处理连接中断和服务列表变化等事件
	- 创建etcd客户端，使用默认配置
	- 创建租约，成功时，将租约ID保存至leaseId
	- 注册服务到etcd
	- 设置keepalive 心跳检测，确保租约保持活跃状态
	- 创建服务管理器并监听服务变化

*/

func Register(service string, addr string, stop chan error) error {
	cli, err := clientv3.New(config.DefaultEtcdConfig)
	if err != nil {
		return fmt.Errorf("create etcd client falied: %v", err)
	}
	defer cli.Close()

	resp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		return fmt.Errorf("create lease failed: %v", err)
	}
	leaseId := resp.ID

	// 注意：如果将重建哈希环操作放在 etcdAdd 之后，那么就无需用户手动 put 实例地址了
	err = etcdAdd(cli, leaseId, service, addr)
	if err != nil {
		panic(err)
	}

	ch1, err := cli.KeepAlive(context.Background(), leaseId)
	if err != nil {
		return fmt.Errorf("set keepalive failed: %v", err)
	}

	// manager 管理服务的注册和注销
	//watchChan 监听服务变化
	manager, _ := endpoints.NewManager(cli, service)
	watchChan, _ := manager.NewWatchChannel(context.Background())

	/*
		for循环 处理事件
			- stop通道接收到信号，则停止服务并返回相应的错误
			- 客户端的上下文Done通道接收到信号，表示客户端连接已关闭，打印日志并返回nil
			- keepalive通道关闭，表示租约已过期或被撤销，撤销租约并返回错误（如果有的话）
			- watchChan接收到信号，表示服务列表已更改，可以更新本地缓存或执行其他操作。
	*/
	for {
		select {
		case err := <-stop:
			if err != nil {
				logger.LogrusObj.Errorf(err.Error())
			}
			return err
		case <-cli.Ctx().Done():
			logger.LogrusObj.Infof("service closed")
			return nil
		case _, ok := <-ch1: // 监听租约撤销信号
			if !ok {
				logger.LogrusObj.Info("keepalive channel closed")
				_, err := cli.Revoke(context.Background(), leaseId)
				return err
			}
		case <-watchChan:
			// map[string]Endpoint
			key2EndpointMap, _ := manager.List(context.Background())
			var addrs []string
			for _, endpoint := range key2EndpointMap {
				addrs = append(addrs, endpoint.Addr)
			}

		}
	}
}

/*
etcdAdd 用于在 etcd 中添加服务端点的工具函数
  - 创建服务管理器
  - 添加端点
*/
func etcdAdd(client *clientv3.Client, leaseId clientv3.LeaseID, service string, addr string) error {
	endPointsManager, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}

	return endPointsManager.AddEndpoint(client.Ctx(),
		fmt.Sprintf("%s/%s", service, addr),
		/*
			Addr 是将在其上建立连接的服务器地址。
			Metadata 是与Addr相关联的信息，可用于做出负载平衡决策。
			Endpoint 表示可以用来建立连接的单个地址。
		*/
		endpoints.Endpoint{Addr: addr, Metadata: "GroupCache services"},
		//租约选项（lease option）：使用 clientv3.WithLease(leaseId) 来指定租约 ID，
		//这样 etcd 中的这个端点项就会有一个过期时间，当租约过期时，这个项会自动从 etcd 中删除。
		clientv3.WithLease(leaseId))
}

/*
etcdDel 用于从etcd中删除指定服务地址的端点的工具函数
  - 创建服务管理器
  - 删除端点
*/
func etcdDel(client *clientv3.Client, service string, addr string) error {
	endPointsManager, err := endpoints.NewManager(client, service)
	if err != nil {
		return err
	}
	//通过fmt.Sprintf构造了一个键（key），服务名和地址的组合，格式为service/addr
	return endPointsManager.DeleteEndpoint(client.Ctx(),
		fmt.Sprintf("%s/%s", service, addr), nil)
}
