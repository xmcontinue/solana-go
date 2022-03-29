package main

import (
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/go/akit/config"

	handler "git.cplus.link/crema/backend/internal/services/exchange"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/iface"
)

func main() {
	configer := config.NewConfiger()

	domain.SetPublicPrefix(configer.Get("namespace").(string))

	serviceConf, err := configer.Service(iface.ExchangeServiceName)
	if err != nil {
		panic(err)
	}

	service, err := handler.NewExchangeService(configer)
	if err != nil {
		panic(err)
	}

	logger.Info("service start", logger.String("name", serviceConf.Name), logger.Any("config", serviceConf))
	if err = rpcx.DefaultServe(serviceConf, service); err != nil {
		panic(err)
	}
}
