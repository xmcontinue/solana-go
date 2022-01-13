package watcher

import (
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/pkg/worker/xcron"

	"git.cplus.link/crema/backend/chain/sol"
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

	// 每组key创建一个定时任务
	keys := sol.PublicKeys()
	for _, v := range keys {
		tvl := sol.NewTVL(v)
		_, err = cron.AddFunc("0 */10 * * * *", tvl.Start)
		if err != nil {
			panic(err)
		}
	}

	cron.Start()

	return nil
}
