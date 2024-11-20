1. 编写xxx.prote文件
    1. 定义 Request 和 Response message 作为 rpc 从请求和响应的结构体
    2. 定义 GroupCache 服务，Get 方法，以请求结构体作为参数，以响应结构体作为返回值
2. 运行```  protoc --go_out=../groupcachepb --go_opt=paths=source_relative --go-grpc_out=../groupcachepb --go_opt=paths=source_relative groupcache.proto```
    1. --go_out 指定 xxx.pb.go 的输出路径，--go-grpc_out 指定了 xxx_grpc.pb.go 的输出路径
3. 实现RPC client和serve