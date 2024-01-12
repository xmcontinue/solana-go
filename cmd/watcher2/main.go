package main

import (
	"os"
	"os/signal"
	"syscall"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/logger"

	"git.cplus.link/crema/backend/internal/etcd"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"

	"git.cplus.link/crema/backend/internal/worker/watcher"

	"git.cplus.link/crema/backend/chain/sol"
)

func main() {
	configer := config.NewConfiger()

	domain.SetPublicPrefix(configer.Get("namespace").(string))
	domain.SetApiHost(configer.Get("api_host").(string))

	// etcd初始化
	if err := etcd.Init(configer); err != nil {
		panic(err)
	}
	if err := etcd.InitV3(configer); err != nil {
		panic(err)
	}

	// 数据库初始化
	if err := model.Init(configer); err != nil {
		panic(err)
	}

	// sol初始化
	if err := sol.Init(configer, false); err != nil {
		panic(err)
	}

	// 添加支持分表
	configs := sol.SwapConfigList()
	shardingValues := make([]string, 0, len(configs))
	for _, v := range configs {
		shardingValues = append(shardingValues, v.SwapAccount)
	}
	if err := model.InitWithSharding(shardingValues); err != nil {
		panic(err)
	}
	logger.Info("sharding complete!")

	// coinGecko初始化
	// coingecko.Init()

	// cron初始化
	if err := watcher.Init(configer); err != nil {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}