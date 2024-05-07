1. 编写xxx.prote文件
2. 运行``` protoc --go_out=../studentpb --go_opt=paths=source_relative --go-grpc_out=../studentpb --go_opt=paths=source_relative student.proto```    
   1. --go_out 指定 xxx.pb.go 的输出路径，--go-grpc_out 指定了 xxx_grpc.pb.go 的输出路径
3. 实现RPC client和server