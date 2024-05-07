package service

import (
	"gocache/utils/logger"
	"log"
	"net/http"
)

// StartHTTPCacheServer 启动基于HTTP的缓存服务器
func StartHTTPCacheServer(addr string, addrs []string, gocache *Group) {
	peers := NewHTTPPool(addr)

	peers.UpdatePeers(addrs...)
	gocache.RegisterServer(peers)

	logger.LogrusObj.Infof("service is running at %v ", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// TODO: gin 路由拆分请求负载
// StartHTTPAPIServer 简单的 HTTP 服务器，用于处理对 /api 路径的 GET 请求，并从 gocache 中检索数据。
func StartHTTPAPIServer(apiAddr string, gocache *Group) {
	//设置路由和处理器
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			/*
				- 从请求 URL 的查询参数中获取 key。
				- gocache.Get(key) 来从缓存中获取数据
				- 成功获取数据，设置响应头
				- 使用 w.Write(view.ByteSlice()) 将数据写入响应体
			*/
			key := r.URL.Query().Get("key")
			view, err := gocache.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	logger.LogrusObj.Infof("fontend server is running at %v", apiAddr)
	// 启动
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}
