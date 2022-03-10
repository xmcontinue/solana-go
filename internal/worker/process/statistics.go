package process

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
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
		return nil
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

// 同步用户交易额数据到redis 启动时第一次全部同步，后续的同步只同步间隔两倍同步时间内与更新的数据，或者是同步完一次之后，就记录一下同步时间，下次同步时，数据的更新时间必须大于等于同步时间
func syncTORedis() error {
	ctx := context.Background()
	for _, swapConfig := range sol.SwapConfigList() {
		swapCount, err := model.QuerySwapCount(ctx, model.OrderFilter("id desc"), model.SwapAddress(swapConfig.SwapAccount))
		if err != nil {
			return errors.Wrap(err)
		}

		// swap address 最新tvl,单位是价格
		swapCountKey := domain.SwapTvlCountKey(swapCount.SwapAddress)
		if err = redisClient.Set(ctx, swapCountKey.Key, swapCount.TokenABalance.Add(swapCount.TokenBBalance).String(), swapCountKey.Timeout).Err(); err != nil {
			return errors.Wrap(err)
		}

		// swap address 总的交易额（vol），单位是价格
		swapVolKey := domain.AccountSwapVolCountKey(swapCount.SwapAddress)
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
			userVolKey := domain.AccountSwapVolCountKey(v.UserAddress)
			userSwapCountMap[userVolKey.Key] = v.UserTokenBVolume.Add(v.UserTokenBVolume).String()
		}

		if err = redisClient.MSet(ctx, userSwapCountMap).Err(); err != nil {
			return errors.Wrap(err)
		}

		index = userSwapCount[len(userSwapCount)-1].ID
	}

	return nil
}

