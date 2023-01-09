package watcher

import (
	"context"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/chain/sol/parse"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/crema"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/kline"
)

type swapPairPrice struct {
	TokenASymbol string
	TokenBSymbol string
	Price        decimal.Decimal
}

type tokenPrice struct {
	Price decimal.Decimal
}

func SyncSwapPrice() error {
	logger.Info("price syncing ......")

	configs := sol.SwapConfigListV2()

	ctx, swapPairPrices := context.Background(), make([]*swapPairPrice, 0, len(configs))

	// 同步swap pair price
	for _, config := range configs {
		res, err := sol.GetRpcClient().GetAccountInfo(context.Background(), config.SwapPublicKey)
		if err != nil {
			return errors.Wrap(err)
		}

		// 获取价格
		var swapPrice decimal.Decimal

		if config.Version == "v2" {
			swapPrice = parse.GetSwapPriceV2(res, config)
		} else {
			swapPrice = parse.GetSwapPrice(res, config)
		}

		swapPairPrices = append(swapPairPrices, &swapPairPrice{
			TokenASymbol: config.TokenA.Symbol,
			TokenBSymbol: config.TokenB.Symbol,
			Price:        swapPrice,
		})

		// 处理kline数据
		now := time.Now().UTC()
		Kline := kline.NewKline(&now)

		swapPairPriceKLine := &domain.SwapPairPriceKLine{
			SwapAddress: config.SwapAccount,
			Open:        swapPrice,
			High:        swapPrice,
			Low:         swapPrice,
			Avg:         swapPrice,
			Settle:      swapPrice,
			DateType:    domain.DateMin,
		}

		for _, t := range Kline.Types {
			swapPairPriceKLine.Date = t.Date
			swapPairPriceKLine.DateType = t.DateType

			err = updateSwapPairPrice(ctx, config, t, swapPairPriceKLine)
			if err != nil {
				return errors.Wrap(err)
			}

			// 插入数据
			_, err = model.UpsertSwapPairPriceKLine(ctx, swapPairPriceKLine)
			if err != nil {
				return errors.Wrap(err)
			}

		}
	}

	// 同步 token price
	usdcPrice, err := crema.GetPriceForSymbol("USDC")
	if err != nil {
		return errors.Wrap(err)
	}
	solPrice, err := crema.GetPriceForSymbol("SOL")
	if err != nil {
		return errors.Wrap(err)
	}
	tokenPrices := map[string]*tokenPrice{
		"USDC":  {usdcPrice},
		"SOL":   {solPrice},
		"pUSDC": {decimal.NewFromInt(1)},
		"CUSDC": {decimal.NewFromInt(1)},
		// TODO 新增价格，测试使用
		"tUSDC": {decimal.NewFromInt(1)},
		"aUSDC": {decimal.NewFromInt(1)},
		"xUSDC": {decimal.NewFromInt(1)},
	}

	pairPriceToTokenPrice(swapPairPrices, tokenPrices)

	for k, v := range tokenPrices {
		// 处理kline数据
		now := time.Now().UTC()
		Kline := kline.NewKline(&now)

		swapTokenPriceKLine := &domain.SwapTokenPriceKLine{
			Symbol:   k,
			Open:     v.Price,
			High:     v.Price,
			Low:      v.Price,
			Avg:      v.Price,
			Settle:   v.Price,
			DateType: domain.DateMin,
		}

		for _, t := range Kline.Types {
			swapTokenPriceKLine.Date = t.Date
			swapTokenPriceKLine.DateType = t.DateType

			err := updateSwapTokenPrice(ctx, t, swapTokenPriceKLine)
			if err != nil {
				return errors.Wrap(err)
			}

			// 插入数据
			_, err = model.UpsertSwapTokenPriceKLine(ctx, swapTokenPriceKLine)
			if err != nil {
				return errors.Wrap(err)
			}

		}
	}

	return nil
}

