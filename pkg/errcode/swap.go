package errcode

import (
	"git.cplus.link/go/akit/errors"
)

var (
	GetPriceFailed = errors.NewRestError(400, "GET_PRICE_FAILED", "Failed to get price") // 获取价格失败
)
