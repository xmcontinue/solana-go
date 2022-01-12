package app

import (
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"github.com/gin-gonic/gin"

	handler "git.cplus.link/crema/backend/integrate/cmd/gateway/modules/tool"
)

type cremaTool struct{}

func (m *cremaTool) Name() string {
	return "tool"
}

func init() {
	registerModule(&cremaTool{})
}

func (m *cremaTool) Start(c *config.Config, engine *gin.Engine) error {
	if err := handler.Init(c); err != nil {
		return errors.Wrap(err)
	}

	engine.POST("/swap/count", handler.SwapCount) // NoAuth()

	return nil
}
