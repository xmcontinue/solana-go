package watcher

import (
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/pkg/worker/xcron"
)

var (
	cronConf *xcron.Config
	cron     *xcron.Cron
)

// Init 定时任务
func Init(conf *config.Config) error {
	err := conf.UnmarshalKey("cron", &cronConf)
	if err != nil {
		return errors.Wrap(err)
	}

	cron = cronConf.Build()

	_, err = cron.AddFunc("*/10 * * * * *", SwapCountCacheJob)
	if err != nil {
		panic(err)
	}

	cron.Start()

	return nil
}
