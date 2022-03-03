package process

import (
	"time"

	"git.cplus.link/crema/backend/pkg/domain"
)

//type KLineType interface {
//	Name() domain.DateType
//	GetDate(domain.DateType) *time.Time
//}

type KLineTyp struct {
	Date     *time.Time
	DateType domain.DateType
}

func (m *KLineTyp) Name() domain.DateType {
	return domain.DateMin
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
		date = time.Date(m.Date.Year(), m.Date.Month(), 0, 0, 0, 0, 0, m.Date.Location())
	}

	return &date
}
