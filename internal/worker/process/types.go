package process

import (
	"time"

	"git.cplus.link/crema/backend/pkg/domain"
)

var (
	dateMin = KLineTyp{
		BeforeIntervalDateType: domain.DateNone,
		DateType:               domain.DateMin,
	}

	dateTwelfth = KLineTyp{
		BeforeIntervalDateType: domain.DateMin,
		DateType:               domain.DateTwelfth,
		Interval:               5,
		TimeInterval:           time.Minute,
	}

	dateQuarter = KLineTyp{
		BeforeIntervalDateType: domain.DateTwelfth,
		DateType:               domain.DateQuarter,
		Interval:               3,
		TimeInterval:           time.Minute * 5,
	}

	dateHalfAnHour = KLineTyp{
		BeforeIntervalDateType: domain.DateQuarter,
		DateType:               domain.DateHalfAnHour,
		Interval:               2,
		TimeInterval:           time.Minute * 15,
	}

	dateHour = KLineTyp{
		BeforeIntervalDateType: domain.DateHalfAnHour,
		DateType:               domain.DateHour,
		Interval:               2,
		TimeInterval:           time.Minute * 30,
	}

	dateDay = KLineTyp{
		BeforeIntervalDateType: domain.DateHour,
		DateType:               domain.DateDay,
		Interval:               24,
		TimeInterval:           time.Hour,
	}

	dateWek = KLineTyp{
		BeforeIntervalDateType: domain.DateDay,
		DateType:               domain.DateWek,
		Interval:               7,
		TimeInterval:           time.Hour * 24,
	}

	dateMon = KLineTyp{
		BeforeIntervalDateType: domain.DateDay,
		DateType:               domain.DateMon,
		Interval:               31,
		TimeInterval:           time.Hour * 24,
	}
)

type KLineTyp struct {
	Date                   *time.Time
	DateType               domain.DateType
	BeforeIntervalDateType domain.DateType
	Interval               int // 相较于前一个时间段，用多少前一个时间段可以填满当前时间段
	TimeInterval           time.Duration
}

func (m *KLineTyp) Name() domain.DateType {
	return m.DateType
}

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
