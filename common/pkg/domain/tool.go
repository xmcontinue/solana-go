package domain

import (
	"time"
)

// TokenVolumeCount  市场货币价格表
type TokenVolumeCount struct {
	ID                int64      `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt         *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt         *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	TokenAVolume      uint64     `gorm:""`
	TokenBVolume      uint64     `gorm:""`
	TokenABalance     uint64     `gorm:""`
	TokenBBalance     uint64     `gorm:""`
	TokenAPoolAddress string     `gorm:"type:varchar(64);  index"`
	TokenBPoolAddress string     `gorm:"type:varchar(64);  index"`
	TokenSwapAddress  string     `gorm:"type:varchar(64);  index"`
	LastTransaction   string     `json:"-" gorm:"type:varchar(1024);"`
	Signature         string     `json:"-" gorm:"type:varchar(1024);"`
}
