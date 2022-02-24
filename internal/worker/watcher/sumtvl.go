package watcher

import (
	"context"
	"encoding/json"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/coingecko"
	"git.cplus.link/crema/backend/pkg/domain"
)

func SyncVol24H() error {
	logger.Info("24h vol syncing ......")

	keys := sol.SwapConfigList()

	tvl, pairs := domain.Tvl{}, make([]domain.PairTvl, 0, len(keys))
	totalTvlInUsd, totalVolInUsd := decimal.Decimal{}, decimal.Decimal{}

	// 获取单个swap pair count
	for _, v := range keys {
		count, err := model.QuerySwapPairCount(context.TODO(), model.NewFilter("token_swap_address = ?", v.SwapAccount))
		if err != nil {
			continue
		}

		tvlInUsd, volInUsd, apr := compute(count, v.Fee)

		totalTvlInUsd = totalTvlInUsd.Add(tvlInUsd)
		totalVolInUsd = totalVolInUsd.Add(volInUsd)

		pairs = append(pairs, domain.PairTvl{
			Name:        count.PairName,
			TvlInUsd:    tvlInUsd.String(),
			VolInUsd:    volInUsd.String(),
			TxNum:       count.TxNum,
			SwapAccount: v.SwapAccount,
			Apr:         apr,
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
		logger.Error("24h vol sync fail:", logger.Errorv(err))
		return errors.Wrap(err)
	}

	logger.Info("24h vol sync complete!")
	return nil
}

func SyncTotalVol() error {
	logger.Info("total vol syncing ......")

	keys := sol.SwapConfigList()

	ctx := context.Background()

	for _, v := range keys {

		info, err := model.QuerySwapPairBase(ctx, model.SwapAddress(v.SwapAccount))
		if err != nil || !info.IsSync {
			logger.Info("vol sync Failed : the transaction did not complete synchronously", logger.String("swap_address:", v.SwapAccount))
			continue
		}

		vol, err := model.CountTxNum(ctx, model.SwapAddress(v.SwapAccount), model.SwapTransferFilter())
		if err != nil {
			continue
		}

		err = model.UpdateSwapPairBase(ctx, map[string]interface{}{"total_tx_num": vol.TxNum, "total_vol": vol.TotalVol}, model.SwapAddress(v.SwapAccount))
		if err != nil {
			continue
		}

		logger.Info("vol sync complete ", logger.String("swap_address:", v.SwapAccount))
	}

	logger.Info("total vol sync complete!")

	return nil
}

// compute 计算 tvl vol apr数量
func compute(count *domain.SwapPairCount, feeStr string) (decimal.Decimal, decimal.Decimal, string) {
	tvlInUsd, volInUsd, apr := decimal.Decimal{}, decimal.Decimal{}, ""
	// token 价格
	tokenAPrice, _ := coingecko.GetPriceFromTokenAccount(count.TokenAPoolAddress)
	tokenBPrice, _ := coingecko.GetPriceFromTokenAccount(count.TokenBPoolAddress)

	// token 余额
	tokenABalance := count.TokenABalance.Mul(tokenAPrice)
	tokenBBalance := count.TokenBBalance.Mul(tokenBPrice)
	// token 交易额
	tokenAVolume := count.TokenAVolume.Mul(tokenAPrice)
	tokenBVolume := count.TokenBVolume.Mul(tokenBPrice)

	tvlInUsd = tokenABalance.Add(tokenBBalance)
	volInUsd = tokenAVolume.Add(tokenBVolume)

	if tvlInUsd.IsZero() {
		apr = "0%"
	} else {
		fee, _ := decimal.NewFromString(feeStr)
		apr = volInUsd.Mul(fee).Mul(decimal.NewFromInt(36500)).Div(tvlInUsd).Round(2).String() + "%" // 36500为365天*百分比转化100得出
	}

	return tvlInUsd, volInUsd, apr
}
