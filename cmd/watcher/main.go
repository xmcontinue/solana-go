package main

import (
	"os"
	"os/signal"
	"syscall"

	"git.cplus.link/go/akit/config"

	model "git.cplus.link/crema/backend/internal/model/market"

	"git.cplus.link/crema/backend/internal/worker/watcher"

	"git.cplus.link/crema/backend/chain/sol"
)

func main() {
	configer := config.NewConfiger()

	// 数据库初始化
	if err := model.Init(configer); err != nil {
		panic(err)
	}

	// sol 初始化
	if err := sol.Init(configer); err != nil {
		panic(err)
	}

	// cron初始化
	if err := watcher.Init(configer); err != nil {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
