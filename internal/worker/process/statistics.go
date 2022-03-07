package process

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"
	"github.com/go-redis/redis/v8"

	"git.cplus.link/crema/backend/chain/sol"
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
		tokenAPrice, tokenBPrice := coingecko.GetPriceForTokenAccount(v.TokenAAddress), coingecko.GetPriceForTokenAccount(v.TokenBAddress)
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
		tokenAPrice, tokenBPrice := coingecko.GetPriceForTokenAccount(v.TokenAAddress), coingecko.GetPriceForTokenAccount(v.TokenBAddress)
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

// 同步用户交易额数据到redis 启动时第一次全部同步，后续的同步只同步间隔两倍同步时间内与更新的数据，或者是同步完一次之后，就记录一下同步时间，下次同步时，数据的更新时间必须大于等于同步时间
func syncTORedis() error {
	ctx := context.Background()
	for _, swapConfig := range sol.SwapConfigList() {
		swapCount, err := model.QuerySwapCount(ctx, model.OrderFilter("id desc"), model.SwapAddress(swapConfig.SwapAccount))
		if err != nil {
			return errors.Wrap(err)
		}

		// swap address 最新tvl,单位是价格
		swapCountKey := domain.SwapCountKey(swapCount.SwapAddress)
		if err = redisClient.Set(ctx, swapCountKey.Key, swapCount.TokenABalance.Add(swapCount.TokenBBalance).String(), swapCountKey.Timeout).Err(); err != nil {
			return errors.Wrap(err)
		}

		// swap address 总的交易额（vol），单位是价格
		swapVolKey := domain.AccountSwapVolCountKey(swapCount.SwapAddress, "")
		if err = redisClient.Set(ctx, swapVolKey.Key, swapCount.TokenAVolume.Add(swapCount.TokenBVolume).String(), swapVolKey.Timeout).Err(); err != nil {
			return errors.Wrap(err)
		}
	}

	// user address 总的交易额（vol）
	index := int64(0)
	for {
		userSwapCount, err := model.QueryUserSwapCounts(ctx, 1000, 0, model.NewFilter("id > ?", index), model.OrderFilter("id asc"))
		if err != nil {
			return errors.Wrap(err)
		}

		if len(userSwapCount) == 0 {
			break
		}

		userSwapCountMap := make(map[string]string)
		for _, v := range userSwapCount {
			userVolKey := domain.AccountSwapVolCountKey(v.UserAddress, v.SwapAddress)
			userSwapCountMap[userVolKey.Key] = v.UserTokenBVolume.Add(v.UserTokenBVolume).String()
		}

		if err = redisClient.MSet(ctx, userSwapCountMap).Err(); err != nil {
			return errors.Wrap(err)
		}

		index = userSwapCount[len(userSwapCount)-1].ID
	}

	return nil
}

// 采用redis list 数据结构，先查询是否有数据存在，如果没有则同步全部诗句，有则现获取已同步的数据的最后一条，然后同步新数据
func syncKLine() error {
	ctx := context.Background()
	for _, v := range []KLineTyp{dateMin, dateTwelfth, dateQuarter, dateHalfAnHour, dateHour, dateDay, dateWek, dateMon} {
		swapCountKlines, err := model.QuerySwapCountKLines(ctx, 1000, 0, model.NewFilter("date_type = ?", v.DateType), model.OrderFilter("id asc"))
		if err != nil {
			return errors.Wrap(err)
		}

		type price struct {
			Open   decimal.Decimal
			High   decimal.Decimal
			Low    decimal.Decimal
			Settle decimal.Decimal
			Avg    decimal.Decimal
		}
		key := getKLineKey(v.DateType)
		prices := make([]*redis.Z, 0, len(swapCountKlines))
		for _, v := range swapCountKlines {
			p, _ := json.Marshal(price{
				Open:   v.Open,
				High:   v.High,
				Low:    v.Low,
				Settle: v.Settle,
				Avg:    v.Avg,
			})

			prices = append(prices, &redis.Z{
				Score:  float64(v.Date.Unix()),
				Member: string(p),
			})
		}

		if err = redisClient.ZAdd(context.TODO(), key, prices...).Err(); err != nil {
			logger.Error("sync swap account last 24h vol to redis err")
			return errors.Wrap(err)
		}
	}
	return nil
}

func getKLineKey(dateType domain.DateType) string {
	return fmt.Sprintf("kline:swap:count:%s", dateType)
}
