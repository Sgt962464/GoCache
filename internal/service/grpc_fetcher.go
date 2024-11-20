package service

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	pb "gocache/api/groupcachepb"
	"gocache/discovery"
	"gocache/utils/logger"
	"time"
)

// 测试 Client 是否实现了 Fetcher 接口
var _ Fetcher = (*Client)(nil)

// Client 模块实现了groupcache访问其他远程节点以获取缓存的能力。
type Client struct {
	serviceName string // 服务名称 groupcache/ip:addr
}

func NewClient(service string) *Client {
	return &Client{service}
}

/*
Fetch 从gRPC服务中根据group和key获取数据
  - 创建Etcd客户端
  - 服务发现
  - 创建gRPC客户端并设置超时
  - 调用gRPC服务
*/
func (c *Client) Fetch(group string, key string) ([]byte, error) {
	cli, err := clientv3.NewFromURL("http://localhost:2379")
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	start := time.Now()
	conn, err := discovery.Discovery(cli, c.serviceName)
	logger.LogrusObj.Warnf("本次 grpc dial 的耗时为: %v ms", time.Since(start).Milliseconds())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	grpcClient := pb.NewGroupCacheClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	start = time.Now()
	resp, err := grpcClient.Get(ctx, &pb.GetRequest{
		Group: group,
		Key:   key,
	})
	logger.LogrusObj.Warnf("本次 grpc Call 的耗时为: %v ms", time.Since(start).Milliseconds())

	if err != nil {
		return nil, fmt.Errorf("could not get %s/%s from peer %s", group, key, c.serviceName)
	}

	return resp.Value, nil
}
