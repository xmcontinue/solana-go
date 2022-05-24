package app

import (
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"github.com/gin-gonic/gin"

	handler "git.cplus.link/crema/backend/cmd/gateway/modules/exporter"
)

type cremaExproter struct{}

func (m *cremaExproter) Name() string {
	return "exchange"
}

func init() {
	registerModule(&cremaExproter{})
}

func (m *cremaExproter) Start(c *config.Config, engine *gin.Engine) error {
	if err := handler.Init(c); err != nil {
		return errors.Wrap(err)
	}
	engine.POST("/log", handler.AddLog)

	return nil
}
