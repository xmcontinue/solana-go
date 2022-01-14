package etcd

import (
	"sync"

	"git.cplus.link/go/akit/client/etcdv3"
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
)

var (
	once   sync.Once
	client *etcdv3.Client
)

// Init etcd初始化
func Init(conf *config.Config) error {
	var rErr error
	once.Do(func() {
		etcdConf := etcdv3.DefaultConfig()
		err := conf.UnmarshalKey("etcds", &etcdConf.Endpoints)
		if err != nil {
			rErr = errors.Wrapf(err, "init etcd")
			return
		}
		client = etcdConf.Build()
	})

	return rErr
}

func Client() *etcdv3.Client {
	return client
}
