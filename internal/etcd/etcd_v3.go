package etcd

import (
	"sync"

	"git.cplus.link/go/akit/client/etcdv3"
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
)

var (
	onceV3       sync.Once
	etcdV3Config *etcdv3.Config
	etcdV3Client *etcdv3.Client
)

// InitV3 etcd初始化
func InitV3(conf *config.Config) error {
	var rErr error
	onceV3.Do(func() {
		host := conf.Get("config_center.host")
		if host == nil {
			rErr = errors.New("etcd host not found")
			return
		}

		etcdV3Config = etcdv3.DefaultConfig()
		etcdV3Config.Endpoints = []string{host.(string)}

		etcdV3Client = etcdV3Config.Build()

	})

	return rErr
}

func ClientV3() *etcdv3.Client {
	return etcdV3Client
}

func ConfigV3() *etcdv3.Config {
	return etcdV3Config
}
