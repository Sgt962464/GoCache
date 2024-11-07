package main

import (
	"flag"
	"fmt"
	"gocache/config"
	"gocache/internal/middleware/etcd/discovery/discovery3"
	"gocache/internal/pkg/student/dao"
	"gocache/internal/service"
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
	gm := service.NewGroupManager([]string{"scores", "website"}, serviceAddr)

	//通过通信来共享内存而不是通过共享内存来通信
	updateChan := make(chan bool)
	svr, err := service.NewServer(updateChan, serviceAddr)
	if err != nil {
		logger.LogrusObj.Errorf("acquire grpc server instance failed, %v", err)
		//logger.LogrusObj
		return
	}

	go discovery3.DynamicServices(updateChan, config.Conf.Services["groupcache"].Name)

	peers, err := discovery3.ListServicePeers(config.Conf.Services["groupcache"].Name)
	if err != nil {
		peers = []string{"serviceAddr"}
	}

	svr.SetPeers(peers)

	gm["scores"].RegisterServer(svr)

	// start grpc service
	svr.Start()
}
