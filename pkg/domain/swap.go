package domain

import (
	"database/sql/driver"
	"time"
)

type SwapPairCount struct {
	ID                int64      `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt         *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt         *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	TokenAVolume      uint64     `json:"token_a_volume" gorm:""`
	TokenBVolume      uint64     `json:"token_b_volume" gorm:""`
	TokenABalance     uint64     `json:"token_a_balance" gorm:""`
	TokenBBalance     uint64     `json:"token_b_balance" gorm:""`
	TokenAPoolAddress string     `json:"-" gorm:"type:varchar(64);  index"`
	TokenBPoolAddress string     `json:"-" gorm:"type:varchar(64);  index"`
	TokenSwapAddress  string     `json:"-" gorm:"type:varchar(64);  index"`
	LastTransaction   string     `json:"-" gorm:"type:varchar(1024);"`
	Signature         string     `json:"-" gorm:"type:varchar(1024);"`
	PairName          string     `json:"-" gorm:"type:varchar(64);"`
	TokenASymbol      string     `json:"-" gorm:"type:varchar(32);"`
	TokenBSymbol      string     `json:"-" gorm:"type:varchar(32);"`
	TokenADecimal     int        `json:"-" gorm:"type:int2;"`
	TokenBDecimal     int        `json:"-" gorm:"type:int2;"`
}

type Tvl struct {
	ID            int64      `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt     *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt     *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	TotalTvlInUsd string     `json:"total_tvl_in_usd" gorm:"type:varchar(32);"`
	TotalVolInUsd string     `json:"total_vol_in_usd" gorm:"type:varchar(32);"`
	Pairs         JsonString `json:"pairs" gorm:"type:text;"`
}

type PairTvl struct {
	Name     string `json:"name"`
	TvlInUsd string `json:"tvl_in_usd"`
	VolInUsd string `json:"vol_in_usd"`
}

// JsonString 自定义json gorm byte类型
type JsonString string

func (j *JsonString) MarshalJSON() ([]byte, error) {
	return []byte(*j), nil
}

func (j *JsonString) UnmarshalJSON(data []byte) error {
	*j = JsonString(data)
	return nil
}

func (j *JsonString) Value() (driver.Value, error) {
	return driver.String.ConvertValue(*j)
}

func (j *JsonString) Scan(v interface{}) error {
	*j = JsonString(v.(string))
	return nil
}
