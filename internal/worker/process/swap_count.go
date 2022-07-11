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

const countDecimal = 6

// SwapTotalCount 汇总统计
func SwapTotalCount() error {
	// 将要统计的数据
	swapCountToApi := &domain.SwapCountToApi{
		Pools:  make([]*domain.SwapCountToApiPool, 0),
		Tokens: make([]*domain.SwapCountToApiToken, 0),
	}

	ctx := context.Background()

	// 获取swap pair 24h 内交易统计
	totalVolInUsd24h, totalVolInUsd, totalTvlInUsd, totalTxNum24h, totalTxNum, before24hDate, before7dDate := decimal.Decimal{}, decimal.Decimal{}, decimal.Decimal{}, uint64(0), uint64(0), time.Now().Add(-24*time.Hour), time.Now().Add(-24*7*time.Hour)

	for _, v := range sol.SwapConfigList() {
		// 获取token价格
		newTokenAPrice, err := model.GetPriceForSymbol(ctx, v.TokenA.Symbol)
		newTokenBPrice, err := model.GetPriceForSymbol(ctx, v.TokenB.Symbol)
		if err != nil || newTokenAPrice.IsZero() || newTokenBPrice.IsZero() {
			continue
		}

		beforeTokenAPrice, err := model.GetPriceForSymbol(ctx, v.TokenA.Symbol, model.NewFilter("date < ?", before24hDate))
		if err != nil || newTokenAPrice.IsZero() {
			beforeTokenAPrice = newTokenAPrice
		}
		beforeTokenBPrice, err := model.GetPriceForSymbol(ctx, v.TokenB.Symbol, model.NewFilter("date < ?", before24hDate))
		if err != nil || newTokenBPrice.IsZero() {
			beforeTokenBPrice = newTokenBPrice
		}

		// 获取24h交易额，交易笔数 不做错误处理，有可能无交易
		swapCount24h, _ := model.SumSwapCountVolForKLines(ctx, model.NewFilter("date > ?", before24hDate), model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = ?", "1min"))

		// 获取过去7天交易额，交易笔数 不做错误处理，有可能无交易
		swapCount7d, _ := model.SumSwapCountVolForKLines(ctx, model.NewFilter("date > ?", before7dDate), model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = ?", "1min"))

		// 获取总交易额，交易笔数 不做错误处理，有可能无交易
		swapCountTotal, _ := model.SumSwapCountVolForKLines(ctx, model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = ?", "1min"))

		// 计算pairs vol,tvl 计算单边
		tokenATvl, tokenBTvl := v.TokenA.Balance.Mul(newTokenAPrice).Round(countDecimal), v.TokenB.Balance.Mul(newTokenBPrice).Round(countDecimal)
		tokenAVol24h, tokenBVol24h := swapCount24h.TokenAVolumeForUsd.Round(countDecimal), swapCount24h.TokenBVolumeForUsd.Round(countDecimal)
		tokenAVol7d, tokenBVol7d := swapCount7d.TokenAVolumeForUsd.Round(countDecimal), swapCount7d.TokenBVolumeForUsd.Round(countDecimal)
		tokenAVol, tokenBVol := swapCountTotal.TokenAVolumeForUsd.Round(countDecimal), swapCountTotal.TokenBVolumeForUsd.Round(countDecimal)
		tvlInUsd, volInUsd24h, volInUsd7d, volInUsd := tokenATvl.Add(tokenBTvl), tokenAVol24h.Add(tokenBVol24h), tokenAVol7d.Add(tokenBVol7d), tokenAVol.Add(tokenBVol)

		// 下面为token交易额，算双边
		tokenA24hVol, tokenB24hVol := tokenAVol24h.Add(swapCount24h.TokenAQuoteVolumeForUsd).Round(countDecimal), tokenBVol24h.Add(swapCount24h.TokenBQuoteVolumeForUsd).Round(countDecimal)
		tokenATotalVol, tokenBTotalVol := tokenAVol.Add(swapCountTotal.TokenAQuoteVolumeForUsd).Round(countDecimal), tokenBVol.Add(swapCountTotal.TokenAQuoteVolumeForUsd).Round(countDecimal)

		// 计算apr
		apr := "0%"
		if !tvlInUsd.IsZero() {
			fee, _ := decimal.NewFromString(v.Fee)
			apr = volInUsd7d.Div(decimal.NewFromInt(7)).Mul(fee).Mul(decimal.NewFromInt(36500)).Div(tvlInUsd).Round(2).String() + "%" // 7天vol均值 * fee * 36500（365天*百分比转化100得出）/tvl
		}

		// 查找合约内价格
		newContractPrice, err := model.QuerySwapPairPriceKLine(ctx, model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = ?", "1min"), model.OrderFilter("id desc"))
		beforeContractPrice, err := model.QuerySwapPairPriceKLine(ctx, model.NewFilter("date > ?", newContractPrice.Date.Add(-24*time.Hour)), model.SwapAddress(v.SwapAccount), model.NewFilter("date_type = ?", "1min"), model.OrderFilter("id asc"))
		if err != nil || newContractPrice.Settle.IsZero() {
			continue
		}
		if beforeContractPrice.Open.IsZero() {
			beforeContractPrice.Open = newContractPrice.Settle
		}

		// pool统计
		newSwapPrice, beforeSwapPrice := newContractPrice.Settle.Round(countDecimal), beforeContractPrice.Open.Round(countDecimal)
		swapCountToApiPool := &domain.SwapCountToApiPool{
			Name:                        v.Name,
			SwapAccount:                 v.SwapAccount,
			TokenAReserves:              v.TokenA.SwapTokenAccount,
			TokenBReserves:              v.TokenB.SwapTokenAccount,
			VolInUsd24h:                 volInUsd24h.String(),
			TxNum24h:                    swapCount24h.TxNum,
			VolInUsd:                    volInUsd.String(),
			TxNum:                       swapCountTotal.TxNum,
			Apr:                         apr,
			TvlInUsd:                    tvlInUsd.String(),
			PriceInterval:               v.PriceInterval,
			Price:                       FormatFloat(newSwapPrice, countDecimal),
			PriceRate24h:                newSwapPrice.Sub(beforeSwapPrice).Div(beforeSwapPrice).Mul(decimal.NewFromInt(100)).Round(2).String() + "%",
			VolumeInTokenA24h:           swapCount24h.TokenAVolume.Add(swapCount24h.TokenAQuoteVolume).Round(countDecimal).String(),
			VolumeInTokenB24h:           swapCount24h.TokenBVolume.Add(swapCount24h.TokenBQuoteVolume).Round(countDecimal).String(),
			VolumeInTokenA24hUnilateral: swapCount24h.TokenAVolume.Round(countDecimal).String(),
			VolumeInTokenB24hUnilateral: swapCount24h.TokenBVolume.Round(countDecimal).String(),
			TokenAAddress:               v.TokenA.TokenMint,
			TokenBAddress:               v.TokenB.TokenMint,
		}
		swapCountToApi.Pools = append(swapCountToApi.Pools, swapCountToApiPool)

		// token统计
		appendTokensToSwapCount(
			swapCountToApi,
			&domain.SwapCountToApiToken{
				Name:         v.TokenA.Symbol,
				VolInUsd24h:  tokenA24hVol.String(),
				TxNum24h:     swapCount24h.TxNum,
				VolInUsd:     tokenATotalVol.String(),
				TxNum:        swapCountTotal.TxNum,
				TvlInUsd:     tokenATvl.String(),
				Price:        newTokenAPrice.StringFixedBank(int32(sol.GetTokenShowDecimalForTokenAccount(v.TokenA.SwapTokenAccount))),
				PriceRate24h: newTokenAPrice.Sub(beforeTokenAPrice).Div(beforeTokenAPrice).Mul(decimal.NewFromInt(100)).Round(2).String() + "%",
			},
			&domain.SwapCountToApiToken{
				Name:         v.TokenB.Symbol,
				VolInUsd24h:  tokenB24hVol.String(),
				TxNum24h:     swapCount24h.TxNum,
				VolInUsd:     tokenBTotalVol.String(),
				TxNum:        swapCountTotal.TxNum,
				TvlInUsd:     tokenBTvl.String(),
				Price:        newTokenBPrice.StringFixedBank(int32(sol.GetTokenShowDecimalForTokenAccount(v.TokenB.SwapTokenAccount))),
				PriceRate24h: newTokenBPrice.Sub(beforeTokenBPrice).Div(beforeTokenBPrice).Mul(decimal.NewFromInt(100)).Round(2).String() + "%",
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
	total, err := model.CountUserNumber(context.Background())
	if err != nil {
		return errors.Wrap(err)
	}

	swapCountToApi.UserNum = total
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

func FormatFloat(num decimal.Decimal, d int) string {
	f, _ := num.Float64()
	return strconv.FormatFloat(f, 'f', d, 64)
}
