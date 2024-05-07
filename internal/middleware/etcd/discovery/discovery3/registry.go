package discovery3

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"gocache/config"
	"gocache/utils/logger"
	"time"
)

/*
Register 服务注册

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
		logger.LogrusObj.Fatalf("error: %v", err)
		return err
	}
	// 创建新租约
	resp, err := cli.Grant(context.Background(), 5)
	if err != nil {
		return fmt.Errorf("grant creates a new lease failed: %v", err)
	}

	leaseId := resp.ID
	err = etcdAdd(cli, leaseId, service, addr)
	if err != nil {
		return fmt.Errorf("failed to add services as endpoint to etcd endpoint Manager: %v", err)
	}

	ch, err := cli.KeepAlive(context.Background(), leaseId)
	if err != nil {
		return fmt.Errorf("set keepalive for lease failed: %v", err)
	}

	logger.LogrusObj.Debugf("[%s] register service success", addr)

	/*
		循环处理各种情况
		 - <-stop：如果接收到服务撤销的信号（通过stop channel传递），则调用etcdDel函数从Etcd中删除服务
		 - <-cli.Ctx().Done()：如果Etcd客户端连接断开，则记录错误并返回
		 - <-ch:用于接收租约的心跳响应。如果channel关闭（即Etcd服务可能出现问题），则调用etcdDel函数删除服务，并返回错误。
		 - default: 如果没有上述情况发生，则使线程休眠200毫秒，以避免空转
	*/
	for {
		select {
		case err := <-stop: // service revocation signal
			etcdDel(cli, service, addr)
			if err != nil {
				logger.LogrusObj.Error(err.Error())
			}
			return err
		case <-cli.Ctx().Done(): // etcd client connect 断开
			return fmt.Errorf("etcd client connect broken")
		case _, ok := <-ch: // lease keepalive responses
			if !ok {
				logger.LogrusObj.Error("keepalive channel closed, revoke given lease") // 比如 etcd 断开服务，通知 server 停止
				etcdDel(cli, service, addr)
				return fmt.Errorf("keepalive channel closed, revoke given lease") // 返回非 nil 的 error，上层就会关闭 stopsChan 从而关闭 server
			}
		default:
			time.Sleep(200 * time.Millisecond)
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
	return endPointsManager.AddEndpoint(context.TODO(),
		fmt.Sprintf("%s/%s", service, addr),
		/*
			Addr 是将在其上建立连接的服务器地址。
			Metadata 是与Addr相关联的信息，可用于做出负载平衡决策。
			Endpoint 表示可以用来建立连接的单个地址。
		*/
		endpoints.Endpoint{Addr: addr, Metadata: "gocache services"},
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
