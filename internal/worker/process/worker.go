package process

import (
	redisV8 "git.cplus.link/go/akit/client/redis/v8"
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/pkg/worker/xcron"
	"git.cplus.link/go/akit/pkg/xlog"
)

var (
	cronConf           *xcron.Config
	cron               *xcron.Cron
	redisClient        *redisV8.Client
	conf               *config.Config
	contractAccountMap = make(map[string]bool)
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
	err = conf.UnmarshalKey("cron_job_conf", &cronConf)
	if err != nil {
		return errors.Wrap(err)
	}

	contractAccounts := make([]string, 0, 2)
	err = conf.UnmarshalKey("contract_account", &contractAccounts)
	if err != nil {
		return errors.Wrap(err)
	}

	for _, v := range contractAccounts {
		contractAccountMap[v] = true
	}

	cronConf.WithLogger(xlog.Config{}.Build())
	cron = cronConf.Build()

	_, err = cron.AddFunc(getSpec("sync_swap_cache"), transactionIDCache)
	if err != nil {
		panic(err)
	}

	_, err = cron.AddFunc(getSpec("sync_swap_cache"), parserUserCountAndSwapCount)
	if err != nil {
		panic(err)
	}

	//_, err = cron.AddFunc(getSpec("sync_swap_cache"), syncTORedis)
	//if err != nil {
	//	panic(err)
	//}

	_, err = cron.AddFunc(getSpec("sync_swap_cache"), syncVolAndTvlHistogram)
	if err != nil {
		panic(err)
	}

	//_, err = cron.AddFunc(getSpec("sync_swap_cache"), syncKLineToRedis)
	//if err != nil {
	//	panic(err)
	//}

	_, err = cron.AddFunc(getSpec("sync_swap_cache"), swapAddressLast24HVol)
	if err != nil {
		panic(err)
	}

	_, err = cron.AddFunc(getSpec("sync_swap_cache"), sumTotalSwapAccount)
	if err != nil {
		panic(err)
	}

	_, err = cron.AddFunc(getSpec("sync_swap_cache"), SwapTotalCount)
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
