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

func syncKLineByDateType(ctx context.Context, klineT KLineTyp, swapAccount string) error {
	var (
		key = domain.KLineKey(klineT.DateType, swapAccount)
		err error
	)

	// 构造初始零值数据
	priceZ := make([]*PriceZ, klineT.DataCount, klineT.DataCount)
	for index := range priceZ {
		date := klineT.SkipIntervalTime(-(klineT.DataCount - (index + 1)))
		priceZ[index] = &PriceZ{
			Score: date.Unix(),
			Member: &Price{
				Date: date,
			},
		}
	}

	swapCountKlines, err := model.QuerySwapCountKLines(ctx, klineT.DataCount, 0,
		model.NewFilter("date_type = ?", klineT.DateType),
		model.OrderFilter("date desc"),
		model.SwapAddress(swapAccount),
	)

	if err != nil {
		return errors.Wrap(err)
	}

	if len(swapCountKlines) == 0 {
		return nil
	}

	// 转换成map，减少for循环
	PriceMap := make(map[int64]*Price, klineT.DataCount)
	for _, v := range swapCountKlines {
		PriceMap[v.Date.Unix()] = &Price{
			High:   v.High,
			Open:   v.Open,
			Low:    v.Low,
			Settle: v.Settle,
			Date:   v.Date,
		}
	}

	// 找到第一个数据
	lastPrice := &Price{}
	for _, v := range swapCountKlines {
		if v.Date.After(*priceZ[0].Member.Date) {
			break
		}
		lastPrice = &Price{
			Settle: v.Settle,
			Date:   v.Date,
		}
	}

	for index, v := range priceZ {
		price, ok := PriceMap[v.Score]
		if ok {
			lastPrice = price
			priceZ[index].Score = v.Score
			priceZ[index].Member.High = priceZ[index].Member.High.Add(price.High).Round(6)
			priceZ[index].Member.Open = priceZ[index].Member.Open.Add(price.Open).Round(6)
			priceZ[index].Member.Low = priceZ[index].Member.Low.Add(price.Low).Round(6)
			priceZ[index].Member.Settle = priceZ[index].Member.Settle.Add(price.Settle).Round(6)
			priceZ[index].Member.Avg = priceZ[index].Member.Avg.Add(price.Avg).Round(6)

		} else {
			priceZ[index].Score = v.Score
			priceZ[index].Member.High = priceZ[index].Member.High.Add(lastPrice.Settle).Round(6)
			priceZ[index].Member.Open = priceZ[index].Member.Open.Add(lastPrice.Settle).Round(6)
			priceZ[index].Member.Low = priceZ[index].Member.Low.Add(lastPrice.Settle).Round(6)
			priceZ[index].Member.Settle = priceZ[index].Member.Settle.Add(lastPrice.Settle).Round(6)
			priceZ[index].Member.Avg = priceZ[index].Member.Avg.Add(lastPrice.Settle).Round(6)
		}
	}

	// 去掉列表前面的零值
	for i, v := range priceZ {
		if !v.Member.Avg.IsZero() {
			priceZ = priceZ[i:]
			break
		}
	}

	// lua 通过脚本更新
	newZ := make([]interface{}, 0, len(priceZ)+1)
	for i := range priceZ {
		newZ = append(newZ, priceZ[i].Score)
		newZ = append(newZ, priceZ[i].Member)
	}

	_, err = delAndAddSZSet.Run(ctx, redisClient, []string{key}, newZ).Result()
	if err != nil {
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

// 采用redis zset 数据结构，先查询是否有数据存在，如果没有则同步全部数据，有则现获取已同步的数据的最后一条，然后同步新数据
func syncKLineToRedis() error {
	var (
		ctx = context.Background()
		now = time.Now()
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
			v.Date = &now
			if err = syncKLineByDateType(ctx, v, swapConfig.SwapAccount); err != nil {
				logger.Error("sync k line to redis err", logger.Errorv(err))
				return errors.Wrap(err)
			}
		}
	}
	return nil
}

func sumTotalSwapAccount() error {
	var (
		ctx = context.Background()
		now = time.Now()
	)
	// 获取时间类型
	kLines, err := model.QuerySwapCountKLines(ctx, 1, 0, model.IDDESCFilter())
	if err != nil {
		logger.Error("query swap_transactions err", logger.Errorv(err))
		return err
	}

	if len(kLines) == 0 {
		return nil
	}

	for _, v := range []KLineTyp{DateMin, DateTwelfth, DateQuarter, DateHalfAnHour, DateHour, DateDay, DateWek, DateMon} {
		date := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, kLines[0].Date.Location())
		v.Date = &date
		if err := sumDateTypeSwapAccount(ctx, v); err != nil {
			logger.Error("sync k line to redis err", logger.Errorv(err))
			return errors.Wrap(err)
		}
	}

	return nil
}

