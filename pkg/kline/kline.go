package kline

import (
	"time"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/pkg/domain"
)

type Kline struct {
	Date  *time.Time
	Types []*Type
}

type Type struct {
	Date                   *time.Time
	DateType               domain.DateType
	BeforeIntervalDateType domain.DateType // 前一个间隔类型
	Interval               int             // 相较于前一个时间段，用多少前一个时间段可以填满当前时间段
	InnerTimeInterval      time.Duration   // 每个当前时间间隔内部最小单位的间隔
	TimeInterval           time.Duration   // 当前时间类型间隔
	DataCount              int             // 数据量，存入redis的数据大小
}

type InterTime struct {
	Date      time.Time
	Avg       decimal.Decimal
	TokenAUSD decimal.Decimal
	TokenBUSD decimal.Decimal
}

func NewKline(date *time.Time) *Kline {
	if date.IsZero() {
		temp := time.Now().UTC()
		date = &temp
	}

	k := &Kline{
		Date: date,
		Types: []*Type{
			{
				BeforeIntervalDateType: domain.DateNone,
				DateType:               domain.DateMin,
				TimeInterval:           time.Minute,
				DataCount:              500,
			}, {
				BeforeIntervalDateType: domain.DateMin,
				DateType:               domain.DateHour,
				Interval:               2,
				InnerTimeInterval:      time.Minute * 30,
				TimeInterval:           time.Hour,
				DataCount:              500,
			}, {
				BeforeIntervalDateType: domain.DateHour,
				DateType:               domain.DateDay,
				Interval:               24,
				InnerTimeInterval:      time.Hour,
				TimeInterval:           time.Hour * 24,
				DataCount:              500,
			}, {
				BeforeIntervalDateType: domain.DateDay,
				DateType:               domain.DateWek,
				Interval:               7,
				InnerTimeInterval:      time.Hour * 24,
				TimeInterval:           time.Hour * 24 * 7,
				DataCount:              500,
			}, {
				BeforeIntervalDateType: domain.DateDay,
				DateType:               domain.DateMon,
				Interval:               31,
				InnerTimeInterval:      time.Hour * 24,
				DataCount:              500,
			},
		},
	}

	k.setTypeDateForTime()

	return k
}

// setTypeDateForTime 根据当前时间获取间隔时间
func (k *Kline) setTypeDateForTime() {

	for _, t := range k.Types {
		var data time.Time
		if t.DateType == domain.DateMin {
			data = time.Date(k.Date.Year(), k.Date.Month(), k.Date.Day(), k.Date.Hour(), k.Date.Minute(), 0, 0, k.Date.Location())
		} else if t.DateType == domain.DateTwelfth {
			exactDiv := k.Date.Minute() / 5
			data = time.Date(k.Date.Year(), k.Date.Month(), k.Date.Day(), k.Date.Hour(), exactDiv*5, 0, 0, k.Date.Location())
		} else if t.DateType == domain.DateQuarter {
			// 计算时间所在的四分之一 时间区间
			exactDiv := k.Date.Minute() / 15
			data = time.Date(k.Date.Year(), k.Date.Month(), k.Date.Day(), k.Date.Hour(), exactDiv*15, 0, 0, k.Date.Location())
		} else if t.DateType == domain.DateHalfAnHour {
			exactDiv := k.Date.Minute() / 30
			data = time.Date(k.Date.Year(), k.Date.Month(), k.Date.Day(), k.Date.Hour(), exactDiv*30, 0, 0, k.Date.Location())
		} else if t.DateType == domain.DateHour {
			data = time.Date(k.Date.Year(), k.Date.Month(), k.Date.Day(), k.Date.Hour(), 0, 0, 0, k.Date.Location())
		} else if t.DateType == domain.DateDay {
			data = time.Date(k.Date.Year(), k.Date.Month(), k.Date.Day(), 0, 0, 0, 0, k.Date.Location())
		} else if t.DateType == domain.DateWek {
			data = k.getWeekFirstDay()
		} else if t.DateType == domain.DateMon {
			firstDateTime := k.Date.AddDate(0, 0, -k.Date.Day()+1)
			data = time.Date(firstDateTime.Year(), firstDateTime.Month(), firstDateTime.Day(), 0, 0, 0, 0, firstDateTime.Location())
		}

		t.Date = &data
	}
}

// getWeekFirstDay 获取给定时间所在周的第一天
func (k *Kline) getWeekFirstDay() time.Time {
	var date time.Time
	// 计算时间所在周的第一天的日期，周日对应的是0，周一到周六对应1，2，3，4，5，6
	innerDate := time.Date(k.Date.Year(), k.Date.Month(), k.Date.Day(), 0, 0, 0, 0, k.Date.Location())
	if innerDate.Weekday() == time.Monday {
		date = innerDate
	} else {
		offset := int(time.Monday - innerDate.Weekday())
		if offset > 0 {
			offset = -6
		}
		date = innerDate.AddDate(0, 0, offset)
	}
	return date
}

// CalculateAvg 按照上一个周期计算平均值，month除外（按照天计算）
func (m *Type) CalculateAvg(f func(time.Time, *[]*InterTime) error) (*InterTime, error) {

	var (
		count        = int32(0)
		sumAvg       = decimal.Zero
		sumTokenAUSD = decimal.Zero
		sumTokenBUSD = decimal.Zero
		beginTime    time.Time
		endTime      time.Time
	)

	avgList := make([]*InterTime, m.Interval, m.Interval)

	beginTime = *m.Date
	if m.DateType == domain.DateMon {
		lastDateTime := m.Date.AddDate(0, 1, -m.Date.Day())
		endTime = time.Date(lastDateTime.Year(), lastDateTime.Month(), lastDateTime.Day(), 0, 0, 0, 0, m.Date.Location()).Add(time.Hour * 24)
	} else {
		endTime = beginTime.Add(m.InnerTimeInterval * time.Duration(m.Interval))
	}

	for index := range avgList {
		avgList[index] = &InterTime{
			Date: m.Date.Add(m.InnerTimeInterval * time.Duration(index)),
		}
	}

	err := f(endTime, &avgList)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// calculate avg
	for _, v := range avgList {
		if !v.Avg.IsZero() {
			sumAvg = sumAvg.Add(v.Avg)
			sumTokenAUSD = sumTokenAUSD.Add(v.TokenAUSD)
			sumTokenBUSD = sumTokenBUSD.Add(v.TokenBUSD)
			count++
		}
	}

	if count == 0 {
		return nil, errors.RecordNotFound
	}

	return &InterTime{
		Avg:       sumAvg.Div(decimal.NewFromInt32(count)),
		TokenAUSD: sumTokenAUSD.Div(decimal.NewFromInt32(count)),
		TokenBUSD: sumTokenBUSD.Div(decimal.NewFromInt32(count)),
	}, nil

	//return sumAvg.Div(decimal.NewFromInt32(count)), nil
}
