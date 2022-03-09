package market

import (
	"context"
	"encoding/json"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/pkg/worker/xcron"
	"git.cplus.link/go/akit/pkg/xlog"

	"git.cplus.link/crema/backend/internal/etcd"
	"git.cplus.link/crema/backend/pkg/domain"
)

var (
	cronConf       *xcron.Config
	cron           *xcron.Cron
	swapConfigList []domain.SwapConfig
	conf           *config.Config
)

// Init 定时任务
func Init(viperConf *config.Config) error {
	conf = viperConf
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
	err = conf.UnmarshalKey("cron_job_conf", &cronConf)
	if err != nil {
		return errors.Wrap(err)
	}
	cronConf.WithLogger(xlog.Config{}.Build())
	cron = cronConf.Build()

	_, err = cron.AddFunc(getSpec("swap_count_cache"), SwapCountCacheJob)
	if err != nil {
		panic(err)
	}

	_, err = cron.AddFunc(getSpec("tvl_cache"), TvlCacheJob)
	if err != nil {
		panic(err)
	}

	_, err = cron.AddFunc(getSpec("sync_config"), SyncConfigJob)
	if err != nil {
		panic(err)
	}

	cron.Start()

	return nil
}

func getSpec(key string) string {
	return conf.Get("cron_job_interval." + key).(string)
}
