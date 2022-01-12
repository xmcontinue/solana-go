package main

import (
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/transport/rpcx"

	handler "git.cplus.link/crema/backend/common/internal/services/tool"

	"git.cplus.link/crema/backend/common/pkg/iface"
)

func main() {
	configer := config.NewConfiger()

	serviceConf, err := configer.Service(iface.ToolServiceName)
	if err != nil {
		panic(err)
	}

	service, err := handler.NewToolService(configer)
	if err != nil {
		panic(err)
	}

	logger.Info("service start", logger.String("name", serviceConf.Name), logger.Any("config", serviceConf))
	if err = rpcx.DefaultServe(serviceConf, service); err != nil {
		panic(err)
	}
}
