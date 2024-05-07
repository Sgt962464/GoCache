package main

import (
	"flag"
	"fmt"
	"gocache/internal/service"
)

var (
	port = flag.Int("port", 9999, "service node default port")
	api  = flag.Bool("api", false, "Start a api server?")

	apiServerAddr1 = "http://127.0.0.1:8000"
	apiServerAddr2 = "http://127.0.0.1:8001"
)

func main() {
	flag.Parse()

	serverAddrMap := map[int]string{
		9999:  "http://127.0.0.1:9999",
		10000: "http://127.0.0.1:10000",
		10001: "http://127.0.0.1:10001",
	}
	var serverAddrs []string
	for _, v := range serverAddrMap {
		serverAddrs = append(serverAddrs, v)
	}
	gm := service.NewGroupManager([]string{"scores", "website"}, fmt.Sprintf("127.0.0.1:%d", *port))

	if *api {
		go service.StartHTTPAPIServer(apiServerAddr1, gm["scores"])
		go service.StartHTTPAPIServer(apiServerAddr2, gm["website"])
	}

	service.StartHTTPCacheServer(serverAddrMap[*port], []string(serverAddrs), gm["scores"])
	service.StartHTTPCacheServer(serverAddrMap[*port], []string(serverAddrs), gm["website"])
}
