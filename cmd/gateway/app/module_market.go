package app

import (
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"github.com/gin-gonic/gin"

	handler "git.cplus.link/crema/backend/cmd/gateway/modules/exchange"
)

type cremaExchange struct{}

func (m *cremaExchange) Name() string {
	return "exchange"
}

func init() {
	registerModule(&cremaExchange{})
}

func (m *cremaExchange) Start(c *config.Config, engine *gin.Engine) error {
	if err := handler.Init(c); err != nil {
		return errors.Wrap(err)
	}
	engine.GET("/price", handler.GetPrices)

	return nil
}