func updateSwapPairPrice(ctx context.Context, config *domain.SwapConfig, t *kline.Type, swapPairPriceKLine *domain.SwapPairPriceKLine) error {
	swapPriceKline, err := model.QuerySwapPairPriceKLine(ctx,
		model.SwapAddressFilter(config.SwapAccount),
		model.NewFilter("date = ?", t.Date),
		model.NewFilter("date_type = ?", t.DateType))

	if err != nil && !errors.Is(err, errors.RecordNotFound) {
		return errors.Wrap(err)
	}

	if swapPriceKline != nil {
		if swapPriceKline.High.GreaterThan(swapPairPriceKLine.High) {
			swapPairPriceKLine.High = swapPriceKline.High
		}
		if swapPriceKline.Low.LessThan(swapPairPriceKLine.Low) {
			swapPairPriceKLine.Low = swapPriceKline.Low
		}
	}

	if t.DateType != domain.DateMin {

		InnerAvg, err := t.CalculateAvg(func(endTime time.Time, avgList *[]*kline.InterTime) error {
			swapCountKLines, err := model.QuerySwapPairPriceKLines(ctx, t.Interval, 0,
				model.SwapAddressFilter(config.SwapAccount),
				model.NewFilter("date_type = ?", t.BeforeIntervalDateType),
				model.NewFilter("date < ?", endTime),
				model.OrderFilter("date desc"),
			)

			if err != nil {
				return errors.Wrap(err)
			}

			for index := range swapCountKLines {
				for _, avg := range *avgList {
					if swapCountKLines[len(swapCountKLines)-1-index].Date.Equal(avg.Date) || swapCountKLines[len(swapCountKLines)-1-index].Date.Before(avg.Date) {
						avg.Avg = swapCountKLines[len(swapCountKLines)-1-index].Avg // 以上一个时间区间的平均值作为新的时间区间的平均值
					}
				}
			}

			return nil
		})
		if err != nil {
			return errors.Wrap(err)
		}

		swapPairPriceKLine.Avg = InnerAvg.Avg

	}

	return nil
}

func updateSwapTokenPrice(ctx context.Context, t *kline.Type, swapPairPriceKLine *domain.SwapTokenPriceKLine) error {
	swapPriceKline, err := model.QuerySwapTokenPriceKLine(ctx,
		model.NewFilter("symbol = ?", swapPairPriceKLine.Symbol),
		model.NewFilter("date_type = ?", t.DateType),
		model.NewFilter("date = ?", t.Date),
	)

	if err != nil && !errors.Is(err, errors.RecordNotFound) {
		return errors.Wrap(err)
	}

	if swapPriceKline != nil {
		if swapPriceKline.High.GreaterThan(swapPairPriceKLine.High) {
			swapPairPriceKLine.High = swapPriceKline.High
		}
		if swapPriceKline.Low.LessThan(swapPairPriceKLine.Low) {
			swapPairPriceKLine.Low = swapPriceKline.Low
		}
	}

	if t.DateType != domain.DateMin {

		InnerAvg, err := t.CalculateAvg(func(endTime time.Time, avgList *[]*kline.InterTime) error {
			swapCountKLines, err := model.QuerySwapTokenPriceKLines(ctx, t.Interval, 0,
				model.NewFilter("date_type = ?", t.BeforeIntervalDateType),
				model.NewFilter("symbol = ?", swapPairPriceKLine.Symbol),
				model.NewFilter("date < ?", endTime),
				model.OrderFilter("date desc"),
			)

			if err != nil {
				return errors.Wrap(err)
			}

			for index := range swapCountKLines {
				for _, avg := range *avgList {
					if swapCountKLines[len(swapCountKLines)-1-index].Date.Equal(avg.Date) || swapCountKLines[len(swapCountKLines)-1-index].Date.Before(avg.Date) {
						avg.Avg = swapCountKLines[len(swapCountKLines)-1-index].Avg // 以上一个时间区间的平均值作为新的时间区间的平均值
					}
				}
			}

			return nil
		})
		if err != nil {
			return errors.Wrap(err)
		}

		swapPairPriceKLine.Avg = InnerAvg.Avg

	}

	return nil
}

func pairPriceToTokenPrice(pairPriceList []*swapPairPrice, tokenPriceList map[string]*tokenPrice) {
	beforeLen := len(tokenPriceList)

	for _, v := range pairPriceList {
		tokenAPrice, tokenAHas := tokenPriceList[v.TokenASymbol]
		tokenBPrice, tokenBHas := tokenPriceList[v.TokenBSymbol]

		if tokenAHas && !tokenBHas {
			tokenPriceList[v.TokenBSymbol] = &tokenPrice{
				Price: tokenAPrice.Price.Div(v.Price).Round(12),
			}
			continue
		}

		if tokenBHas && !tokenAHas {
			tokenPriceList[v.TokenASymbol] = &tokenPrice{
				Price: tokenBPrice.Price.Mul(v.Price).Round(12),
			}
			continue
		}
	}

	if beforeLen == len(tokenPriceList) {
		return
	}

	pairPriceToTokenPrice(pairPriceList, tokenPriceList)
}
