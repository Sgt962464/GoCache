package conf

import (
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"

	"os"
	"path"
	"path/filepath"
	"time"
)

var Conf *Config
var DefaultEtcdConfig clientv3.Config //etcd客户端配置

type Config struct {
	Mysql   *MySQL              `yaml:"mysql"`
	Etcd    *Etcd               `yaml:"etcd"`
	Service map[string]*Service `yaml:"services"`
	Domain  map[string]*Domain  `yaml:"domain"`
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
	viper.SetConfigName("config") //文件名
	viper.SetConfigType("yaml")   //格式
	viper.AddConfigPath(path.Join(rootDir, "config"))

	//防止Conf空指针
	if err := viper.Unmarshal(&Conf); err != nil {
		panic(err)
	}
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	InitClientV3Config()

}
func InitClientV3Config() {
	DefaultEtcdConfig = clientv3.Config{
		Endpoints:   Conf.Etcd.Address,
		DialTimeout: time.Second * 5, //超时时间为5s
	}

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
