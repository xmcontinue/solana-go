package process

import (
	"context"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/chain/sol"
	model "git.cplus.link/crema/backend/internal/model/market"
	"git.cplus.link/crema/backend/pkg/domain"
)

var (
	swapHistogramZ = make([]*HistogramZ, 500+1, 500+1)
	newZ           = make([]interface{}, 0, 2*len(swapHistogramZ))
)

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
	//swapHistogramZ = swapHistogramZ[:0]
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
		swapHistogramZ[index+1] = &HistogramZ{
			Score: 0,
			Member: &SwapHistogram{
				Tvl:  decimal.Decimal{},
				Vol:  decimal.Decimal{},
				Date: nil,
			},
		}
	}

	for _, swapConfig := range sol.SwapConfigList() {

		swapCountKlines, err := model.QuerySwapCountKLines(ctx, klineT.DataCount, 0,
			model.SwapAddressFilter(swapConfig.SwapAccount),
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
	newZ = newZ[:0]
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
