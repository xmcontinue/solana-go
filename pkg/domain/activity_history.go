package domain

import (
	"time"
)

type ActivityHistory struct {
	ID           int64      `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt    *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt    *time.Time `json:"-" gorm:"not null;type:timestamp(6);index"`
	UserKey      string     `json:"user_key" gorm:"type:varchar(64); not null; index:idx_user"`
	MintKey      string     `json:"mint_key"  gorm:"type:varchar(64); not null; index:idx_mint"`
	Crm          uint64      `json:"crm" gorm:"type:numeric(64); not null;default:0"`      //
	Marinade     uint64      `json:"marinade" gorm:"type:numeric(64); not null;default:0"` //
	Port         uint64      `json:"port" gorm:"type:numeric(64); not null;default:0"`     //
	Hubble       uint64      `json:"hubble" gorm:"type:numeric(64); not null;default:0"`   //
	Nirv         uint64      `json:"nirv" gorm:"type:numeric(64); not null;default:0"`     //
	SignatureCrm string     `json:"signature_crm" gorm:"type:varchar(100); not null"`
	Signature    string     `json:"signature" gorm:"type:varchar(100); not null"`
	BlockTime    int64     `json:"block_time" gorm:"type:numeric(64); not null; index"`
	Degree       uint8      `json:"degree"  gorm:"type:int2; not null"`
	Caffeine     uint64      `json:"caffeine" gorm:"type:numeric(64); not null"` //
}
