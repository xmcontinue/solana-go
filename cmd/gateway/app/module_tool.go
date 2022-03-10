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
	engine.POST("/v2/tvl", handler.GetTvlV2)
	engine.POST("/v2/vol/24h", handler.Get24hVolV2)
	engine.POST("/v2/vol", handler.GetVolV2)
	engine.POST("/v2/kline", handler.GetKline)
	engine.POST("/v2/histogram", handler.GetHistogram)

	return nil
}
