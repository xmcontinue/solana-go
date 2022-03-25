package services

import (
	"context"
	"strings"

	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/errcode"
	"git.cplus.link/crema/backend/pkg/iface"
)

// GetPrice ...
func (es *ExchangeService) GetPrice(ctx context.Context, args *iface.GetPriceReq, reply *iface.GetPriceResp) error {
	defer rpcx.Recover(ctx)

	if args.BaseSymbol == "" {
		prices, err := es.exchangerC.GetPricesForMarket("coingecko", args.QuoteSymbol)
		reply.Prices = *prices
		if err != nil {
			logger.Error("get price failed", logger.Errorv(err))
			return errcode.GetPriceFailed
		}
	} else {
		// price, err := es.exchangerC.GetPriceForMarket("coingecko", args.BaseSymbol, args.QuoteSymbol)
		// if err != nil {
		// 	logger.Error("get price failed", logger.Errorv(err))
		// 	return errcode.GetPriceFailed
		// }
		price, err := es.exchangerC.GetPriceForMarketForShotPath("coingecko", args.BaseSymbol, args.QuoteSymbol)
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
