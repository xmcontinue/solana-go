package main

import (
	"os"
	"os/signal"
	"syscall"

	"git.cplus.link/go/akit/config"

	"git.cplus.link/crema/backend/common/chain/sol"
	model "git.cplus.link/crema/backend/common/internal/model/tool"
	"git.cplus.link/crema/backend/common/internal/worker/wacher"
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
	if err := wacher.Init(configer); err != nil {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
