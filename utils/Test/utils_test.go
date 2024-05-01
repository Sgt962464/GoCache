package Test

import (
	"fmt"
	"gocache/utils/trace"
	"gocache/utils/validate"
	"testing"
)

func TestTrace(t *testing.T) {
	errStr := trace.Trace("an error occurred:")
	fmt.Println(errStr)
}

func TestValidate(t *testing.T) {
	addrs := []string{
		"127.0.0.1:8080",
		"localhost:80",
		"::1:8080",        // IPv6本地地址
		"192.168.1.1",     // 缺少端口号
		"127.0.0.1:65536", // 端口号超出范围
		"example.com:80",  // 非IP地址
	}

	for _, addr := range addrs {
		fmt.Printf("Is '%s' a valid peer address? %t\n", addr, validate.ValidPeerAddr(addr))
	}
}
