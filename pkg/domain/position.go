package domain

import (
	"git.cplus.link/go/akit/util/decimal"
	"gorm.io/gorm"
)

type PositionCountSnapshot struct {
	gorm.Model
	UserAddress  string          `json:"user_address" gorm:"not null;type:varchar(64);uniqueIndex:positions_count_snapshot_user_address_swap_address_position_id_date_unique_key"` // 用户 address
	SwapAddress  string          `json:"swap_address" gorm:"not null;type:varchar(64);uniqueIndex:positions_count_snapshot_user_address_swap_address_position_id_date_unique_key"` // swap address
	PositionID   string          `json:"position_id" gorm:"not null;type:varchar(64);uniqueIndex:positions_count_snapshot_user_address_swap_address_position_id_date_unique_key"`  // position id
	Date         string          `json:"date" gorm:"not null;type:timestamp(6);uniqueIndex:positions_count_snapshot_user_address_swap_address_position_id_date_unique_key"`        // 统计日期
	TokenAAmount decimal.Decimal `json:"token_a_amount" gorm:"type:decimal(36,18);default:0"`                                                                                      // tokenA数量
	TokenBAmount decimal.Decimal `json:"token_b_amount" gorm:"type:decimal(36,18);default:0"`                                                                                      // tokenB数量
	TokenAPrice  decimal.Decimal `json:"token_a_price" gorm:"type:decimal(36,18);default:0"`                                                                                       // tokenA价格
	TokenBPrice  decimal.Decimal `json:"token_b_price" gorm:"type:decimal(36,18);default:0"`                                                                                       // tokenB价格
	Raw          []byte          `json:"raw"`                                                                                                                                      // 原始数据
}

type PositionV2Snapshot struct {
	gorm.Model
	ClmmPool         string          `json:"clmm_pool,omitempty" gorm:"not null;type:varchar(64);uniqueIndex:positions_V2_snapshot_user_address_swap_address_position_id_date_unique_key"`
	PositionNFTMint  string          `json:"position_nft_mint,omitempty"  gorm:"not null;type:varchar(64);uniqueIndex:positions_V2_snapshot_user_address_swap_address_position_id_date_unique_key"`
	Date             string          `json:"date" gorm:"not null;type:timestamp(6);uniqueIndex:positions_V2_snapshot_user_address_swap_address_position_id_date_unique_key"` // 统计日期
	Liquidity        decimal.Decimal `json:"liquidity" gorm:"type:decimal(36,18);default:0"`
	TickLowerIndex   int32           `json:"tick_lower_index,omitempty"`
	TickUpperIndex   int32           `json:"tick_upper_index,omitempty"`
	FeeGrowthInsideA decimal.Decimal `json:"fee_growth_inside_a" gorm:"type:decimal(36,18);default:0"`
	FeeOwedA         uint64          `json:"fee_owed_a,omitempty"`
	FeeGrowthInsideB decimal.Decimal `json:"fee_growth_inside_b" gorm:"type:decimal(36,18);default:0"`
	FeeOwedB         uint64          `json:"fee_owed_b,omitempty"`
	//GrowthInside1    decimal.Decimal `json:"growth_inside_1" gorm:"type:decimal(36,18);default:0"`
	//AmountOwed1      uint64          `json:"amount_owed_1,omitempty"`
	//GrowthInside2    decimal.Decimal `json:"growth_inside_2" gorm:"type:decimal(36,18);default:0"`
	//AmountOwed2      uint64          `json:"amount_owed_2,omitempty"`
	//GrowthInside3    decimal.Decimal `json:"growth_inside_3" gorm:"type:decimal(36,18);default:0"`
	//AmountOwed3      uint64          `json:"amount_owed_3,omitempty"`
}
