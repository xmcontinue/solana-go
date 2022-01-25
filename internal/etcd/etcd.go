package etcd

import (
	"sync"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"go.etcd.io/etcd/client/v2"
)

var (
	once       sync.Once
	etcdConf   client.Config
	etcdClient client.Client
)

// Init etcd初始化
func Init(conf *config.Config) error {
	var rErr error
	once.Do(func() {
		var err error
		host := conf.Get("config_center.host")
		if host == nil {
			rErr = errors.New("etcd host not found")
			return
		}
		etcdConf.Endpoints = []string{host.(string)}

		etcdClient, err = client.New(etcdConf)
		if err != nil {
			rErr = errors.Wrapf(err, "init etcd")
			return
		}
	})

	return rErr
}

func Client() client.Client {
	return etcdClient
}

func Api() client.KeysAPI {
	return client.NewKeysAPI(etcdClient)
}
