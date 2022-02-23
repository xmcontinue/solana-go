package process

import (
	"context"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

// swapAddressLast24HVol 最近24小时swap address的总交易量
func swapAddressLast24HVol() error {

	endTime := time.Now().UTC()
	beginTime := endTime.Add(-24 * time.Hour)
	swapVols, err := model.SumSwapAccountLast24Vol(context.TODO(), model.NewFilter("block_time > ?", beginTime), model.NewFilter("block_time < ?", endTime))
	if err != nil {
		logger.Error("sum swap account from db last 24h vol err", logger.Errorv(err))
	}

	swapVol := make(map[string]string)
	for _, v := range swapVols {
		swapVol[v.AccountAddress] = v.Vol.String()
	}

	if err = redisClient.MSet(context.TODO(), swapVol).Err(); err != nil {
		logger.Error("sync swap account last 24h vol to redis err")
		return errors.Wrap(err)
	}
	return nil
}

// userAddressLast24hVol 普通用户最近24小时的总交易量
func userAddressLast24hVol() error {
	var (
		typ       = "user_address"
		endTime   = time.Now().UTC()
		beginTime = endTime.Add(-24 * time.Hour)
		ctx       = context.Background()
	)

	userAddress, err := model.QueryUserAddressGroupByUserAddress(ctx)
	if err != nil {
		return errors.Wrap(err)
	}

	if len(userAddress) == 0 {
		return nil
	}

	for _, userAddr := range userAddress {
		sumVol, err := model.SumLast24hVol(ctx, typ, model.NewFilter("user_address = ?", userAddr), model.NewFilter("block_time > ?", beginTime), model.NewFilter("block_time < ?", endTime))
		if err != nil {
			return errors.Wrap(err)
		}
		redisKey := domain.SwapVolCountLast24HKey(userAddr)
		if err = redisClient.Set(ctx, redisKey.Key, sumVol.String(), redisKey.Timeout).Err(); err != nil { // todo 如何设置过期时间
			return errors.Wrap(err)
		}

	}

	return nil
}
