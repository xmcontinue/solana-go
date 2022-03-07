package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/transport/rpcx"
	"github.com/go-redis/redis/v8"

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
	)

	// 最后一个数据和前端给的时间比较
	lastKLine, err := t.redisClient.ZRange(ctx, key, -1, -1).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	if len(lastKLine) == 0 {
		return nil
	}

	price := &process.Price{}
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

	// 要求时间大于存入redis的时间，则返回最后limit条数据
	if price.Date.Before(args.EndTime) {

		if args.Offset == 0 {
			offset = -1
		} else {
			offset = -int64(args.Offset)
		}

		values, err := t.redisClient.ZRange(ctx, key, -int64(args.Limit+args.Offset), offset).Result()
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

		total, err := t.redisClient.ZCount(ctx, key, "", strconv.FormatInt(args.EndTime.Unix(), 10)).Result()
		if err != nil {
			return errors.Wrap(err)
		}

		reply.Limit = limit(args.Limit)
		reply.Offset = args.Offset
		reply.List = list
		reply.Total = total
		return nil
	}

	// 要求时间比最后一个时间要早
	// ZSet 中第一个数据的score
	scoreWithScores := &redis.ZRangeBy{
		Min:    "",
		Max:    strconv.FormatInt(args.EndTime.Unix(), 10),
		Offset: 0,
		Count:  1,
	}
	firstValue, err := t.redisClient.ZRangeByScoreWithScores(ctx, key, scoreWithScores).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	if len(firstValue) == 0 {
		return nil
	}

	endTime := dateType.GetDate()
	beginTime := dateType.GetDate().Add(dateType.TimeInterval * (-time.Duration(args.Limit + args.Offset + 1)))

	opt := &redis.ZRangeBy{
		Min:    strconv.FormatInt(beginTime.Unix(), 10),
		Max:    strconv.FormatInt(endTime.Unix(), 10),
		Offset: int64(args.Offset),
		Count:  int64(args.Limit),
	}
	var values []redis.Z

	for {
		values, err = t.redisClient.ZRangeByScoreWithScores(ctx, key, opt).Result()
		if err != nil {
			return errors.Wrap(err)
		}

		if len(values) == 0 {
			if float64(beginTime.Unix()) < firstValue[0].Score {
				break
			}
		}

		if float64(beginTime.Unix()) < firstValue[0].Score {
			break
		}

		beginTime = beginTime.Add(dateType.TimeInterval * (-time.Duration(args.Limit + args.Offset + 1)))
		opt.Min = strconv.FormatInt(beginTime.Add(dateType.TimeInterval*(-time.Duration(args.Limit+args.Offset+1))).Unix(), 10)

	}
	total, err := t.redisClient.ZCount(ctx, key, "", strconv.FormatInt(endTime.Unix(), 10)).Result()
	if err != nil {
		return errors.Wrap(err)
	}

	for i := range list {
		date := dateType.GetDate().Add(-dateType.TimeInterval * time.Duration(i))
		list[len(list)-(i+1)] = &process.Price{
			Date: &date,
		}
	}

	for _, v := range values {
		_ = json.Unmarshal([]byte(v.Member.(string)), price)
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

	reply.Limit = limit(args.Limit)
	reply.Offset = args.Offset
	reply.List = list
	reply.Total = total

	return nil
}
