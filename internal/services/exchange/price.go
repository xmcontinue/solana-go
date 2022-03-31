package services

import (
	"context"
	"strings"

	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/transport/rpcx"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/internal/exchanger"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/errcode"
	"git.cplus.link/crema/backend/pkg/iface"
)

// GetPrice ...
func (es *ExchangeService) GetPrice(ctx context.Context, args *iface.GetPriceReq, reply *iface.GetPriceResp) error {
	defer rpcx.Recover(ctx)
	var (
		err error
	)
	if args.BaseSymbol == "" {
		var prices *exchanger.Prices
		if args.Market == "" {
			prices, err = es.exchangerC.GetPricesForAvg(args.QuoteSymbol)
		} else {
			prices, err = es.exchangerC.GetPricesForMarket(args.Market, args.QuoteSymbol)
		}
		if err != nil {
			logger.Error("get price failed", logger.Errorv(err))
			return errcode.GetPriceFailed
		}
		reply.Prices = *prices
	} else {
		var price decimal.Decimal

		if args.Market == "" {
			price, err = es.exchangerC.GetPriceForAvgForShotPath(args.BaseSymbol, args.QuoteSymbol)
		} else {
			price, err = es.exchangerC.GetPriceForMarketForShotPath(args.Market, args.BaseSymbol, args.QuoteSymbol)
		}

		if err != nil {
			logger.Error("get price failed", logger.Errorv(err))
			return errcode.GetPriceFailed
		}

		reply.Prices = []*domain.Price{
			{
				strings.ToUpper(args.BaseSymbol),
				strings.ToUpper(args.QuoteSymbol),
				price,
			},
		}
	}

	return nil
}
