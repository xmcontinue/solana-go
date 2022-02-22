package process

import (
	"context"
	"fmt"
	"time"

	"git.cplus.link/go/akit/errors"

	model "git.cplus.link/crema/backend/internal/model/market"
)

//
//// syncSwapAddressLastTvl swap address 最新tvl
//func syncSwapAddressLastTvl() error {
//	swapAddress, err := model.QuerySwapAddressGroupBySwapAddress(context.TODO())
//	if err != nil {
//		return errors.Wrap(err)
//	}
//	if len(swapAddress) == 0 {
//		return errors.Wrap(err)
//	}
//
//	ctx := context.Background()
//	for _, swapAddr := range swapAddress {
//
//		swapTvlCounts, total, err := model.QuerySwapTvlCount(ctx, 1, 0, model.NewFilter("swap_address =?", swapAddr))
//		if err != nil {
//			return errors.Wrap(err)
//		}
//
//		if total == 0 {
//			continue
//		}
//
//		data, _ := json.Marshal(swapTvlCounts)
//		if err = redisClient.Set(ctx, getSwapTvlCountKey(swapTvlCounts[0].SwapAddress), data, 0).Err(); err != nil {
//			return errors.Wrap(err)
//		}
//
//	}
//	return nil
//}

func getSwapTvlCountKey(key string) string {
	return fmt.Sprintf("swap:lasttvl:%s", key)
}

// swapAddressLast24HVol 最近24小时swap address的总交易量
func swapAddressLast24HVol() error {
	swapAddress, err := model.QuerySwapAddressGroupBySwapAddress(context.TODO())
	if err != nil {
		return errors.Wrap(err)
	}
	if len(swapAddress) == 0 {
		return errors.Wrap(err)
	}

	ctx := context.Background()
	endTime := time.Now().UTC()
	beginTime := endTime.Add(-24 * time.Hour)
	for _, swapAddr := range swapAddress {
		sumVol, err := model.SumLast24hVol(ctx, "swap_address", model.NewFilter("swap_address =?", swapAddr), model.NewFilter("block_time > ?", beginTime), model.NewFilter("block_time < ?", endTime))
		if err != nil {
			return errors.Wrap(err)
		}

		if err = redisClient.Set(ctx, getSwapTvlCountLast24HKey(swapAddr), sumVol.String(), 0).Err(); err != nil {
			return errors.Wrap(err)
		}

	}

	return nil
}

func getSwapTvlCountLast24HKey(key string) string {
	return fmt.Sprintf("swap:vol:count:last24h:%s", key)
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

		if err = redisClient.Set(ctx, getUserSwapVolCountLast24HKey(userAddr), sumVol.String(), time.Hour*24).Err(); err != nil { // todo 如何设置过期时间
			return errors.Wrap(err)
		}

	}

	return nil
}

func getUserSwapVolCountLast24HKey(key string) string {
	return fmt.Sprintf("user:swap:vol:count:last24h:%s", key)
}
