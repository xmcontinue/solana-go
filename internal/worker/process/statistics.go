package process

import (
	"context"
	"encoding/json"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
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
		// tokenAPrice, tokenBPrice := coingecko.GetPriceForTokenAccount(v.TokenAAddress), coingecko.GetPriceForTokenAccount(v.TokenBAddress)
		tokenAPrice, tokenBPrice := decimal.NewFromInt(1), decimal.NewFromInt(1)
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
	for index := range swapCountKlines {
		if swapCountKlines[len(swapCountKlines)-index-1].Date.After(*priceZ[0].Member.Date) {
			break
		}
		lastPrice = &Price{
			Settle: swapCountKlines[len(swapCountKlines)-index-1].Settle,
			Date:   swapCountKlines[len(swapCountKlines)-index-1].Date,
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

	_, err = delAndAddByZSet.Run(ctx, redisClient, []string{key}, newZ).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}

func syncVolAndTvlHistogram() error {
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
			if err = syncSwapAccountVolAndTvlByDateType(ctx, v, swapConfig.SwapAccount); err != nil {
				logger.Error("sync histogram to redis err", logger.Errorv(err))
				return errors.Wrap(err)
			}
		}
	}

	return nil
}

func syncSwapAccountVolAndTvlByDateType(ctx context.Context, klineT KLineTyp, swapAccount string) error {
	var (
		key = domain.HistogramKey(klineT.DateType, swapAccount)
	)

	// 构造初始零值数据
	singleSwapHistogramZ := make([]*HistogramZ, klineT.DataCount, klineT.DataCount)
	for index := range singleSwapHistogramZ {
		date := klineT.SkipIntervalTime(-(klineT.DataCount - (index + 1)))
		singleSwapHistogramZ[index] = &HistogramZ{
			Score: date.Unix(),
			Member: &SwapHistogram{
				Tvl:  decimal.Decimal{},
				Vol:  decimal.Decimal{},
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
	for index := range swapCountKlines {
		if swapCountKlines[len(swapCountKlines)-index-1].Date.After(*singleSwapHistogramZ[0].Member.Date) {
			break
		}
		lastHistogram = &SwapHistogram{
			Tvl:  swapCountKlines[len(swapCountKlines)-index-1].TvlInUsd,
			Vol:  swapCountKlines[len(swapCountKlines)-index-1].VolInUsd,
			Date: swapCountKlines[len(swapCountKlines)-index-1].Date,
		}
	}

	for index, v := range singleSwapHistogramZ {
		price, ok := swapHistogramMap[v.Score]
		if ok {
			lastHistogram = price
			singleSwapHistogramZ[index].Score = v.Score
			singleSwapHistogramZ[index].Member.Tvl = singleSwapHistogramZ[index].Member.Tvl.Add(price.Tvl).Round(6)
			singleSwapHistogramZ[index].Member.Vol = singleSwapHistogramZ[index].Member.Vol.Add(price.Vol).Round(6)
		} else {
			singleSwapHistogramZ[index].Score = v.Score
			singleSwapHistogramZ[index].Member.Tvl = singleSwapHistogramZ[index].Member.Tvl.Add(lastHistogram.Tvl).Round(6)
		}
	}

	// 去掉列表前面的零值
	for i, v := range singleSwapHistogramZ {
		if !v.Member.Tvl.IsZero() {
			singleSwapHistogramZ = singleSwapHistogramZ[i:]
			break
		}
	}

	// lua 通过脚本更新
	newZ := make([]interface{}, 0, len(singleSwapHistogramZ)+1)
	for i := range singleSwapHistogramZ {
		newZ = append(newZ, singleSwapHistogramZ[i].Score)
		newZ = append(newZ, singleSwapHistogramZ[i].Member)
	}

	_, err = delAndAddByZSet.Run(ctx, redisClient, []string{key}, newZ).Result()
	if err != nil {
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
		if err := sumDateTypeSwapAccount(ctx, v, &now); err != nil {
			logger.Error("sync k line to redis err", logger.Errorv(err))
			return errors.Wrap(err)
		}
	}

	return nil
}

func sumDateTypeSwapAccount(ctx context.Context, klineT KLineTyp, now *time.Time) error {

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
			model.SwapAddress(swapConfig.SwapAccount),
			model.NewFilter("date_type = ?", klineT.DateType),
			model.OrderFilter("date desc"),
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
				Tvl:  v.TokenABalance.Mul(v.TokenAUSDForContract).Add(v.TokenBBalance.Mul(v.TokenBUSDForContract)),
				Vol:  v.TokenAVolume.Mul(v.TokenAUSDForContract).Add(v.TokenBVolume.Mul(v.TokenBUSDForContract)),
				Date: v.Date,
			}
		}

		// 找到第一个数据
		lastHistogram := &SwapHistogram{}
		for index := range swapCountKlines {
			if swapCountKlines[len(swapCountKlines)-index-1].Date.After(*swapHistogramZ[0].Member.Date) {
				break
			}
			v := swapCountKlines[len(swapCountKlines)-index-1]
			lastHistogram = &SwapHistogram{
				Tvl:  v.TokenABalance.Mul(v.TokenAUSDForContract).Add(v.TokenBBalance.Mul(v.TokenBUSDForContract)),
				Vol:  v.TokenAVolume.Mul(v.TokenAUSDForContract).Add(v.TokenBVolume.Mul(v.TokenBUSDForContract)),
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

	// 注释，暂时不使用当前时间点这个坐标
	// if len(swapHistogramZ) > 1 {
	//	swapHistogramZ = append(swapHistogramZ, &HistogramZ{
	//		Score: now.Unix(),
	//		Member: &SwapHistogram{
	//			Tvl:  swapHistogramZ[len(swapHistogramZ)-1].Member.Tvl,
	//			Vol:  swapHistogramZ[len(swapHistogramZ)-1].Member.Vol,
	//			Date: now,
	//		},
	//	})
	//
	//	// 时间计时改为每个时间段的最后时间点，最后一个时间段就前一个时间段到当前的时间，就用英文current 表示
	//	for i := 0; i < len(swapHistogramZ)-1; i++ {
	//		swapHistogramZ[len(swapHistogramZ)-1-i].Member.Tvl = swapHistogramZ[len(swapHistogramZ)-2-i].Member.Tvl
	//		swapHistogramZ[len(swapHistogramZ)-1-i].Member.Vol = swapHistogramZ[len(swapHistogramZ)-2-i].Member.Vol
	//	}
	//
	//	swapHistogramZ[0].Member.Tvl = decimal.Zero
	//	swapHistogramZ[0].Member.Vol = decimal.Zero
	// }

	// lua 通过脚本更新
	newZ := make([]interface{}, 0, len(swapHistogramZ)+1)
	for i := range swapHistogramZ {
		newZ = append(newZ, swapHistogramZ[i].Score)
		newZ = append(newZ, swapHistogramZ[i].Member)
	}

	_, err := delAndAddByZSet.Run(ctx, redisClient, []string{key}, newZ).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	return nil
}
