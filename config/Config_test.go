package config

import (
	"fmt"
	"testing"
)

func TestConf(t *testing.T) {
	InitConfig()
	if Conf == nil {
		t.Fatal("Conf is not initialized")
	}

	// 然后检查 Conf.Services 是否为 nil
	if Conf.Services == nil {
		t.Fatal("Conf.Services is not initialized")
	}

	// 现在安全地尝试访问 Conf.Services["groupcache"]
	if serviceConfig, ok := Conf.Services["groupcache"]; ok {
		// 在解引用之前检查 serviceConfig 是否为 nil
		if serviceConfig != nil {
			// 直接访问TTL字段
			fmt.Println(serviceConfig.TTL)
		} else {
			t.Error("serviceConfig is nil")
		}
	} else {
		fmt.Println("groupcache service not found in configuration")
	}
}
