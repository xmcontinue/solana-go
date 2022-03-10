package process

import (
	"encoding/json"
	"reflect"
	"time"

	"git.cplus.link/go/akit/util/decimal"

	"git.cplus.link/crema/backend/pkg/domain"
)

var (
	DateMin = KLineTyp{
		BeforeIntervalDateType: domain.DateNone,
		DateType:               domain.DateMin,
		TimeInterval:           time.Minute,
		DataCount:              60 * 5,
	}

	DateTwelfth = KLineTyp{
		BeforeIntervalDateType: domain.DateMin,
		DateType:               domain.DateTwelfth,
		Interval:               5,
		InnerTimeInterval:      time.Minute,
		TimeInterval:           time.Minute * 5,
		DataCount:              12 * 24,
	}

	DateQuarter = KLineTyp{
		BeforeIntervalDateType: domain.DateTwelfth,
		DateType:               domain.DateQuarter,
		Interval:               3,
		InnerTimeInterval:      time.Minute * 5,
		TimeInterval:           time.Minute * 15,
		DataCount:              4 * 24,
	}

	DateHalfAnHour = KLineTyp{
		BeforeIntervalDateType: domain.DateQuarter,
		DateType:               domain.DateHalfAnHour,
		Interval:               2,
		InnerTimeInterval:      time.Minute * 15,
		TimeInterval:           time.Minute * 30,
		DataCount:              2 * 24,
	}

	DateHour = KLineTyp{
		BeforeIntervalDateType: domain.DateHalfAnHour,
		DateType:               domain.DateHour,
		Interval:               2,
		InnerTimeInterval:      time.Minute * 30,
		TimeInterval:           time.Hour,
		DataCount:              2 * 24,
	}

	DateDay = KLineTyp{
		BeforeIntervalDateType: domain.DateHour,
		DateType:               domain.DateDay,
		Interval:               24,
		InnerTimeInterval:      time.Hour,
		TimeInterval:           time.Hour * 24,
		DataCount:              30,
	}

	DateWek = KLineTyp{
		BeforeIntervalDateType: domain.DateDay,
		DateType:               domain.DateWek,
		Interval:               7,
		InnerTimeInterval:      time.Hour * 24,
		TimeInterval:           time.Hour * 24 * 7,
		DataCount:              54,
	}

	DateMon = KLineTyp{
		BeforeIntervalDateType: domain.DateDay,
		DateType:               domain.DateMon,
		Interval:               31,
		InnerTimeInterval:      time.Hour * 24,
		DataCount:              24,
	}
)

type KLineTyp struct {
	Date                   *time.Time
	DateType               domain.DateType
	BeforeIntervalDateType domain.DateType // 前一个间隔类型
	Interval               int             // 相较于前一个时间段，用多少前一个时间段可以填满当前时间段
	InnerTimeInterval      time.Duration   // 每个当前时间间隔内部最小单位的间隔
	TimeInterval           time.Duration   // 当前时间类型间隔
	DataCount              int             // 数据量，存入redis的数据大小
}

func (m *KLineTyp) Name() domain.DateType {
	return m.DateType
}

// 最近一个时间点，并替换原来的值
func (m *KLineTyp) GetDate() *time.Time {
	var date time.Time
	if m.DateType == domain.DateMin {
		date = time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), m.Date.Hour(), m.Date.Minute(), 0, 0, m.Date.Location())
	} else if m.DateType == domain.DateTwelfth {
		exactDiv := m.Date.Minute() / 5
		date = time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), m.Date.Hour(), exactDiv*5, 0, 0, m.Date.Location())
	} else if m.DateType == domain.DateQuarter {
		// 计算时间所在的四分之一 时间区间
		exactDiv := m.Date.Minute() / 15
		date = time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), m.Date.Hour(), exactDiv*15, 0, 0, m.Date.Location())
	} else if m.DateType == domain.DateHalfAnHour {
		exactDiv := m.Date.Minute() / 30
		date = time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), m.Date.Hour(), exactDiv*30, 0, 0, m.Date.Location())
	} else if m.DateType == domain.DateHour {
		date = time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), m.Date.Hour(), 0, 0, 0, m.Date.Location())
	} else if m.DateType == domain.DateDay {
		date = time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), 0, 0, 0, 0, m.Date.Location())
	} else if m.DateType == domain.DateWek {
		// 计算时间所在周的第一天的日期，周日对应的是0，周一到周六对应1，2，3，4，5，6
		innerDate := time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), 0, 0, 0, 0, m.Date.Location())
		if innerDate.Weekday() == time.Monday {
			date = innerDate
		} else {
			offset := int(time.Monday - innerDate.Weekday())
			if offset > 0 {
				offset = -6
			}
			date = innerDate.AddDate(0, 0, offset)
		}
	} else if m.DateType == domain.DateMon {
		firstDateTime := m.Date.AddDate(0, 0, -m.Date.Day()+1)
		date = time.Date(firstDateTime.Year(), firstDateTime.Month(), firstDateTime.Day(), 0, 0, 0, 0, firstDateTime.Location())
	}

	return &date
}

