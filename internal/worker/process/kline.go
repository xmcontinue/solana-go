package process

import (
	"time"

	"git.cplus.link/crema/backend/pkg/domain"
)

var (
	DateMin = KLineTyp{
		BeforeIntervalDateType: domain.DateNone,
		DateType:               domain.DateMin,
		TimeInterval:           time.Minute,
		DataCount:              500,
	}

	DateTwelfth = KLineTyp{
		BeforeIntervalDateType: domain.DateMin,
		DateType:               domain.DateTwelfth,
		Interval:               5,
		InnerTimeInterval:      time.Minute,
		TimeInterval:           time.Minute * 5,
		DataCount:              500,
	}

	DateQuarter = KLineTyp{
		BeforeIntervalDateType: domain.DateTwelfth,
		DateType:               domain.DateQuarter,
		Interval:               3,
		InnerTimeInterval:      time.Minute * 5,
		TimeInterval:           time.Minute * 15,
		DataCount:              500,
	}

	DateHalfAnHour = KLineTyp{
		BeforeIntervalDateType: domain.DateQuarter,
		DateType:               domain.DateHalfAnHour,
		Interval:               2,
		InnerTimeInterval:      time.Minute * 15,
		TimeInterval:           time.Minute * 30,
		DataCount:              500,
	}

	DateHour = KLineTyp{
		BeforeIntervalDateType: domain.DateHalfAnHour,
		DateType:               domain.DateHour,
		Interval:               2,
		InnerTimeInterval:      time.Minute * 30,
		TimeInterval:           time.Hour,
		DataCount:              500,
	}

	DateDay = KLineTyp{
		BeforeIntervalDateType: domain.DateHour,
		DateType:               domain.DateDay,
		Interval:               24,
		InnerTimeInterval:      time.Hour,
		TimeInterval:           time.Hour * 24,
		DataCount:              500,
	}

	DateWek = KLineTyp{
		BeforeIntervalDateType: domain.DateDay,
		DateType:               domain.DateWek,
		Interval:               7,
		InnerTimeInterval:      time.Hour * 24,
		TimeInterval:           time.Hour * 24 * 7,
		DataCount:              500,
	}

	DateMon = KLineTyp{
		BeforeIntervalDateType: domain.DateDay,
		DateType:               domain.DateMon,
		Interval:               31,
		InnerTimeInterval:      time.Hour * 24,
		DataCount:              500,
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

// GetDate 获取最近的时间类型的时间点
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
		date = m.getWeekFirstDay()
	} else if m.DateType == domain.DateMon {
		firstDateTime := m.Date.AddDate(0, 0, -m.Date.Day()+1)
		date = time.Date(firstDateTime.Year(), firstDateTime.Month(), firstDateTime.Day(), 0, 0, 0, 0, firstDateTime.Location())
	}

	return &date
}

// SkipIntervalTime if skip >0 向后跳，skip<0 向前跳
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
		date = m.getWeekFirstDay().Add(m.TimeInterval * time.Duration(skip))
	} else if m.DateType == domain.DateMon {
		firstDateTime := m.Date.AddDate(0, 0, -m.Date.Day()+1)
		date = time.Date(firstDateTime.Year(), firstDateTime.Month(), firstDateTime.Day(), 0, 0, 0, 0, firstDateTime.Location()).AddDate(0, skip, 0)
	}

	return &date
}

// getWeekFirstDay 获取给定时间所在周的第一天
func (m *KLineTyp) getWeekFirstDay() time.Time {
	var date time.Time
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
	return date
}
