package handler

import (
	"sync"

	redisV8 "git.cplus.link/go/akit/client/redis/v8"
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/types"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-playground/validator/v10"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/internal/etcd"
	"git.cplus.link/crema/backend/internal/worker/exporter"
	"git.cplus.link/crema/backend/pkg/prometheus"

	"git.cplus.link/crema/backend/pkg/iface"
)

type ExporterService struct {
	conf        *config.Config
	redisClient *redisV8.Client
}

var (
	instance         *ExporterService
	once             sync.Once
	defaultValidator *validator.Validate
)

func NewExporterService(conf *config.Config) (iface.ExporterService, error) {
	var rErr error
	once.Do(func() {
		instance = &ExporterService{
			conf: conf,
		}

		// etcd初始化
		if rErr = etcd.Init(conf); rErr != nil {
			return
		}

		// 验证器初始化
		defaultValidator = validator.New()
		defaultValidator.SetTagName("binding")
		defaultValidator.RegisterCustomTypeFunc(types.ValidateDecimalFunc, decimal.Decimal{})

		// prometheus初始化
		if rErr = prometheus.Init(conf); rErr != nil {
			return
		}

		// sol 初始化
		if err := sol.Init(conf, true); err != nil {
			panic(err)
		}

		// cron初始化
		if rErr = exporter.Init(conf); rErr != nil {
			return
		}

	})
	return instance, rErr
}

func validate(req interface{}) error {
	if err := defaultValidator.Struct(req); err != nil {
		return err
	}
	return nil
}
