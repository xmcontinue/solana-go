package iface

import (
	"context"

	"git.cplus.link/crema/backend/pkg/domain"
)

const ExchangeServiceName = "CremaExchangeService"

type ExchangeService interface {
	GetPrice(context.Context, *GetPriceReq, *GetPriceResp) error
}

type GetPriceReq struct {
	Market      string `form:"market"`
	BaseSymbol  string `form:"base_symbol"`
	QuoteSymbol string `form:"quote_symbol"`
}

type GetPriceResp struct {
	Time   string          `json:"time"`
	Prices []*domain.Price `json:"prices"`
}
