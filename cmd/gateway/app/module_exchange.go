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

	engine.GET("/config", handler.GetConfig)
	engine.GET("/token/config", handler.GetTokenConfig)
	engine.GET("/tvl/24hour", handler.GetTvl)

	engine.GET("/v1/swap/count", handler.SwapCount)
	engine.POST("/v1/tvl", handler.GetTvlV2)
	engine.POST("/v1/vol/24h", handler.Get24hVolV2)
	engine.POST("/v1/vol", handler.GetVolV2)
	engine.POST("/v1/kline", handler.GetKline)
	engine.GET("/v1/histogram", handler.GetHistogram)
	engine.GET("/v1/token/tvl", handler.TvlOfSingleToken)

	return nil
}