package services

import (
	"context"
	"encoding/json"
	"strings"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/internal/exchanger"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/errcode"
	"git.cplus.link/crema/backend/pkg/iface"
)

var dataCache = map[string][]*domain.Price{}

// GetPrice ...
func (es *ExchangeService) GetPrice(ctx context.Context, args *iface.GetPriceReq, reply *iface.GetPriceResp) error {
	defer rpcx.Recover(ctx)
	var (
		err    error
		prices *exchanger.Prices
	)
	cacheKey, err := json.Marshal(args)
	if err != nil {
		if err != nil {
			return errors.Wrap(errcode.GetPriceFailed)
		}
	}

	if data, ok := getCache(string(cacheKey)); ok {
		reply.Prices = data
		return nil
	}

	if args.BaseSymbol == "" {
		if args.Market == "" {
			if args.QuoteSymbol == "" {
				prices = es.exchangerC.GetAllPricesForAvg()
			} else {
				prices, err = es.exchangerC.GetPricesForAvg(args.QuoteSymbol)
			}
		} else {
			prices, err = es.exchangerC.GetPricesForMarket(args.Market, args.QuoteSymbol)
		}
		if err != nil {
			logger.Error("get price failed", logger.Errorv(err))
			return errcode.GetPriceFailed
		}
		reply.Prices = *prices
	} else {
		var price domain.Price

		if args.Market == "" {
			price.Price, err = es.exchangerC.GetPriceForAvgForShotPath(args.BaseSymbol, args.QuoteSymbol)
		} else {
			price.Price, err = es.exchangerC.GetPriceForMarketForShotPath(args.Market, args.BaseSymbol, args.QuoteSymbol)
		}

		if err != nil {
			logger.Error("get price failed", logger.Errorv(err))
			return errcode.GetPriceFailed
		}

		price.BaseSymbol = strings.ToUpper(args.BaseSymbol)
		price.QuoteSymbol = strings.ToUpper(args.QuoteSymbol)
		prices = (*exchanger.Prices)(&[]*domain.Price{&price})
	}

	reply.Prices = *prices
	setCache(string(cacheKey), *prices)

	return nil
}

func getCache(key string) ([]*domain.Price, bool) {
	prices, ok := dataCache[key]
	return prices, ok
}

func setCache(key string, val []*domain.Price) {
	dataCache[key] = val
}

func cleanCache() {
	dataCache = map[string][]*domain.Price{}
}
