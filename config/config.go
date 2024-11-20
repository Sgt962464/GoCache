package config

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
	"time"

	"os"
	"path"
	"path/filepath"
)

var Conf *Config
var DefaultEtcdConfig clientv3.Config //etcd客户端配置
var once sync.Once

type Config struct {
	Mysql    *MySQL              `yaml:"mysql"`
	Etcd     *Etcd               `yaml:"etcd"`
	Services map[string]*Service `yaml:"services"`
	Domain   map[string]*Domain  `yaml:"domain"`
}

type MySQL struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Charset  string `yaml:"charset"`
}

type Etcd struct {
	Address []string `yaml:"address"`
	TTL     int      `yaml:"ttl"`
}

type Service struct {
	Name         string   `yaml:"name"`
	LoadBalancer bool     `yaml:"loadBalancer"`
	Addr         []string `yaml:"addr"`
	TTL          int      `yaml:"ttl"`
}

type Domain struct {
	Name string `yaml:"name"`
}

func InitConfig() {
	rootDir := findRootDir()
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(path.Join(rootDir, "config"))
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	// parse into Conf object
	err = viper.Unmarshal(&Conf)
	if err != nil {
		panic(err)
	}

	InitClientV3Config()

}
func InitClientV3Config() {
	once.Do(func() {
		DefaultEtcdConfig = clientv3.Config{
			Endpoints:   Conf.Etcd.Address,
			DialTimeout: time.Second * 5, //超时时间为5s
		}
	})
}

// 从当前工作目录开始，逐级向上查找，直到找到go.mod
func findRootDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir
		}
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			panic("reached top of file system without finding go.mod")
		}
		currentDir = parentDir
	}

}
