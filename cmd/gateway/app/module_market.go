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
	engine.GET("activity/nft/:Mint", handler.GetActivityNftMetadata)
	engine.GET("activity/history/:User", handler.GetActivityHistoryByUser)

	engine.GET("/v1/swap/count", handler.SwapCount)
	engine.GET("/v2/swap/count", handler.SwapCountSharding)
	engine.GET("/v1/swap/count/new", handler.SwapCount)
	engine.GET("/v1/swap/count/list", handler.SwapCountList)
	engine.GET("/v1/swap/count/kline", handler.QuerySwapKline)
	engine.POST("/v1/tvl", handler.GetTvlV2)
	engine.POST("/v1/vol/24h", handler.Get24hVolV2)
	engine.POST("/v1/vol", handler.GetVolV2)
	engine.POST("/v1/kline", handler.GetKline)
	engine.GET("/v1/histogram", handler.GetHistogram)
	engine.GET("/v2/histogram", handler.GetHistogramSharding)
	engine.GET("/v1/token/tvl", handler.TvlOfSingleToken)
	engine.GET("/v1/transaction", handler.GetTransaction)
	engine.GET("/v1/position", handler.GetPosition)

	engine.POST("/gallery", handler.GetGallery)
	engine.GET("/gallery/type", handler.GetGalleryType)

	return nil
}
