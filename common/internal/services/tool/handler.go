package handler

import (
	"sync"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/types"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-playground/validator/v10"

	"git.cplus.link/crema/backend/common/chain/sol"
	model "git.cplus.link/crema/backend/common/internal/model/tool"
	wacher "git.cplus.link/crema/backend/common/internal/worker/tool"
	"git.cplus.link/crema/backend/common/pkg/iface"
)

type ToolService struct {
	conf *config.Config
}

var (
	instance         *ToolService
	once             sync.Once
	defaultValidator *validator.Validate
)

const (
	// MaxLimit 列表查询MaxLimit
	MaxLimit = 100

	// DefaultLimit  列表查询默认Limit
	DefaultLimit = 10
)

func NewToolService(conf *config.Config) (iface.ToolService, error) {
	var rErr error
	once.Do(func() {
		instance = &ToolService{
			conf: conf,
		}

		// 数据库初始化
		if err := model.Init(conf); err != nil {
			rErr = errors.Wrap(err)
			return
		}

		// sol 初始化
		if err := sol.Init(conf); err != nil {
			rErr = errors.Wrap(err)
			return
		}

		// 验证器初始化
		defaultValidator = validator.New()
		defaultValidator.SetTagName("binding")
		defaultValidator.RegisterCustomTypeFunc(types.ValidateDecimalFunc, decimal.Decimal{})

		// cron初始化
		if err := wacher.Init(conf); err != nil {
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
