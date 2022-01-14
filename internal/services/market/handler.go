package handler

import (
	"sync"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/types"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-playground/validator/v10"

	"git.cplus.link/crema/backend/internal/etcd"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/internal/worker/market"

	"git.cplus.link/crema/backend/pkg/iface"
)

type MarketService struct {
	conf *config.Config
}

var (
	instance         *MarketService
	once             sync.Once
	defaultValidator *validator.Validate
)

const (
	// MaxLimit 列表查询MaxLimit
	MaxLimit = 100

	// DefaultLimit  列表查询默认Limit
	DefaultLimit = 10
)

func NewMarketService(conf *config.Config) (iface.MarketService, error) {
	var rErr error
	once.Do(func() {
		instance = &MarketService{
			conf: conf,
		}

		// etcd初始化
		if err := etcd.Init(conf); err != nil {
			panic(err)
		}

		// 数据库初始化
		if err := model.Init(conf); err != nil {
			rErr = errors.Wrap(err)
			return
		}

		// 验证器初始化
		defaultValidator = validator.New()
		defaultValidator.SetTagName("binding")
		defaultValidator.RegisterCustomTypeFunc(types.ValidateDecimalFunc, decimal.Decimal{})

		// cron初始化
		if err := market.Init(conf); err != nil {
			panic(err)
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

func limit(limit int) int {
	if limit == 0 {
		return DefaultLimit
	}
	if limit >= MaxLimit {
		return MaxLimit
	}
	return limit
}
