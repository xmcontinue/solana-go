package watcher

import (
	"context"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

func SyncVol24H() error {
	logger.Info("24h vol syncing ......")

	keys := sol.SwapConfigList()

	tvl, pairs := domain.Tvl{}, domain.PairTvlList{}
	totalTvlInUsd, totalVolInUsd, txNum, cumuTxNum, cumuVolInUsd := decimal.Decimal{}, decimal.Decimal{}, uint64(0), uint64(0), decimal.Decimal{}
	ctx := context.Background()
	// 获取单个swap pair count
	for _, v := range keys {
		count, err := model.QuerySwapPairCount(ctx, model.NewFilter("token_swap_address = ?", v.SwapAccount))
		if err != nil {
			continue
		}

		tvlInUsd, volInUsd, apr := compute(count, v.Fee)

		totalTvlInUsd = totalTvlInUsd.Add(tvlInUsd)
		totalVolInUsd = totalVolInUsd.Add(volInUsd)
		txNum = txNum + count.TxNum

		swapInfo, err := model.QuerySwapPairBase(ctx, model.SwapAddress(v.SwapAccount))

		cumuTxNum = cumuTxNum + swapInfo.TotalTxNum
		cumuVolInUsd = cumuVolInUsd.Add(swapInfo.TotalVol)

		pairs = append(pairs, &domain.PairTvl{
			Name:         count.PairName,
			TvlInUsd:     tvlInUsd.String(),
			VolInUsd:     volInUsd.String(),
			TxNum:        count.TxNum,
			SwapAccount:  v.SwapAccount,
			Apr:          apr,
			CumuTxNum:    swapInfo.TotalTxNum,
			CumuVolInUsd: swapInfo.TotalVol.String(),
		})
	}

	if len(pairs) > 0 {
		tvl.Pairs = &pairs
		tvl.TotalTvlInUsd = totalTvlInUsd.String()
		tvl.TotalVolInUsd = totalVolInUsd.String()
		tvl.CumuTxNum = cumuTxNum
		tvl.TxNum = txNum
		tvl.CumuVolInUsd = cumuVolInUsd.String()
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
		// v.TokenA.Symbol
		// 获取token价格
		tokenAPrice, err := model.GetPriceForSymbol(ctx, v.TokenA.Symbol)
		tokenBPrice, err := model.GetPriceForSymbol(ctx, v.TokenB.Symbol)
		if err != nil || tokenAPrice.IsZero() || tokenBPrice.IsZero() {
			continue
		}

		err = model.UpdateSwapPairBase(ctx, map[string]interface{}{"total_tx_num": vol.TxNum, "total_vol": vol.TokenATotalVol.Mul(tokenAPrice).Add(vol.TokenBTotalVol.Mul(tokenBPrice))}, model.SwapAddress(v.SwapAccount))
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
	// 获取token价格
	tokenAPrice, err := model.GetPriceForSymbol(context.TODO(), count.TokenASymbol)
	tokenBPrice, err := model.GetPriceForSymbol(context.TODO(), count.TokenBSymbol)
	if err != nil || tokenAPrice.IsZero() || tokenBPrice.IsZero() {
		return tvlInUsd, volInUsd, apr
	}

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
