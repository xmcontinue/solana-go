package process

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

type pairPrice struct {
	TokenASymbol string
	TokenBSymbol string
	newPrice     decimal.Decimal
	beforePrice  decimal.Decimal
}

type tokenPrice struct {
	newPrice    decimal.Decimal
	beforePrice decimal.Decimal
}

// SwapTotalCount 汇总统计
func SwapTotalCount() error {
	// 将要统计的数据
	swapCountToApi := &domain.SwapCountToApi{
		Pools:  make([]*domain.SwapCountToApiPool, 0),
		Tokens: make([]*domain.SwapCountToApiToken, 0),
	}

	pairPriceList, tokenPriceMap := make([]*pairPrice, 0), map[string]*tokenPrice{
		"USDC":  {decimal.NewFromInt(1), decimal.NewFromInt(1)},
		"pUSDC": {decimal.NewFromInt(1), decimal.NewFromInt(1)},
	}
	ctx := context.Background()

	// 获取swap pair 24h 内交易统计
	totalVolInUsd24h, totalVolInUsd, totalTvlInUsd, totalTxNum24h, totalTxNum, date := decimal.Decimal{}, decimal.Decimal{}, decimal.Decimal{}, uint64(0), uint64(0), time.Now().Add(-24*time.Hour)

	for _, v := range sol.SwapConfigList() {
		// 获取coingecko币价
		newSwapCount, err := model.QuerySwapCountKLine(ctx, model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = ?", "1min"), model.OrderFilter("id desc"))
		if err != nil {
			continue
		}
		beforeSwapCount, err := model.QuerySwapCountKLine(ctx, model.NewFilter("date > ?", newSwapCount.Date.Add(-24*time.Hour)), model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = ?", "1min"), model.OrderFilter("id asc"))
		if err != nil {
			continue
		}
		if newSwapCount.TokenAUSD.IsZero() || newSwapCount.TokenBUSD.IsZero() || beforeSwapCount.TokenAUSD.IsZero() || beforeSwapCount.TokenBUSD.IsZero() {
			continue
		}

		// 获取24h交易额，交易笔数 不做错误处理，有可能无交易
		swapCount24h, _ := model.SumSwapCountVolForKLines(ctx, model.NewFilter("date > ?", date), model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = ?", "1min"))

		// 获取总交易额，交易笔数 不做错误处理，有可能无交易
		swapCountTotal, _ := model.SumSwapCountVolForKLines(ctx, model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = ?", "mon"))

		// 计算vol,tvl
		tokenATvl, tokenBTvl := v.TokenA.Balance.Mul(newSwapCount.TokenAUSD).Round(6), v.TokenB.Balance.Mul(newSwapCount.TokenBUSD).Round(6)
		tokenAVol24h, tokenBVol24h := swapCount24h.TokenAVolume.Mul(newSwapCount.TokenAUSD).Round(6), swapCount24h.TokenBVolume.Mul(newSwapCount.TokenBUSD).Round(6)
		tokenAVol, tokenBVol := swapCountTotal.TokenAVolume.Mul(newSwapCount.TokenAUSD).Round(6), swapCountTotal.TokenBVolume.Mul(newSwapCount.TokenBUSD).Round(6)
		tvlInUsd, volInUsd24h, volInUsd := tokenATvl.Add(tokenBTvl), tokenAVol24h.Add(tokenBVol24h), tokenAVol.Add(tokenBVol)
		tokenATotalVol, tokenBTotalVol := tokenAVol.Add(swapCountTotal.TokenAQuoteVolume).Mul(newSwapCount.TokenAUSD).Round(6), tokenBVol.Add(swapCountTotal.TokenBQuoteVolume).Mul(newSwapCount.TokenBUSD).Round(6)

		// 计算apr
		apr := "0%"
		if !tvlInUsd.IsZero() {
			fee, _ := decimal.NewFromString(v.Fee)
			apr = volInUsd24h.Mul(fee).Mul(decimal.NewFromInt(36500)).Div(tvlInUsd).Round(2).String() + "%" // 36500为365天*百分比转化100得出
		}

		// 查找合约内价格
		newContractPrice, err := model.QuerySwapPairPriceKLine(ctx, model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = ?", "1min"), model.OrderFilter("id desc"))
		if err != nil {
			continue
		}
		beforeContractPrice, err := model.QuerySwapPairPriceKLine(ctx, model.NewFilter("date > ?", newContractPrice.Date.Add(-24*time.Hour)), model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = ?", "1min"), model.OrderFilter("id asc"))
		if err != nil {
			continue
		}
		if newContractPrice.Settle.IsZero() || beforeContractPrice.Settle.IsZero() {
			continue
		}
		// pool统计
		newSwapPrice, beforeSwapPrice := newContractPrice.Settle.Round(6), beforeContractPrice.Open.Round(6)
		swapCountToApiPool := &domain.SwapCountToApiPool{
			Name:           v.Name,
			SwapAccount:    v.SwapAccount,
			TokenAReserves: v.TokenA.SwapTokenAccount,
			TokenBReserves: v.TokenB.SwapTokenAccount,
			VolInUsd24h:    volInUsd24h.String(),
			TxNum24h:       swapCount24h.TxNum,
			VolInUsd:       volInUsd.String(),
			TxNum:          swapCountTotal.TxNum,
			Apr:            apr,
			TvlInUsd:       tvlInUsd.String(),
			PriceInterval:  v.PriceInterval,
			Price:          FormatFloat(newSwapPrice, 2),
			PriceRate24h:   newSwapPrice.Sub(beforeSwapPrice).Div(beforeSwapPrice).Mul(decimal.NewFromInt(100)).Round(2).String() + "%",
		}
		swapCountToApi.Pools = append(swapCountToApi.Pools, swapCountToApiPool)

		// token统计
		appendTokensToSwapCount(
			swapCountToApi,
			&domain.SwapCountToApiToken{
				Name:        v.TokenA.Symbol,
				VolInUsd24h: tokenAVol24h.String(),
				TxNum24h:    swapCount24h.TxNum,
				VolInUsd:    tokenATotalVol.String(),
				TxNum:       swapCountTotal.TxNum,
				TvlInUsd:    tokenATvl.String(),
			},
			&domain.SwapCountToApiToken{
				Name:        v.TokenB.Symbol,
				VolInUsd24h: tokenBVol24h.String(),
				TxNum24h:    swapCount24h.TxNum,
				VolInUsd:    tokenBTotalVol.String(),
				TxNum:       swapCountTotal.TxNum,
				TvlInUsd:    tokenBTvl.String(),
			},
		)

		pairPriceList = append(pairPriceList, &pairPrice{v.TokenA.Symbol, v.TokenB.Symbol, newSwapPrice, beforeSwapPrice})

		// 汇总处理
		totalVolInUsd24h = totalVolInUsd24h.Add(volInUsd24h)
		totalVolInUsd = totalVolInUsd.Add(volInUsd)
		totalTvlInUsd = totalTvlInUsd.Add(tvlInUsd)
		totalTxNum24h = totalTxNum24h + swapCount24h.TxNum
		totalTxNum = totalTxNum + swapCountTotal.TxNum
	}

	// 获取token价格
	pairPriceToTokenPrice(pairPriceList, tokenPriceMap)
	for _, v := range swapCountToApi.Tokens {

		if price, ok := tokenPriceMap[v.Name]; ok {
			v.Price = FormatFloat(price.newPrice, 2)
			v.PriceRate24h = price.newPrice.Sub(price.beforePrice).Div(price.beforePrice).Mul(decimal.NewFromInt(100)).Round(2).String() + "%"
		} else {
			v.Price = "0.00"
			v.PriceRate24h = "0%"
		}

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

	// 按tvl数量排序
	sort.Slice(swapCountToApi.Pools, func(i, j int) bool {
		tvl, _ := decimal.NewFromString(swapCountToApi.Pools[i].TvlInUsd)
		nextTvl, _ := decimal.NewFromString(swapCountToApi.Pools[j].TvlInUsd)
		if tvl.LessThan(nextTvl) {
			return false
		}
		if tvl.GreaterThan(nextTvl) {
			return true
		}
		return false
	})

	sort.Slice(swapCountToApi.Tokens, func(i, j int) bool {
		tvl, _ := decimal.NewFromString(swapCountToApi.Tokens[i].TvlInUsd)
		nextTvl, _ := decimal.NewFromString(swapCountToApi.Tokens[j].TvlInUsd)
		if tvl.LessThan(nextTvl) {
			return false
		}
		if tvl.GreaterThan(nextTvl) {
			return true
		}
		return false
	})

	// 缓存至redis
	data, err := json.Marshal(swapCountToApi)
	if err != nil {
		return errors.Wrap(err)
	}

	swapCountKey := domain.SwapTotalCountKey()
	if err := redisClient.Set(context.Background(), swapCountKey.Key, data, swapCountKey.Timeout).Err(); err != nil {
		return errors.Wrap(err)
	}

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

func pairPriceToTokenPrice(pairPriceList []*pairPrice, tokenPriceList map[string]*tokenPrice) {
	beforeLen := len(tokenPriceList)

	for _, v := range pairPriceList {
		tokenAPrice, tokenAHas := tokenPriceList[v.TokenASymbol]
		tokenBPrice, tokenBHas := tokenPriceList[v.TokenBSymbol]

		if tokenAHas && !tokenBHas {
			tokenPriceList[v.TokenBSymbol] = &tokenPrice{
				newPrice:    tokenAPrice.newPrice.Mul(decimal.NewFromInt(1).Div(v.newPrice)).Round(6),
				beforePrice: tokenAPrice.beforePrice.Mul(decimal.NewFromInt(1).Div(v.beforePrice)).Round(6),
			}
			continue
		}

		if tokenBHas && !tokenAHas {
			tokenPriceList[v.TokenASymbol] = &tokenPrice{
				newPrice:    tokenBPrice.newPrice.Mul(v.newPrice).Round(6),
				beforePrice: tokenBPrice.beforePrice.Mul(v.beforePrice).Round(6),
			}
			continue
		}
	}

	if beforeLen == len(tokenPriceList) {
		return
	}

	pairPriceToTokenPrice(pairPriceList, tokenPriceList)
}

func FormatFloat(num decimal.Decimal, d int) string {
	f, _ := num.Float64()
	return strconv.FormatFloat(f, 'f', d, 64)
}
