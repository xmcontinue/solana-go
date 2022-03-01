package process

import (
	"context"
	"encoding/json"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/coingecko"
	"git.cplus.link/crema/backend/pkg/domain"
)

// swapAddressLast24HVol 最近24小时swap address的总交易量
func swapAddressLast24HVol() error {
	var (
		endTime   = time.Now()
		beginTime = endTime.Add(-24 * time.Hour)
	)
	lastSwapTransactionID, err := getTransactionID()
	if err != nil {
		return errors.Wrap(err)
	}

	swapVols, err := model.SumSwapAccountLast24Vol(context.TODO(), model.SwapTransferFilter(), model.NewFilter("block_time > ?", beginTime), model.NewFilter("block_time < ?", endTime), model.NewFilter("id <= ?", lastSwapTransactionID))
	if err != nil {
		logger.Error("sum swap account from db last 24h vol err", logger.Errorv(err))
	}

	if len(swapVols) == 0 {
		return nil
	}

	swapVolMap := make(map[string]string)
	for _, v := range swapVols {
		tokenAPrice, tokenBPrice := coingecko.GetPriceForCache(v.TokenAAddress), coingecko.GetPriceForCache(v.TokenBAddress)
		v.Vol = v.TokenAVolume.Mul(tokenAPrice).Abs().Add(v.TokenBVolume.Mul(tokenBPrice).Abs())
		volCount, _ := json.Marshal(v)
		swapVolMap[domain.SwapVolCountLast24HKey(v.SwapAddress).Key] = string(volCount)
	}

	if err = redisClient.MSet(context.TODO(), swapVolMap).Err(); err != nil {
		logger.Error("sync swap account last 24h vol to redis err")
		return errors.Wrap(err)
	}
	return nil
}

// userAddressLast24hVol 普通用户最近24小时的总交易量
func userAddressLast24hVol() error {
	var (
		endTime   = time.Now()
		beginTime = endTime.Add(-24 * time.Hour)
	)
	lastSwapTransactionID, err := getTransactionID()
	if err != nil {
		return errors.Wrap(err)
	}

	swapVols, err := model.SumUserSwapAccountLast24Vol(context.TODO(), model.SwapTransferFilter(), model.NewFilter("block_time > ?", beginTime), model.NewFilter("block_time < ?", endTime), model.NewFilter("id <= ?", lastSwapTransactionID))
	if err != nil {
		logger.Error("sum swap account from db last 24h vol err", logger.Errorv(err))
	}

	if len(swapVols) == 0 {
		return nil
	}

	swapVolMap := make(map[string]string)

	for _, v := range swapVols {
		tokenAPrice, tokenBPrice := coingecko.GetPriceForCache(v.TokenAAddress), coingecko.GetPriceForCache(v.TokenBAddress)
		v.Vol = v.TokenAVolume.Mul(tokenAPrice).Abs().Add(v.TokenBVolume.Mul(tokenBPrice).Abs())
		volCount, _ := json.Marshal(v)
		swapVolMap[domain.SwapVolCountLast24HKey(v.UserAddress).Key] = string(volCount)
	}

	if err = redisClient.MSet(context.TODO(), swapVolMap).Err(); err != nil {
		logger.Error("sync swap account last 24h vol to redis err")
		return errors.Wrap(err)
	}
	return nil
}

func statisticsOneMints() {

}
