package main

import (
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/transport/rpcx"

	handler "git.cplus.link/crema/backend/internal/services/market"
	"git.cplus.link/crema/backend/pkg/domain"

	"git.cplus.link/crema/backend/pkg/iface"
)

func main() {
	configer := config.NewConfiger()

	domain.SetPublicPrefix(configer.Get("namespace").(string))
	domain.SetApiHost(configer.Get("api_host").(string))

	serviceConf, err := configer.Service(iface.MarketServiceName)
	if err != nil {
		panic(err)
	}

	service, err := handler.NewMarketService(configer)
	if err != nil {
		panic(err)
	}

	logger.Info("service start", logger.String("name", serviceConf.Name), logger.Any("config", serviceConf))
	if err = rpcx.DefaultServe(serviceConf, service); err != nil {
		panic(err)
	}
}
