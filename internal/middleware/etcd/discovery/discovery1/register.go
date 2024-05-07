package discovery1

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"gocache/utils/logger"
)

const (
	etcdUrl     = "http://localhost:2379"
	serviceName = "groupcache"
	ttl         = 10
)

var etcdClient *clientv3.Client

/*
EtcdRegister

  - 初始化etcd客户端：使用clientv3.NewFromURL(etcdUrl)来创建一个etcd客户端。
    但这里有个问题，etcdUrl是一个常量，而函数参数addr没有被使用。
  - 创建endpoints管理器：endpoints.NewManager(etcdClient, serviceName)
    用于创建一个管理etcd中endpoints的实例。
  - 授予租约：etcdClient.Grant(context.TODO(), ttl)
    授予一个TTL（time-to-live）为10秒的租约。
  - 添加endpoint：使用em.AddEndpoint将服务地址注册到etcd中，并且使用上面获得的租约ID。
  - 保持租约活跃：通过etcdClient.KeepAlive函数，代码启动了一个goroutine来持续更新租约，从而确保服务在etcd中的注册是活跃的。
    这个goroutine会一直阻塞在<-alive直到租约过期或etcd连接断开。
*/
func EtcdRegister(addr string) error {
	logger.LogrusObj.Debugf("EtcdRegister  %s\b", addr)
	etcdClient, err := clientv3.NewFromURL(etcdUrl)
	if err != nil {
		return err
	}

	em, err := endpoints.NewManager(etcdClient, serviceName)
	if err != nil {
		return err
	}

	lease, err := etcdClient.Grant(context.TODO(), ttl)
	if err != nil {
		return err
	}

	err = em.AddEndpoint(context.TODO(), fmt.Sprintf("%s/%s", serviceName, addr),
		endpoints.Endpoint{Addr: addr},
		clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	alive, err := etcdClient.KeepAlive(context.TODO(), lease.ID)
	if err != nil {
		return err
	}
	go func() {
		for {
			<-alive
		}
	}()
	return nil
}

/*
EtcdUnRegister
  - 删除endpoint：使用em.DeleteEndpoint从etcd中删除指定的服务地址。
*/
func EtcdUnRegister(addr string) error {
	logger.LogrusObj.Debugf("etcdUnRegister %s\b", addr)
	if etcdClient != nil {
		em, err := endpoints.NewManager(etcdClient, serviceName)
		if err != nil {
			return err
		}
		err = em.DeleteEndpoint(context.TODO(), fmt.Sprintf("%s/%s", serviceName, addr))
		if err != nil {
			return err
		}
		return err
	}

	return nil
}
