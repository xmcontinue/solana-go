package market

import (
	"context"
	"encoding/json"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/pkg/worker/xcron"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/internal/etcd"
)

var (
	cronConf       *xcron.Config
	cron           *xcron.Cron
	swapConfigList []sol.SwapConfig
)

// Init 定时任务
func Init(conf *config.Config) error {

	// 地址init
	confVal, err := etcd.Api().Get(context.TODO(), "/crema/swap-pairs", nil)
	if err != nil || confVal == nil {
		return errors.Wrap(err)
	}
	err = json.Unmarshal([]byte(confVal.Node.Value), &swapConfigList)
	if err != nil {
		return errors.Wrap(err)
	}
	
	// cron init
	err = conf.UnmarshalKey("cron", &cronConf)
	if err != nil {
		return errors.Wrap(err)
	}
	cron = cronConf.Build()

	_, err = cron.AddFunc("*/10 * * * * *", SwapCountCacheJob)
	if err != nil {
		panic(err)
	}

	_, err = cron.AddFunc("*/10 * * * * *", TvlCacheJob)
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