// if skip >0 向后跳，skip<0 向前跳
func (m *KLineTyp) SkipIntervalTime(skip int) *time.Time {
	var date time.Time
	if m.Date.IsZero() {
		return &time.Time{}
	}
	if m.DateType == domain.DateMin {
		date = time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), m.Date.Hour(), m.Date.Minute(), 0, 0, m.Date.Location()).Add(m.TimeInterval * time.Duration(skip))
	} else if m.DateType == domain.DateTwelfth {
		exactDiv := m.Date.Minute() / 5
		date = time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), m.Date.Hour(), exactDiv*5, 0, 0, m.Date.Location()).Add(m.TimeInterval * time.Duration(skip))
	} else if m.DateType == domain.DateQuarter {
		// 计算时间所在的四分之一 时间区间
		exactDiv := m.Date.Minute() / 15
		date = time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), m.Date.Hour(), exactDiv*15, 0, 0, m.Date.Location()).Add(m.TimeInterval * time.Duration(skip))
	} else if m.DateType == domain.DateHalfAnHour {
		exactDiv := m.Date.Minute() / 30
		date = time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), m.Date.Hour(), exactDiv*30, 0, 0, m.Date.Location()).Add(m.TimeInterval * time.Duration(skip))
	} else if m.DateType == domain.DateHour {
		date = time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), m.Date.Hour(), 0, 0, 0, m.Date.Location()).Add(m.TimeInterval * time.Duration(skip))
	} else if m.DateType == domain.DateDay {
		date = time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), 0, 0, 0, 0, m.Date.Location()).Add(m.TimeInterval * time.Duration(skip))
	} else if m.DateType == domain.DateWek {
		// 计算时间所在周的第一天的日期，周日对应的是0，周一到周六对应1，2，3，4，5，6
		innerDate := time.Date(m.Date.Year(), m.Date.Month(), m.Date.Day(), 0, 0, 0, 0, m.Date.Location()).Add(m.TimeInterval * time.Duration(skip))
		if innerDate.Weekday() == time.Monday {
			date = innerDate
		} else {
			offset := int(time.Monday - innerDate.Weekday())
			if offset > 0 {
				offset = -6
			}
			date = innerDate.AddDate(0, 0, offset).Add(m.TimeInterval * time.Duration(-skip))
		}
	} else if m.DateType == domain.DateMon {
		firstDateTime := m.Date.AddDate(0, 0, -m.Date.Day()+1)
		date = time.Date(firstDateTime.Year(), firstDateTime.Month(), firstDateTime.Day(), 0, 0, 0, 0, firstDateTime.Location()).AddDate(0, skip, 0)
	}

	return &date
}

type Price struct {
	Open   decimal.Decimal
	High   decimal.Decimal
	Low    decimal.Decimal
	Settle decimal.Decimal
	Avg    decimal.Decimal
	Date   *time.Time
}

func (s *Price) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Price) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

type SwapHistogram struct {
	Tvl  decimal.Decimal
	Vol  decimal.Decimal
	Date *time.Time
}

type SwapHistogramPrice struct {
	Price decimal.Decimal
	Date  *time.Time
}

func (s *SwapHistogram) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *SwapHistogram) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *SwapHistogram) IsEmpty() bool {
	return reflect.DeepEqual(s, SwapHistogram{})
}