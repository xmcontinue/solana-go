package process

import (
	"context"
	"encoding/json"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

// swapAddressLast24HVol 最近24小时swap address的总交易量
func swapAddressLast24HVol() error {
	var (
		endTime   = time.Now()
		beginTime = endTime.Add(-108 * time.Hour) // todo 修改为24
	)

	swapVols, err := model.SumSwapAccountLast24Vol(context.TODO(), model.SwapTransferFilter(), model.NewFilter("block_time > ?", beginTime), model.NewFilter("block_time < ?", endTime))
	if err != nil {
		logger.Error("sum swap account from db last 24h vol err", logger.Errorv(err))
	}

	if len(swapVols) == 0 {
		return nil
	}

	swapVolMap := make(map[string]string)
	for _, v := range swapVols {
		volCount, _ := json.Marshal(v)
		swapVolMap[domain.SwapVolCountLast24HKey(v.AccountAddress).Key] = string(volCount)
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
		beginTime = endTime.Add(-108 * time.Hour) // todo 修改为24
	)

	swapVols, err := model.SumUserSwapAccountLast24Vol(context.TODO(), model.SwapTransferFilter(), model.NewFilter("block_time > ?", beginTime), model.NewFilter("block_time < ?", endTime))
	if err != nil {
		logger.Error("sum swap account from db last 24h vol err", logger.Errorv(err))
	}

	if len(swapVols) == 0 {
		return nil
	}

	swapVolMap := make(map[string]string)
	for _, v := range swapVols {
		volCount, _ := json.Marshal(v)
		swapVolMap[domain.SwapVolCountLast24HKey(v.AccountAddress).Key] = string(volCount)
	}

	if err = redisClient.MSet(context.TODO(), swapVolMap).Err(); err != nil {
		logger.Error("sync swap account last 24h vol to redis err")
		return errors.Wrap(err)
	}
	return nil
}
