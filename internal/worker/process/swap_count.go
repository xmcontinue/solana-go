package process

import (
	"context"
	"time"

	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

// SwapTotalCount 汇总统计
func SwapTotalCount() error {
	// 将要统计的数据
	swapCountToApi := &domain.SwapCountToApi{
		Pools:  make([]*domain.SwapCountToApiPool, 0),
		Tokens: make([]*domain.SwapCountToApiToken, 0),
	}

	// 获取swap pair 24h 内交易统计
	totalVolInUsd24h, totalVolInUsd, totalTvlInUsd, totalTxNum24h, totalTxNum, date := decimal.Decimal{}, decimal.Decimal{}, decimal.Decimal{}, uint64(0), uint64(0), time.Now().Add(-24*time.Hour)

	for _, v := range sol.SwapConfigList() {
		// 获取币价
		newSwapCount, err := model.QuerySwapCountKLine(context.Background(), model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = 1min"), model.OrderFilter("id desc"))
		if err != nil {
			continue
		}
		beforeSwapCount, err := model.QuerySwapCountKLine(context.Background(), model.NewFilter("date > ?", date), model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = 1min"))
		if err != nil {
			continue
		}
		if newSwapCount.TokenAUSD.IsZero() || newSwapCount.TokenBUSD.IsZero() || beforeSwapCount.TokenAUSD.IsZero() || beforeSwapCount.TokenBUSD.IsZero() {
			continue
		}

		// 获取24h交易额，交易笔数
		swapCount24h, err := model.SumSwapCountVolForKLines(context.Background(), model.NewFilter("date > ?", date), model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = 1min"))
		if err != nil {
			continue
		}

		// 获取总交易额，交易笔数
		swapCountTotal, err := model.SumSwapCountVolForKLines(context.Background(), model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = mon"))
		if err != nil {
			continue
		}

		// 计算vol,tvl
		tokenATvl, tokenBTvl := v.TokenA.Balance.Mul(newSwapCount.TokenAUSD), v.TokenB.Balance.Mul(newSwapCount.TokenBUSD)
		tokenAVol24h, tokenBVol24h := swapCount24h.TokenAVolume.Mul(newSwapCount.TokenAUSD), swapCount24h.TokenBVolume.Mul(newSwapCount.TokenBUSD)
		tokenAVol, tokenBVol := swapCountTotal.TokenAVolume.Mul(newSwapCount.TokenAUSD), swapCountTotal.TokenBVolume.Mul(newSwapCount.TokenBUSD)
		tvlInUsd, volInUsd24h, volInUsd := tokenATvl.Add(tokenBTvl), tokenAVol24h.Add(tokenBVol24h), tokenAVol.Add(tokenBVol)

		// 计算apr
		apr := "0%"
		if !tvlInUsd.IsZero() {
			fee, _ := decimal.NewFromString(v.Fee)
			apr = volInUsd24h.Mul(fee).Mul(decimal.NewFromInt(36500)).Div(tvlInUsd).Round(2).String() + "%" // 36500为365天*百分比转化100得出
		}

		// pool统计
		newSwapPrice, beforeSwapPrice := newSwapCount.TokenBUSD.Div(newSwapCount.TokenAUSD).Round(6), beforeSwapCount.TokenBUSD.Div(beforeSwapCount.TokenAUSD).Round(6)
		swapCountToApiPool := &domain.SwapCountToApiPool{
			Name:          v.Name,
			SwapAccount:   v.SwapAccount,
			VolInUsd24h:   volInUsd24h.String(),
			TxNum24h:      swapCount24h.TxNum,
			VolInUsd:      volInUsd.String(),
			TxNum:         swapCountTotal.TxNum,
			Apr:           apr,
			TvlInUsd:      tvlInUsd.String(),
			TickIntervals: []*domain.TickInterval{}, // TODO
			Price:         newSwapPrice.String(),
			PriceRate24h:  newSwapPrice.Sub(beforeSwapPrice).Div(beforeSwapPrice).Round(2).String() + "%",
		}
		swapCountToApi.Pools = append(swapCountToApi.Pools, swapCountToApiPool)

		// token统计
		appendTokensToSwapCount(
			swapCountToApi,
			&domain.SwapCountToApiToken{
				Name:         v.TokenA.Symbol,
				VolInUsd24h:  tokenAVol24h.String(),
				TxNum24h:     swapCount24h.TxNum,
				VolInUsd:     tokenAVol.String(),
				TxNum:        swapCountTotal.TxNum,
				TvlInUsd:     tokenATvl.String(),
				Price:        newSwapCount.TokenAUSD.String(),
				PriceRate24h: newSwapCount.TokenAUSD.Sub(beforeSwapCount.TokenAUSD).Div(beforeSwapCount.TokenAUSD).Round(2).String() + "%",
			},
			&domain.SwapCountToApiToken{
				Name:         v.TokenB.Symbol,
				VolInUsd24h:  tokenBVol24h.String(),
				TxNum24h:     swapCount24h.TxNum,
				VolInUsd:     tokenBVol.String(),
				TxNum:        swapCountTotal.TxNum,
				TvlInUsd:     tokenBTvl.String(),
				Price:        newSwapCount.TokenBUSD.String(),
				PriceRate24h: newSwapCount.TokenBUSD.Sub(beforeSwapCount.TokenBUSD).Div(beforeSwapCount.TokenBUSD).Round(2).String() + "%",
			},
		)

		// 汇总处理
		totalVolInUsd24h = totalVolInUsd24h.Add(volInUsd24h)
		totalVolInUsd = totalVolInUsd.Add(volInUsd)
		totalTvlInUsd = totalTvlInUsd.Add(tvlInUsd)
		totalTxNum24h = totalTxNum24h + swapCount24h.TxNum
		totalTxNum = totalTxNum + swapCountTotal.TxNum
	}

	// token数量
	swapCountToApi.TokenNum = len(swapCountToApi.Tokens)

	// 用户数量
	swapCountToApi.UserNum, _ = model.CountUserSwapCount(context.Background())

	// 总交易额
	swapCountToApi.VolInUsd24h = totalVolInUsd24h.String()
	swapCountToApi.VolInUsd = totalVolInUsd.String()
	swapCountToApi.TvlInUsd = totalTvlInUsd.String()
	swapCountToApi.TxNum24h = totalTxNum24h
	swapCountToApi.TxNum = totalTxNum

	// 缓存至redis
	// swapCountKey := domain.SwapTotalCountKey()
	// if err := redisClient.Set(context.Background(), swapCountKey.Key, swapCountToApi, swapCountKey.Timeout).Err(); err != nil {
	// 	return errors.Wrap(err)
	// }

	return nil
}

func appendTokensToSwapCount(swapCountToApi *domain.SwapCountToApi, tokens ...*domain.SwapCountToApiToken) {
	for _, t := range tokens {

		has := false

		for _, v := range swapCountToApi.Tokens {

			if v.Name == t.Name {

				volInUsd24hV, _ := decimal.NewFromString(v.VolInUsd24h)
				volInUsd24hT, _ := decimal.NewFromString(t.VolInUsd24h)
				volInUsdV, _ := decimal.NewFromString(v.VolInUsd)
				volInUsdT, _ := decimal.NewFromString(t.VolInUsd)
				tvlInUsdV, _ := decimal.NewFromString(v.TvlInUsd)
				tvlInUsdT, _ := decimal.NewFromString(t.TvlInUsd)

				v.VolInUsd24h = volInUsd24hV.Add(volInUsd24hT).String()
				v.VolInUsd = volInUsdV.Add(volInUsdT).String()
				v.TxNum24h = v.TxNum24h + t.TxNum24h
				v.TxNum = v.TxNum + t.TxNum
				v.TvlInUsd = tvlInUsdV.Add(tvlInUsdT).String()

				has = true

				break
			}

		}

		if !has {
			swapCountToApi.Tokens = append(swapCountToApi.Tokens, t)
		}

	}
}

func removeDuplicate(nums []sol.Token) []sol.Token {
	var numLen = len(nums) - 1
	for ; numLen > 0; numLen-- {
		for i := numLen - 1; i >= 0; i-- {
			if nums[numLen].Symbol == nums[i].Symbol {
				nums = append(nums[:numLen], nums[numLen+1:]...)
				break
			}
		}
	}
	return nums
}
