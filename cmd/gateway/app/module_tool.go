package app

import (
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"github.com/gin-gonic/gin"

	handler "git.cplus.link/crema/backend/cmd/gateway/modules/market"
)

type cremaMarket struct{}

func (m *cremaMarket) Name() string {
	return "market"
}

func init() {
	registerModule(&cremaMarket{})
}

func (m *cremaMarket) Start(c *config.Config, engine *gin.Engine) error {
	if err := handler.Init(c); err != nil {
		return errors.Wrap(err)
	}

	engine.GET("/swap/count", handler.SwapCount)
	engine.GET("/config", handler.GetConfig)
	engine.GET("/tvl/24hour", handler.GetTvl)

	return nil
}
