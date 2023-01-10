package process

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol"
	"git.cplus.link/crema/backend/chain/sol/parse"
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
	// get refundPositions
	// refundPositionsTvlForSymbol, err := sol.GetRefundPositionsCount()
	// if err != nil {
	// 	return errors.Wrap(err)
	// }

	// 将要统计的数据
	swapCountToApi := &domain.SwapCountToApi{
		Pools:  make([]*domain.SwapCountToApiPool, 0),
		Tokens: make([]*domain.SwapCountToApiToken, 0),
	}

	ctx := context.Background()

	// 获取swap pair 24h 内交易统计,因为时间戳的原因，这里全部改成utc时间
	totalVolInUsd24h, totalVolInUsd, totalTvlInUsd, totalTxNum24h, totalTxNum, before24hDate, before7dDate, before30dDate := decimal.Decimal{}, decimal.Decimal{}, decimal.Decimal{}, uint64(0), uint64(0), time.Now().Add(-24*time.Hour), time.Now().Add(-24*7*time.Hour), time.Now().Add(-24*30*time.Hour)

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

		// 获取最开始的日期，为了统计apr
		swapCountInfo, _ := model.QuerySwapCountKLine(ctx, model.SwapAddressFilter(v.SwapAccount), model.NewFilter("date_type = ?", "day"), model.OrderFilter("id asc"))

		// 获取24h交易额，交易笔数 不做错误处理，有可能无交易
		swapCount24h, _ := model.SumSwapCountVolForKLines(ctx, model.NewFilter("date > ?", before24hDate), model.SwapAddressFilter(v.SwapAccount), model.NewFilter("date_type = ?", "1min"))

		// 获取过去7天交易额，交易笔数 不做错误处理，有可能无交易
		swapCount7d, _ := model.SumSwapCountVolForKLines(ctx, model.NewFilter("date > ?", before7dDate), model.SwapAddressFilter(v.SwapAccount), model.NewFilter("date_type = ?", "day"))

		// 获取过去30天交易额，交易笔数 不做错误处理，有可能无交易
		swapCount30d, _ := model.SumSwapCountVolForKLines(ctx, model.NewFilter("date > ?", before30dDate), model.SwapAddressFilter(v.SwapAccount), model.NewFilter("date_type = ?", "day"))

		// 获取总交易额，交易笔数 不做错误处理，有可能无交易
		swapCountTotal, _ := model.SumSwapCountVolForKLines(ctx, model.SwapAddressFilter(v.SwapAccount), model.NewFilter("date_type = ?", "day"))

		var tvlInUsd, volInUsd decimal.Decimal

		// 计算pairs vol,tvl 计算单边
		tokenATvl, tokenBTvl := v.TokenA.Balance.Add(v.TokenA.RefundBalance).Mul(newTokenAPrice).Round(countDecimal),
			v.TokenB.Balance.Add(v.TokenB.RefundBalance).Mul(newTokenBPrice).Round(countDecimal) // v.TokenA.Balance.Add() v.TokenB.Balance.Add()

		tvlInUsd = tokenATvl.Add(tokenBTvl)
		tokenAVol, tokenBVol := swapCountTotal.TokenAVolumeForUsd.Round(countDecimal), swapCountTotal.TokenBVolumeForUsd.Round(countDecimal)
		if v.Version == "v2" {

			if swapCount24h.TxNum != 0 {
				swapCount24h.VolInUsdForContract, swapCount24h.TokenAVolumeForUsd, swapCount24h.TokenBVolumeForUsd, swapCount24h.FeeAmount, err = syncPrice(v.SwapAccount, before24hDate)
				if err != nil {
					return errors.Wrap(err)
				}
			}

			if swapCount7d.TxNum != 0 {
				swapCount7d.VolInUsdForContract, swapCount7d.TokenAVolumeForUsd, swapCount7d.TokenBVolumeForUsd, swapCount7d.FeeAmount, err = syncPrice(v.SwapAccount, before7dDate)
				if err != nil {
					return errors.Wrap(err)
				}
			}

			if swapCount30d.TxNum != 0 {
				swapCount30d.VolInUsdForContract, swapCount30d.TokenAVolumeForUsd, swapCount30d.TokenBVolumeForUsd, swapCount30d.FeeAmount, err = syncPrice(v.SwapAccount, before30dDate)
				if err != nil {
					return errors.Wrap(err)
				}
			}

			if swapCountTotal.TxNum != 0 {
				swapCountTotal.VolInUsdForContract, swapCountTotal.TokenAVolumeForUsd, swapCountTotal.TokenBVolumeForUsd, swapCountTotal.FeeAmount, err = syncPrice(v.SwapAccount, time.Now().Add(-24*time.Hour*365))
				if err != nil {
					return errors.Wrap(err)
				}
			}

			volInUsd = swapCountTotal.VolInUsdForContract
		} else {
			volInUsd = tokenAVol.Add(tokenBVol)
		}

		// -------------1day 计算--------------
		Apr24h := "0%"
		var tokenAVol24h, tokenBVol24h, volInUsd24h, tokenA24hVol, tokenB24hVol decimal.Decimal
		if swapCount24h.TxNum != 0 {
			tokenAVol24h, tokenBVol24h = swapCount24h.TokenAVolumeForUsd.Round(countDecimal), swapCount24h.TokenBVolumeForUsd.Round(countDecimal)
			if v.Version == "v2" {
				volInUsd24h = swapCount24h.VolInUsdForContract
			} else {
				volInUsd24h = tokenAVol24h.Add(tokenBVol24h)
			}

			// 下面为token交易额，算双边
			tokenA24hVol, tokenB24hVol = tokenAVol24h.Add(swapCount24h.TokenAQuoteVolumeForUsd).Round(countDecimal), tokenBVol24h.Add(swapCount24h.TokenBQuoteVolumeForUsd).Round(countDecimal)
			if !tvlInUsd.IsZero() {
				Apr24h = parse.BankToString(swapCount24h.FeeAmount.Div(tvlInUsd).Mul(decimal.NewFromInt(36500)), 2) + "%"
			}
		}

		// --------------7day 计算--------------------
		// 计算apr
		apr := "0%"
		Apr7day := "0%"
		rewarderApr := "0%"

		if !tvlInUsd.IsZero() {
			rewarderApr = parse.BankToString(v.RewarderUsd.Div(decimal.NewFromInt(2).Pow(decimal.NewFromInt(64))).Mul(decimal.NewFromInt(3600*24*365)).Div(tvlInUsd).Mul(decimal.NewFromInt(100)), 2) + "%"
		}

		var tokenAVol7d, tokenBVol7d, volInUsd7d decimal.Decimal
		var apr7DayCount int64
		if swapCount7d.TxNum != 0 {
			tokenAVol7d, tokenBVol7d = swapCount7d.TokenAVolumeForUsd.Round(countDecimal), swapCount7d.TokenBVolumeForUsd.Round(countDecimal)
			if v.Version == "v2" {
				volInUsd7d = swapCount7d.VolInUsdForContract
			} else {
				volInUsd7d = tokenAVol7d.Add(tokenBVol7d)
			}

			if swapCount7d.DayNum == 0 {
				swapCount7d.DayNum = 1
			}

			if !tvlInUsd.IsZero() {
				fee, _ := decimal.NewFromString(v.Fee)
				apr = parse.BankToString(volInUsd7d.Div(decimal.NewFromInt(7)).Mul(fee).Div(tvlInUsd).Mul(decimal.NewFromInt(36500)), 2) + "%" // 7天vol均值 * fee * 36500（365天*百分比转化100得出）/tvl
				if swapCountInfo == nil {
					apr7DayCount = 1
				} else {
					apr7DayCount = differenceDays(*swapCountInfo.Date, time.Now(), 7)
				}
				Apr7day = parse.BankToString(swapCount7d.FeeAmount.Div(decimal.NewFromInt(apr7DayCount)).Div(tvlInUsd).Mul(decimal.NewFromInt(36500)), 2) + "%"
			}
		}

		var apr30DayCount int64
		// ----------30 day 计算---------------
		Apr30day := "0%"
		if swapCount30d.TxNum != 0 {
			if swapCount30d.DayNum == 0 {
				swapCount30d.DayNum = 1
			}

			if !tvlInUsd.IsZero() {
				if swapCountInfo == nil {
					apr30DayCount = 1
				} else {
					apr30DayCount = differenceDays(*swapCountInfo.Date, time.Now(), 30)
				}
				Apr30day = parse.BankToString(swapCount30d.FeeAmount.Div(decimal.NewFromInt(apr30DayCount)).Div(tvlInUsd).Mul(decimal.NewFromInt(36500)), 2) + "%"
			}
		}
		// 下面为token交易额，算双边
		tokenATotalVol, tokenBTotalVol := tokenAVol.Add(swapCountTotal.TokenAQuoteVolumeForUsd).Round(countDecimal), tokenBVol.Add(swapCountTotal.TokenAQuoteVolumeForUsd).Round(countDecimal)

		// 查找合约内价格
		newContractPrice, err := model.QuerySwapPairPriceKLine(ctx, model.SwapAddressFilter(v.SwapAccount), model.NewFilter("date_type = ?", "1min"), model.OrderFilter("id desc"))
		if err != nil {
			logger.Error("SwapTotalCount", logger.Errorv(err))
			continue
		}

		beforeContractPrice, err := model.QuerySwapPairPriceKLine(ctx, model.NewFilter("date_type = ?", "1min"), model.NewFilter("date < ?", newContractPrice.Date.Add(-24*time.Hour)), model.SwapAddressFilter(v.SwapAccount), model.OrderFilter("id desc"))
		if err != nil {
			beforeContractPrice = newContractPrice
		}

		// 汇总处理
		totalVolInUsd = totalVolInUsd.Add(volInUsd)
		totalTvlInUsd = totalTvlInUsd.Add(tvlInUsd)
		totalTxNum24h = totalTxNum24h + swapCount24h.TxNum
		totalTxNum = totalTxNum + swapCountTotal.TxNum

		totalVolInUsd24h = totalVolInUsd24h.Add(volInUsd24h)

		if strings.ToLower(v.Version) != "v2" {
			continue // pool和token只统计v2
		}

		// pool统计
		newSwapPrice, beforeSwapPrice := newContractPrice.Settle, beforeContractPrice.Open
		if newContractPrice.Settle.IsZero() {
			logger.Error("settle is zero", logger.Errorv(errors.New("settle is zero")))
			continue
		}
		if beforeContractPrice.Open.Round(countDecimal).IsZero() {
			beforeContractPrice.Open = newContractPrice.Settle
		}
		swapCountToApiPool := &domain.SwapCountToApiPool{
			Name:                        v.Name,
			Fee7D:                       swapCount7d.FeeAmount.Round(countDecimal).String(),
			Fee30D:                      swapCount30d.FeeAmount.Round(countDecimal).String(),
			Apr7DayCount:                apr7DayCount,
			Apr30DayCount:               apr30DayCount,
			SwapAccount:                 v.SwapAccount,
			TokenAReserves:              v.TokenA.SwapTokenAccount,
			TokenBReserves:              v.TokenB.SwapTokenAccount,
			VolInUsd24h:                 volInUsd24h.Round(countDecimal).String(),
			TxNum24h:                    swapCount24h.TxNum,
			VolInUsd:                    volInUsd.Round(countDecimal).String(),
			TxNum:                       swapCountTotal.TxNum,
			Apr:                         apr,
			Fee:                         v.Fee,
			TvlInUsd:                    tvlInUsd.Round(countDecimal).String(),
			PriceInterval:               v.PriceInterval,
			Price:                       PriceFormatFloat(newSwapPrice),
			PriceRate24h:                newSwapPrice.Sub(beforeSwapPrice).Div(beforeSwapPrice).Mul(decimal.NewFromInt(100)).Round(2).String() + "%",
			VolumeInTokenA24h:           swapCount24h.TokenAVolume.Add(swapCount24h.TokenAQuoteVolume).Round(countDecimal).String(),
			VolumeInTokenB24h:           swapCount24h.TokenBVolume.Add(swapCount24h.TokenBQuoteVolume).Round(countDecimal).String(),
			VolumeInTokenA24hUnilateral: swapCount24h.TokenAVolume.Round(countDecimal).String(),
			VolumeInTokenB24hUnilateral: swapCount24h.TokenBVolume.Round(countDecimal).String(),
			TokenAAddress:               v.TokenA.TokenMint,
			TokenBAddress:               v.TokenB.TokenMint,
			Version:                     v.Version,
			Apr24h:                      Apr24h,
			Apr7Day:                     Apr7day,
			Apr30Day:                    Apr30day,
			RewarderApr:                 rewarderApr,
		}

		swapCountToApi.Pools = append(swapCountToApi.Pools, swapCountToApiPool)

		// token统计
		appendTokensToSwapCount(
			swapCountToApi,
			&domain.SwapCountToApiToken{
				Name:         v.TokenA.Symbol,
				VolInUsd24h:  tokenA24hVol.Round(countDecimal).String(),
				TxNum24h:     swapCount24h.TxNum,
				VolInUsd:     tokenATotalVol.Round(countDecimal).String(),
				TxNum:        swapCountTotal.TxNum,
				TvlInUsd:     tokenATvl.Round(countDecimal).String(),
				Price:        newTokenAPrice.StringFixedBank(int32(sol.GetTokenShowDecimalForTokenAccount(v.TokenA.SwapTokenAccount))),
				PriceRate24h: newTokenAPrice.Sub(beforeTokenAPrice).Div(beforeTokenAPrice).Mul(decimal.NewFromInt(100)).Round(2).String() + "%",
			},
			&domain.SwapCountToApiToken{
				Name:         v.TokenB.Symbol,
				VolInUsd24h:  tokenB24hVol.Round(countDecimal).String(),
				TxNum24h:     swapCount24h.TxNum,
				VolInUsd:     tokenBTotalVol.Round(countDecimal).String(),
				TxNum:        swapCountTotal.TxNum,
				TvlInUsd:     tokenBTvl.Round(countDecimal).String(),
				Price:        newTokenBPrice.StringFixedBank(int32(sol.GetTokenShowDecimalForTokenAccount(v.TokenB.SwapTokenAccount))),
				PriceRate24h: newTokenBPrice.Sub(beforeTokenBPrice).Div(beforeTokenBPrice).Mul(decimal.NewFromInt(100)).Round(2).String() + "%",
			},
		)
	}

	// token数量
	swapCountToApi.TokenNum = len(swapCountToApi.Tokens)

	// 用户数量
	// total, err := model.CountUserNumber(context.Background())
	// if err != nil {
	//	logger.Error("get user number err", logger.Errorv(err))
	//	return errors.Wrap(err)
	// }

	total, err := model.CountTransActionUserCount(context.Background())
	if err != nil {
		logger.Error("get user number err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	swapCountToApi.UserNum = total
	// 总交易额
	swapCountToApi.VolInUsd24h = totalVolInUsd24h.Round(countDecimal).String()
	swapCountToApi.VolInUsd = totalVolInUsd.Round(countDecimal).String()
	swapCountToApi.TvlInUsd = totalTvlInUsd.Round(countDecimal).String()
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

	var swapCountKey domain.RedisKey
	if model.ISSharding {
		swapCountKey = domain.SwapTotalCountKeyWithSharding()
	} else {
		swapCountKey = domain.SwapTotalCountKey()
	}

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

func PriceFormatFloat(num decimal.Decimal) string {
	f, _ := num.Float64()
	if num.LessThan(decimal.NewFromFloat(0.000001)) {
		if num.LessThan(decimal.NewFromFloat(0.00000001)) {
			return strconv.FormatFloat(f, 'f', 10, 64)
		} else {
			return strconv.FormatFloat(f, 'f', 8, 64)
		}
	} else {
		return strconv.FormatFloat(f, 'f', 6, 64)
	}
}

func differenceDays(start, end time.Time, max int64) int64 {
	n := int64(end.Sub(start).Hours()/24) + 1
	if n < max {
		return n
	}
	return max
}
