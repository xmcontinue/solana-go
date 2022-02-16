package domain

import (
	"time"

	"gorm.io/gorm"
)

type TransactionBase struct {
	gorm.Model
	BlockTime        *time.Time `gorm:"not null;type:timestamp(6)"`
	Slot             uint64     `gorm:"index"`
	TransactionData  string     `gorm:"text(0)"`
	MateData         string     `gorm:"text(0)"`
	TokenSwapAddress string     `gorm:"varchar(64);index"`
	Signature        string     `gorm:"varchar(128);index"`
}
