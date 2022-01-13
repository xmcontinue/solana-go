package market

import (
	"git.cplus.link/go/akit/client/etcdv3"
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/pkg/worker/xcron"

	"git.cplus.link/crema/backend/common/chain/sol"
)

var (
	cronConf   *xcron.Config
	cron       *xcron.Cron
	addresses  []sol.Address
	etcdClient *etcdv3.Client
)

// Init 定时任务
func Init(conf *config.Config) error {
	etcdConf := etcdv3.DefaultConfig()
	err := conf.UnmarshalKey("etcds", &etcdConf.Endpoints)
	if err != nil {
		return errors.Wrap(err)
	}
	etcdClient = etcdConf.Build()

	err = conf.UnmarshalKey("cron", &cronConf)
	if err != nil {
		return errors.Wrap(err)
	}

	err = conf.UnmarshalKey("tvl_address", &addresses)
	if err != nil {
		return errors.Wrap(err)
	}

	cron = cronConf.Build()

	_, err = cron.AddFunc("*/10 * * * * *", SwapCountCacheJob)
	if err != nil {
		panic(err)
	}

	_, err = cron.AddFunc("*/60 * * * * *", SyncConfigJob)
	if err != nil {
		panic(err)
	}

	cron.Start()

	return nil
}
