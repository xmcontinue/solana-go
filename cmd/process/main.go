package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"git.cplus.link/go/akit/config"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/internal/etcd"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"

	"git.cplus.link/crema/backend/internal/worker/process"
)

func main() {
	go func() {
		// 启动一个 http server，注意 pprof 相关的 handler 已经自动注册过了
		if err := http.ListenAndServe(":6061", nil); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	configer := config.NewConfiger()

	domain.SetPublicPrefix(configer.Get("namespace").(string))
	domain.SetApiHost(configer.Get("api_host").(string))

	// etcd初始化
	if err := etcd.Init(configer); err != nil {
		panic(err)
	}

	// 数据库初始化
	if err := model.Init(configer); err != nil {
		panic(err)
	}

	// sol 初始化
	if err := sol.Init(configer); err != nil {
		panic(err)
	}

	// cron初始化
	if err := process.Init(configer); err != nil {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
