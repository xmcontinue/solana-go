package process

import (
	"context"
	"encoding/json"
	"strconv"
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

type Price struct {
	Open   decimal.Decimal
	High   decimal.Decimal
	Low    decimal.Decimal
	Settle decimal.Decimal
	Avg    decimal.Decimal
	Date   *time.Time
}

func syncDateTypeKLine(ctx context.Context, klineTyp KLineTyp, swapAccount string) error {
	var (
		key  = domain.KLineKey(klineTyp.DateType, swapAccount)
		date = &time.Time{}
	)

	lastValue, err := redisClient.ZRange(ctx, key, -1, -1).Result() // 返回最后一个元素
	if err != nil && !redisClient.ErrIsNil(err) {
		return errors.Wrap(err)
	}

	if len(lastValue) != 0 {
		pri := &Price{}
		if err = json.Unmarshal([]byte(lastValue[0]), pri); err != nil {
			return errors.Wrap(err)
		}
		date = pri.Date

		// 最后一个数据会重复，提前删除，以便于更新
		if lastValue != nil {
			if err = redisClient.ZRemRangeByScore(ctx, key, strconv.FormatInt(date.Unix(), 10), strconv.FormatInt(date.Unix(), 10)).Err(); err != nil {
				return errors.Wrap(err)
			}
		}
	}

	for {
		swapCountKlines, err := model.QuerySwapCountKLines(ctx, 1000, 0,
			model.NewFilter("date_type = ?", klineTyp.DateType),
			model.OrderFilter("date asc"),
			model.NewFilter("date >= ?", date),
			model.SwapAddress(swapAccount),
		)
		if err != nil {
			return errors.Wrap(err)
		}

		prices := make([]*redis.Z, 0, len(swapCountKlines))
		for _, v := range swapCountKlines {
			p, _ := json.Marshal(Price{
				Open:   v.Open,
				High:   v.High,
				Low:    v.Low,
				Settle: v.Settle,
				Avg:    v.Avg,
				Date:   v.Date,
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

		date = swapCountKlines[len(swapCountKlines)-1].Date
		// 退出条件
		if len(swapCountKlines) == 1 {
			break
		}
	}

	return nil
}

// 采用redis list 数据结构，先查询是否有数据存在，如果没有则同步全部数据，有则现获取已同步的数据的最后一条，然后同步新数据
func syncKLineToRedis() error {
	ctx := context.Background()
	for _, swapConfig := range sol.SwapConfigList() {

		swapPairBase, err := model.QuerySwapPairBase(ctx, model.SwapAddress(swapConfig.SwapAccount))
		if err != nil {
			logger.Error("query swap_pair_bases err", logger.Errorv(err))
			return errors.Wrap(err)
		}
		if swapPairBase == nil {
			break
		}

		if swapPairBase.IsSync == false {
			break
		}

		for _, v := range []KLineTyp{DateMin, DateTwelfth, DateQuarter, DateHalfAnHour, DateHour, DateDay, DateWek, DateMon} {
			if err = syncDateTypeKLine(ctx, v, swapConfig.SwapAccount); err != nil {
				return errors.Wrap(err)
			}
		}
	}
	return nil
}