func syncDateTypeKLine(ctx context.Context, klineTyp KLineTyp, swapAccount string) error {
	var (
		key            = domain.KLineKey(klineTyp.DateType, swapAccount)
		date           = &time.Time{}
		isDelete       = false
		lastValuePrice = &Price{}
		lastValue      []string
		err            error
	)

	lastValue, err = redisClient.ZRange(ctx, key, -1, -1).Result() // 返回最后一个元素
	if err != nil && !redisClient.ErrIsNil(err) {
		return errors.Wrap(err)
	}

	if len(lastValue) != 0 {
		isDelete = true

		if err = json.Unmarshal([]byte(lastValue[0]), lastValuePrice); err != nil {
			return errors.Wrap(err)
		}
		date = lastValuePrice.Date

	}

	swapCountKlines, err := model.QuerySwapCountKLines(ctx, klineTyp.DataCount, 0,
		model.NewFilter("date_type = ?", klineTyp.DateType),
		model.NewFilter("date >= ?", date),
		model.OrderFilter("date desc"),
		model.SwapAddress(swapAccount),
	)
	if err != nil {
		return errors.Wrap(err)
	}

	if len(swapCountKlines) == 0 {
		return nil
	}

	klineTyp.Date = swapCountKlines[0].Date

	prices := make([]*redis.Z, klineTyp.DataCount, klineTyp.DataCount)
	for index := range prices {
		prices[index] = &redis.Z{
			Score:  float64(klineTyp.SkipIntervalTime(-(klineTyp.DataCount - (index + 1))).Unix()),
			Member: nil,
		}
	}

	swapCountKlineMap := make(map[int64]*Price, klineTyp.DataCount)
	for _, v := range swapCountKlines {
		swapCountKlineMap[v.Date.Unix()] = &Price{
			Open:   v.Open,
			High:   v.High,
			Low:    v.Low,
			Settle: v.Settle,
			Avg:    v.Avg,
			Date:   v.Date,
		}
	}

	lastPrice := &Price{}
	for index, v := range prices {
		price, ok := swapCountKlineMap[int64(v.Score)]
		if ok {
			lastPrice = price
			prices[index].Score = v.Score
			prices[index].Member = price
		} else if lastPrice.Date != nil {
			prices[index].Score = v.Score
			prices[index].Member = &Price{
				Open:   lastPrice.Settle,
				High:   lastPrice.Settle,
				Low:    lastPrice.Settle,
				Settle: lastPrice.Settle,
				Avg:    lastPrice.Settle,
				Date:   klineTyp.SkipIntervalTime(-(klineTyp.DataCount - (index + 1))),
			}
		}
	}

	for i, v := range prices {
		if v.Member != nil {
			prices = prices[i:]
			break
		}
	}

	if isDelete {
		if len(prices) == 1 {
			// 如果没有数据改变则减少redis io
			if prices[0].Member.(*Price).Settle.Equal(lastValuePrice.Settle) && prices[0].Member.(*Price).Avg.Equal(lastValuePrice.Avg) {
				return nil
			}
		}
		// 最后一个数据会重复，提前删除，以便于更新
		if err = redisClient.ZRem(ctx, key, lastValue).Err(); err != nil {
			return errors.Wrap(err)
		}
	}

	if err = redisClient.ZAdd(context.TODO(), key, prices...).Err(); err != nil {
		logger.Error("sync swap account last 24h vol to redis err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	// 删除多余数据
	klineTyp.Date = prices[len(prices)-1].Member.(*Price).Date
	firstDate := klineTyp.SkipIntervalTime(-klineTyp.DataCount)
	if err = redisClient.ZRemRangeByScore(ctx, key, "", strconv.FormatInt(firstDate.Unix(), 10)).Err(); err != nil {
		logger.Error(" histogram about bol and tvl,deleting redundant data err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	return nil

}

func syncVolAndTvlHistogram() error {
	var (
		ctx = context.Background()
	)

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
			if err = syncDateTypeHistogram(ctx, v, swapConfig.SwapAccount); err != nil {
				logger.Error("sync histogram to redis err", logger.Errorv(err))
				return errors.Wrap(err)
			}
		}
	}

	return nil
}

func syncDateTypeHistogram(ctx context.Context, klineTyp KLineTyp, swapAccount string) error {
	var (
		key                    = domain.HistogramKey(klineTyp.DateType, swapAccount)
		date                   = &time.Time{}
		isDelete               = false
		lastValueSwapHistogram = &SwapHistogram{}
		lastValue              []string
		err                    error
	)

	lastValue, err = redisClient.ZRange(ctx, key, -1, -1).Result() // 返回最后一个元素
	if err != nil && !redisClient.ErrIsNil(err) {
		return errors.Wrap(err)
	}

	if len(lastValue) != 0 {
		isDelete = true
		if err = json.Unmarshal([]byte(lastValue[0]), lastValueSwapHistogram); err != nil {
			return errors.Wrap(err)
		}
		date = lastValueSwapHistogram.Date
	}

	swapCountKlines, err := model.QuerySwapCountKLines(ctx, klineTyp.DataCount, 0,
		model.NewFilter("date_type = ?", klineTyp.DateType),
		model.NewFilter("date >= ?", date),
		model.OrderFilter("date desc"),
		model.SwapAddress(swapAccount),
	)
	if err != nil {
		return errors.Wrap(err)
	}

	if len(swapCountKlines) == 0 {
		return nil
	}

	klineTyp.Date = swapCountKlines[0].Date

	swapHistograms := make([]*redis.Z, klineTyp.DataCount, klineTyp.DataCount)
	for index := range swapHistograms {
		swapHistograms[index] = &redis.Z{
			Score:  float64(klineTyp.SkipIntervalTime(-(klineTyp.DataCount - (index + 1))).Unix()),
			Member: nil,
		}
	}

	swapHistogramMap := make(map[int64]*SwapHistogram, klineTyp.DataCount)
	for _, v := range swapCountKlines {
		swapHistogramMap[v.Date.Unix()] = &SwapHistogram{
			Tvl:  v.TvlInUsd,
			Vol:  v.VolInUsd,
			Date: v.Date,
		}
	}

	lastHistogram := &SwapHistogram{}
	for index, v := range swapHistograms {
		price, ok := swapHistogramMap[int64(v.Score)]
		if ok {
			lastHistogram = price
			swapHistograms[index].Score = v.Score
			swapHistograms[index].Member = price
		} else if lastHistogram.Date != nil {
			swapHistograms[index].Score = v.Score
			swapHistograms[index].Member = &SwapHistogram{
				Tvl:  lastHistogram.Tvl,
				Vol:  lastHistogram.Vol,
				Date: klineTyp.SkipIntervalTime(-(klineTyp.DataCount - (index + 1))),
			}
		}
	}

	for i, v := range swapHistograms {
		if v.Member != nil {
			swapHistograms = swapHistograms[i:]
			break
		}
	}

	if isDelete {
		if len(swapHistograms) == 1 {
			// 如果没有数据改变则减少redis io
			if swapHistograms[0].Member.(*SwapHistogram).Tvl.Equal(lastValueSwapHistogram.Tvl) {
				return nil
			}
		}
		// 最后一个数据会重复，提前删除，以便于更新
		if err = redisClient.ZRem(ctx, key, lastValue).Err(); err != nil {
			return errors.Wrap(err)
		}
	}

	if err = redisClient.ZAdd(context.TODO(), key, swapHistograms...).Err(); err != nil {
		logger.Error("sync swap count histogram about bol and tvl to redis err", logger.Errorv(err))
		return errors.Wrap(err)
	}

	// 删除多余数据
	lastSwapHistogram := swapHistograms[len(swapHistograms)-1]
	klineTyp.Date = lastSwapHistogram.Member.(*SwapHistogram).Date
	firstDate := klineTyp.SkipIntervalTime(-klineTyp.DataCount)
	if err = redisClient.ZRemRangeByScore(ctx, key, "", strconv.FormatInt(firstDate.Unix(), 10)).Err(); err != nil {
		logger.Error(" histogram about bol and tvl,deleting redundant data err", logger.Errorv(err))
		return errors.Wrap(err)
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
				logger.Error("sync k line to redis err", logger.Errorv(err))
				return errors.Wrap(err)
			}
		}
	}
	return nil
}
