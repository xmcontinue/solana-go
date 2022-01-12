package app

import (
	"context"
	"fmt"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/http"
	"github.com/gin-gonic/gin"

	"git.cplus.link/crema/backend/integrate/cmd/gateway/auth"
)

// Module 模块
type Module interface {
	Start(c *config.Config, engine *gin.Engine) error
	Name() string
}

var (
	configer *config.Config
	jAuth    auth.Authenticator
	modules  []Module
)

func registerModule(m Module) {
	modules = append(modules, m)
}

// Start 启动服务
func Start(c *config.Config) error {
	conf, err := c.Service("Gateway")
	if err != nil {
		return errors.Wrap(err)
	}
	configer = c

	// 普通认证
	jAuth, err = auth.NewJWTAuth(context.Background(), configer)
	if err != nil {
		return errors.Wrap(err)
	}

	engine := http.NewEngine()

	for _, m := range modules {
		if c.GetBool(fmt.Sprintf("gateway.modules.%s", m.Name())) {
			if err := m.Start(c, engine); err != nil {
				return errors.Wrapf(err, "start module:%s", m.Name())
			}
		}
	}

	if err := engine.Run(conf.ListenAddress); err != nil {
		return errors.Wrap(err)
	}

	return nil
}
