package domain

import (
	"git.cplus.link/go/akit/util/decimal"
	"gorm.io/gorm"
)

type PositionCountSnapshot struct {
	gorm.Model
	UserAddress  string          `json:"user_address" gorm:"not null;type:varchar(64);uniqueIndex:positions_count_snapshot_user_address_swap_address_date_unique_key"` // 用户 address
	SwapAddress  string          `json:"swap_address" gorm:"not null;type:varchar(64);uniqueIndex:positions_count_snapshot_user_address_swap_address_date_unique_key"` // swap address
	PositionID   string          `json:"position_id"`                                                                                                                  // position id
	Date         string          `json:"date" gorm:"not null;type:timestamp(6);uniqueIndex:positions_count_snapshot_user_address_swap_address_date_unique_key"`        // 统计日期
	TokenAAmount decimal.Decimal `json:"token_a_amount" gorm:"type:decimal(36,18);default:0"`                                                                          // tokenA数量
	TokenBAmount decimal.Decimal `json:"token_b_amount" gorm:"type:decimal(36,18);default:0"`                                                                          // tokenB数量
	TokenAPrice  decimal.Decimal `json:"token_a_price" gorm:"type:decimal(36,18);default:0"`                                                                           // tokenA价格
	TokenBPrice  decimal.Decimal `json:"token_b_price" gorm:"type:decimal(36,18);default:0"`                                                                           // tokenB价格
	Raw          []byte          `json:"raw"`                                                                                                                          // 原始数据
}
