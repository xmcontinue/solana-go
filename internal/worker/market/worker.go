package market

import (
	"context"
	"encoding/json"

	redisV8 "git.cplus.link/go/akit/client/redis/v8"
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
	redisClient    *redisV8.Client
)

const (
	defaultBaseSpec = "*/10 * * * * *"
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

	redisClient, err = initRedis(conf)
	if err != nil {
		return err
	}
	syncGalleryCache()
	// cron init
	err = conf.UnmarshalKey("cron_job_conf", &cronConf)
	if err != nil {
		return errors.Wrap(err)
	}

	cronConf.WithLogger(xlog.Config{}.Build())
	cronConf.Config = etcd.ConfigV3()
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
	_, err = cron.AddFunc(defaultBaseSpec, syncGalleryCache)
	if err != nil {
		panic(err)
	}

	cron.Start()

	return nil
}

func getSpec(key string) string {
	return conf.Get("cron_job_interval." + key).(string)
}

// initRedis 初始化redis
func initRedis(conf *config.Config) (*redisV8.Client, error) {
	c := redisV8.DefaultRedisConfig()
	err := conf.UnmarshalKey("redis", c)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return redisV8.NewClient(c)
}
