package services

import (
	"context"
	"encoding/json"
	"strings"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/internal/exchanger"
	"git.cplus.link/crema/backend/internal/market/crema"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/errcode"
	"git.cplus.link/crema/backend/pkg/iface"
)

var dataCache = map[string]iface.GetPriceResp{}

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
		*reply = data
		return nil
	}

	// 这里更改逻辑默认获取crema平台价格，若传avg则获取加权平均价
	if args.Market == "" {
		args.Market = crema.BusinessName
	}
	if args.Market == "avg" {
		args.Market = ""
	}

	if args.BaseSymbol == "" {
		if args.Market == "" {
			if args.QuoteSymbol == "" {
				prices = es.exchangerC.GetAllPricesForAvg()
			} else {
				prices, err = es.exchangerC.GetPricesForAvg(args.QuoteSymbol)
			}
		} else {
			if args.QuoteSymbol == "" {
				prices, err = es.exchangerC.GetAllPricesForMarket(args.Market)
			} else {
				prices, err = es.exchangerC.GetPricesForMarket(args.Market, args.QuoteSymbol)
			}
		}
		if err != nil {
			logger.Error("get price failed", logger.Errorv(err))
			return errcode.GetPriceFailed
		}
		reply.Prices = *prices
	} else {
		var price domain.Price

		if args.QuoteSymbol == "" {
			args.QuoteSymbol = "usd"
		}

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
	reply.Time = es.exchangerC.GetSyncTime().Format("2006-01-02 15:04:05")

	setCache(string(cacheKey), *reply)

	return nil
}

func getCache(key string) (iface.GetPriceResp, bool) {
	res, ok := dataCache[key]
	return res, ok
}

func setCache(key string, val iface.GetPriceResp) {
	dataCache[key] = val
}

func cleanCache() {
	dataCache = map[string]iface.GetPriceResp{}
}
