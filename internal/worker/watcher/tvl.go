package watcher

import (
	"context"
	"encoding/json"
	"math"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

// SyncTvl 同步总Tvl
func SyncTvl() error {
	logger.Info("total tvl syncing ......")

	keys := sol.SwapConfigList()

	tvl, pairs := domain.Tvl{}, make([]domain.PairTvl, 0, len(keys))
	totalTvlInUsd, totalVolInUsd := decimal.Decimal{}, decimal.Decimal{}

	// 获取单个swap pair count
	for _, v := range keys {
		count, err := model.QuerySwapPairCount(context.TODO(), model.NewFilter("token_swap_address = ?", v.SwapAccount))
		if err != nil {
			continue
		}

		tvlInUsd, volInUsd := compute(count)

		totalTvlInUsd = totalTvlInUsd.Add(tvlInUsd)
		totalVolInUsd = totalVolInUsd.Add(volInUsd)

		pairs = append(pairs, domain.PairTvl{
			Name:     count.PairName,
			TvlInUsd: tvlInUsd.String(),
			VolInUsd: volInUsd.String(),
		})
	}

	if len(pairs) > 0 {
		tvl.TotalTvlInUsd = totalTvlInUsd.String()
		tvl.TotalVolInUsd = totalVolInUsd.String()
		b, _ := json.Marshal(pairs)
		tvl.Pairs = domain.JsonString(b)
	}

	err := model.CreateTvl(context.TODO(), &tvl)
	if err != nil {
		logger.Error("total tvl sync fail:", logger.Errorv(err))
		return errors.Wrap(err)
	}

	logger.Info("total tvl sync complete!")
	return nil
}

// compute 计算数量
func compute(count *domain.SwapPairCount) (decimal.Decimal, decimal.Decimal) {
	tvlInUsd, volInUsd := decimal.Decimal{}, decimal.Decimal{}
	// token 价格 TODO 待获取价格
	tokenAPrice, tokenBPrice := decimal.NewFromInt(1), decimal.NewFromInt(1)
	// token 余额
	tokenABalance := precisionConversion(int64(count.TokenABalance), count.TokenADecimal).Mul(tokenAPrice)
	tokenBBalance := precisionConversion(int64(count.TokenBBalance), count.TokenBDecimal).Mul(tokenBPrice)
	// token 交易额
	tokenAVolume := precisionConversion(int64(count.TokenAVolume), count.TokenADecimal).Mul(tokenAPrice)
	tokenBVolume := precisionConversion(int64(count.TokenBVolume), count.TokenBDecimal).Mul(tokenBPrice)

	tvlInUsd = tokenABalance.Add(tokenBBalance)
	volInUsd = tokenAVolume.Add(tokenBVolume)

	return tvlInUsd, volInUsd
}

// precisionConversion 精度转换
func precisionConversion(num int64, precision int) decimal.Decimal {
	return decimal.NewFromInt(num).Div(decimal.NewFromFloat(math.Pow10(precision)))
}
