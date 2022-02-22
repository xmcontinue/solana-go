package process

import (
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/pkg/worker/xcron"

	redisV8 "git.cplus.link/go/akit/client/redis/v8"
)

var (
	cronConf    *xcron.Config
	cron        *xcron.Cron
	redisClient *redisV8.Client
	conf        *config.Config
)

// Init 定时任务
func Init(viperConf *config.Config) error {
	conf = viperConf
	var err error
	redisClient, err = initRedis(conf)
	if err != nil {
		return errors.Wrap(err)
	}

	// cron init
	err = conf.UnmarshalKey("cron", &cronConf)
	if err != nil {
		return errors.Wrap(err)
	}
	cron = cronConf.Build()

	_, err = cron.AddFunc(getSpec("sync_swap_cache"), syncData)
	if err != nil {
		panic(err)
	}

	_, err = cron.AddFunc(getSpec("sync_swap_cache"), swapAddressLast24HVol)
	if err != nil {
		panic(err)
	}
	_, err = cron.AddFunc(getSpec("sync_swap_cache"), userAddressLast24hVol)
	if err != nil {
		panic(err)
	}

	cron.Start()

	return nil
}

func getSpec(key string) string {
	return conf.Get("cron_spec." + key).(string)
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
