package main

import (
	"flag"
	"fmt"
	"gocache/config"
	"gocache/discovery"
	grpcservice "gocache/internal"
	"gocache/test/pkg/student/dao"
	"gocache/utils/logger"
)

var (
	port = flag.Int("port", 9999, "service node port")
)

func main() {
	config.InitConfig()
	dao.InitDB()
	flag.Parse()

	serviceAddr := fmt.Sprintf("localhost:%d", *port)
	gm := grpcservice.NewGroupManager([]string{"scores", "website"}, serviceAddr)

	//通过通信来共享内存而不是通过共享内存来通信
	updateChan := make(chan struct{})
	svr, err := grpcservice.NewServer(updateChan, serviceAddr)
	if err != nil {
		logger.LogrusObj.Errorf("acquire grpc server instance failed, %v", err)
		//logger.LogrusObj
		return
	}

	go discovery.DynamicServices(updateChan, config.Conf.Services["groupcache"].Name)

	peers, err := discovery.ListServicePeers(config.Conf.Services["groupcache"].Name)
	if err != nil {
		peers = []string{"serviceAddr"}
	}

	svr.SetPeers(peers)

	gm["scores"].RegisterServer(svr)

	// start grpc service
	svr.Start()
}
