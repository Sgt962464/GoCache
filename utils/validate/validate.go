package validate

import (
	"net"
	"strconv"
	"strings"
)

// ValidPeerAddr 检验地址是否格式正确
func ValidPeerAddr(addr string) bool {
	parts := strings.SplitN(addr, ":", 2)
	if len(parts) != 2 {
		return false
	}
	ip := parts[0]
	port := parts[1]

	if (net.ParseIP(ip) == nil && ip != "localhost" && ip != "127.0.0.1") || !isValidPort(port) {
		return false
	}
	return true
}

func isValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	return err == nil && p >= 0 && p <= 65535
}
