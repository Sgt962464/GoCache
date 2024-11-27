# Project introduction

- Fully automated node management has been implemented, with the ability to dynamically add and delete nodes and reconstruct consistent views, without the need to import instance addresses into etcd first

## test

Prerequisite: It is necessary to run the etcd cluster first, as detailed in etcd/cluster/use.md (stop the local 2379 etcd service first)

1. Run an etcd cluster(you can increase or decrease the number of cluster according to your own needs)，refer to etcd/cluster/use.md；
2. Initialize the student database used for testing，refer to  test/sql/create_sql.md;
3. The groupcache service and the student service communicate between processes via grpc,refer to api/use.md

## server start

- Run separately in three terminals
  - go run main.go -port 9999
  - go run main.go -port 10000
  - go run main.go -port 10001

- After all service instances have been started successfully, you can run the grpc client example for RPC call testing

In actual test cases:some student names were inserted into an array,then use rand.Shuffle (shuffling algorithm) to break up the students' names to better simulate the actual query situation.
- Unless the call fails, the test case will continue to run; Generally, it goes through two stages:
  - Preheating stage: All query request results are not cached, and all requests need to request the database (slow query) and then load from the database into the cache
  - Work stage: After the basic construction of the cache is completed, most hotspot requests will hit the cache and return directly, thus significantly speeding up the request processing speed
- The test case applies special handling to requests returned by the GRPC server that were not queried. Assuming that the GRPC server does not retrieve the grades of a student from the database and defaults to 'record not found', the client needs to intercept and process them based on the call results returned by GRPC to prevent panic
- Added retry mechanism (response can be obtained immediately after server crash recovery)
- Consider 'record not found' as a normal result rather than an error


### Test grpc communication

1. Can directly run `go run test/client/grpc_client.go` in the project root directory for testing
2. Alternatively, you can directly enter the test/client directory and run `./client.sh` for testing
