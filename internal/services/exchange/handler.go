package services

import (
	"sync"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/types"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-playground/validator/v10"

	"git.cplus.link/crema/backend/internal/worker/exchange"

	eConfig "git.cplus.link/crema/backend/internal/config"
	"git.cplus.link/crema/backend/internal/etcd"
	"git.cplus.link/crema/backend/internal/exchanger"
	"git.cplus.link/crema/backend/pkg/iface"
)

type ExchangeService struct {
	conf       *config.Config
	exchangerC *exchanger.Exchanger
}

var (
	instance         *ExchangeService
	once             sync.Once
	defaultValidator *validator.Validate
)

const (
	// MaxLimit 列表查询MaxLimit
	MaxLimit = 100

	// DefaultLimit  列表查询默认Limit
	DefaultLimit = 10

	// 柱状图默认返回300
	histogramDefaultLen = 300

	// 最大值
	histogramMaxLen = 500
)

func NewExchangeService(conf *config.Config) (iface.ExchangeService, error) {
	var rErr error
	once.Do(func() {
		instance = &ExchangeService{
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

		// 汇率中心初始化
		exchangeConf, rErr := eConfig.NewExchangeConfigForViper(conf)
		if rErr != nil {
			return
		}
		instance.exchangerC = exchanger.NewExchanger()
		rErr = instance.exchangerC.LoadConfig(exchangeConf)
		if rErr != nil {
			return
		}
		// cron初始化
		if rErr = worker.Init(conf, instance.exchangerC); rErr != nil {
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

func limit(limit int) int {
	if limit == 0 {
		return DefaultLimit
	}
	if limit >= MaxLimit {
		return MaxLimit
	}
	return limit
}

func histogramLimit(histogramLimit int) int {
	if histogramLimit == 0 {
		return histogramDefaultLen
	}
	if histogramLimit >= histogramMaxLen {
		return histogramMaxLen
	}
	return histogramLimit
}
