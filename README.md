## Start
### 开启etcd集群
> cd /GoCache/internal/middleware/etcd/cluster      
> goreman -f Procfile start


### 进入/GoCache
> go run main.go -port 9999     
> go run main.go -port 10000     
> go run main.go -port 10001
 
### 进入GoCache/script
> sh test3.sh