func sumDateTypeSwapAccount(ctx context.Context, klineT KLineTyp) error {

	var (
		key = domain.TotalHistogramKey(klineT.DateType)
	)

	// 构造初始零值数据
	swapHistogramZ := make([]*HistogramZ, klineT.DataCount, klineT.DataCount)
	for index := range swapHistogramZ {
		date := klineT.SkipIntervalTime(-(klineT.DataCount - (index + 1)))
		swapHistogramZ[index] = &HistogramZ{
			Score: date.Unix(),
			Member: &SwapHistogram{
				Tvl:  decimal.Decimal{},
				Vol:  decimal.Decimal{},
				Date: date,
			},
		}
	}

	for _, swapConfig := range sol.SwapConfigList() {

		swapCountKlines, err := model.QuerySwapCountKLines(ctx, klineT.DataCount, 0,
			model.NewFilter("date_type = ?", klineT.DateType),
			model.OrderFilter("date desc"),
			model.SwapAddress(swapConfig.SwapAccount),
		)

		if err != nil {
			return errors.Wrap(err)
		}

		if len(swapCountKlines) == 0 {
			continue
		}

		// 转换成map，减少for循环
		swapHistogramMap := make(map[int64]*SwapHistogram, klineT.DataCount)
		for _, v := range swapCountKlines {
			swapHistogramMap[v.Date.Unix()] = &SwapHistogram{
				Tvl:  v.TvlInUsd,
				Vol:  v.VolInUsd,
				Date: v.Date,
			}
		}

		// 找到第一个数据
		lastHistogram := &SwapHistogram{}
		for _, v := range swapCountKlines {
			if v.Date.After(*swapHistogramZ[0].Member.Date) {
				break
			}
			lastHistogram = &SwapHistogram{
				Tvl:  v.TvlInUsd,
				Vol:  v.VolInUsd,
				Date: v.Date,
			}
		}

		for index, v := range swapHistogramZ {
			price, ok := swapHistogramMap[v.Score]
			if ok {
				lastHistogram = price
				swapHistogramZ[index].Score = v.Score
				swapHistogramZ[index].Member.Tvl = swapHistogramZ[index].Member.Tvl.Add(price.Tvl).Round(6)
				swapHistogramZ[index].Member.Vol = swapHistogramZ[index].Member.Vol.Add(price.Vol).Round(6)
			} else {
				swapHistogramZ[index].Score = v.Score
				swapHistogramZ[index].Member.Tvl = swapHistogramZ[index].Member.Tvl.Add(lastHistogram.Tvl).Round(6)
			}
		}

	}

	// 去掉列表前面的零值
	for i, v := range swapHistogramZ {
		if !v.Member.Tvl.IsZero() {
			swapHistogramZ = swapHistogramZ[i:]
			break
		}
	}

	// lua 通过脚本更新
	newZ := make([]interface{}, 0, len(swapHistogramZ)+1)
	for i := range swapHistogramZ {
		newZ = append(newZ, swapHistogramZ[i].Score)
		newZ = append(newZ, swapHistogramZ[i].Member)
	}

	_, err := delAndAddSZSet.Run(ctx, redisClient, []string{key}, newZ).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
