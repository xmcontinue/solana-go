package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"

	"git.cplus.link/crema/backend/internal/worker/process"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/iface"
)

func (t *MarketService) GetKline(ctx context.Context, args *iface.GetKlineReq, reply *iface.GetKlineResp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}

	var (
		key      = domain.KLineKey(args.DateType, args.SwapAccount)
		dateType = &process.KLineTyp{}
		offset   = int64(0)
		list     = make([]*process.Price, args.Limit, args.Limit)
		price    = &process.Price{}
	)

	// 最后一个数据和前端给的时间比较
	lastKLine, err := t.redisClient.ZRange(ctx, key, -1, -1).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	if len(lastKLine) == 0 {
		return nil
	}

	_ = json.Unmarshal([]byte(lastKLine[0]), price)

	// 构造时间
	for _, v := range []process.KLineTyp{process.DateMin, process.DateTwelfth, process.DateQuarter, process.DateHalfAnHour, process.DateHour, process.DateDay, process.DateWek, process.DateMon} {
		if v.DateType == args.DateType {
			dateType = &v
			dateType.Date = price.Date
			dateType.InnerTimeInterval = v.InnerTimeInterval
			break
		}
	}

	if args.Offset == 0 {
		offset = -1
	} else {
		offset = -int64(args.Offset)
	}

	values, err := t.redisClient.ZRange(ctx, key, -int64(args.Limit)+offset, offset).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	if len(values) == 0 {
		return nil
	}

	_ = json.Unmarshal([]byte(values[len(values)-1]), price)

	for i := range list {
		date := dateType.GetDate().Add(-dateType.TimeInterval * time.Duration(i))
		list[len(list)-(i+1)] = &process.Price{
			Date: &date,
		}
	}

	for _, v := range values {
		_ = json.Unmarshal([]byte(v), price)
		for i, l := range list {
			if l.Date.After(*price.Date) || l.Date.Equal(*price.Date) {
				list[i].Open = price.Open
				list[i].High = price.High
				list[i].Low = price.Low
				list[i].Settle = price.Settle
				list[i].Avg = price.Avg
			}
		}
	}

	total, err := t.redisClient.ZCount(ctx, key, "", strconv.FormatInt(time.Now().Unix(), 10)).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	reply.Limit = limit(args.Limit)
	reply.Offset = args.Offset
	reply.List = list
	reply.Total = total
	return nil
}

func (t *MarketService) GetHistogram(ctx context.Context, args *iface.GetHistogramReq, reply *iface.GetHistogramResp) error {
	defer rpcx.Recover(ctx)
	if err := validate(args); err != nil {
		return errors.Wrapf(errors.ParameterError, "validate:%v", err)
	}

	var (
		key           = domain.HistogramKey(args.DateType, args.SwapAccount)
		dateType      = &process.KLineTyp{}
		offset        = int64(0)
		list          = make([]*process.SwapHistogramPrice, 0, args.Limit)
		swapHistogram = &process.SwapHistogram{}
	)

	// 获取最后一条数据
	lastKLine, err := t.redisClient.ZRange(ctx, key, -1, -1).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	if len(lastKLine) == 0 {
		return nil
	}

	_ = json.Unmarshal([]byte(lastKLine[0]), swapHistogram)

	// 构造时间
	for _, v := range []process.KLineTyp{process.DateMin, process.DateTwelfth, process.DateQuarter, process.DateHalfAnHour, process.DateHour, process.DateDay, process.DateWek, process.DateMon} {
		if v.DateType == args.DateType {
			dateType = &v
			dateType.Date = swapHistogram.Date
			dateType.InnerTimeInterval = v.InnerTimeInterval
			break
		}
	}

	if args.Offset == 0 {
		offset = -1
	} else {
		offset = -int64(args.Offset)
	}

	values, err := t.redisClient.ZRange(ctx, key, -int64(args.Limit)+offset, offset).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	if len(values) == 0 {
		return nil
	}

	_ = json.Unmarshal([]byte(values[len(values)-1]), swapHistogram)

	for i := range values {
		innerSwapHistogram := &process.SwapHistogram{}
		_ = json.Unmarshal([]byte(values[i]), innerSwapHistogram)

		if args.Typ == "vol" {
			list = append(list, &process.SwapHistogramPrice{
				Price: innerSwapHistogram.Vol,
				Date:  innerSwapHistogram.Date,
			})
		} else {
			list = append(list, &process.SwapHistogramPrice{
				Price: swapHistogram.Vol,
				Date:  swapHistogram.Date,
			})
		}
	}

	total, err := t.redisClient.ZCount(ctx, key, "", strconv.FormatInt(time.Now().Unix(), 10)).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	reply.Limit = limit(args.Limit)
	reply.Offset = args.Offset
	reply.List = list
	reply.Total = total
	return nil
}
