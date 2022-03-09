package domain

import (
	"time"

	"git.cplus.link/go/akit/util/decimal"
)

// LiquidityDistributionCount 流动性区间分布统计
type LiquidityDistributionCount struct {
	ID                    int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	LastSwapTransactionID int64           `json:"last_swap_transaction_id" gorm:"not null;default:0"`       // 最后同步的transaction id
	SwapAddress           string          `json:"swap_address" gorm:"not null;type:varchar(64);     index"` // swap地址
	TokenAAddress         string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"` // swap token a 地址
	TokenBAddress         string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"` // swap token b 地址
	StartTick             int64           `json:"start_tick" gorm:"not nul"`                                // 开始区间tick
	EndTick               int64           `json:"end_tick" gorm:"not nul"`                                  // 结束区间tick
	TokenADeposit         decimal.Decimal `json:"token_a_deposit" gorm:"type:decimal(36,18);default:0"`     // token A 质押量
	TokenBDeposit         decimal.Decimal `json:"token_b_deposit" gorm:"type:decimal(36,18);default:0"`     // token B 质押量
	Tvl                   decimal.Decimal `json:"tvl" gorm:"type:decimal(36,18);default:0"`                 // 总锁仓量
}

// LiquidityChangeCount 流动性变动统计
type LiquidityChangeCount struct {
	ID                    int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt             *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	LastSwapTransactionID int64           `json:"last_swap_transaction_id" gorm:"not null;default:0"`       // 最后同步的transaction id
	SwapAddress           string          `json:"swap_address" gorm:"not null;type:varchar(64);     index"` // swap地址
	TokenAAddress         string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"` // swap token a 地址
	TokenBAddress         string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"` // swap token b 地址
	TokenADeposit         decimal.Decimal `json:"token_a_deposit" gorm:"type:decimal(36,18);default:0"`     // token A 质押量
	TokenBDeposit         decimal.Decimal `json:"token_b_deposit" gorm:"type:decimal(36,18);default:0"`     // token B 质押量
	TokenAWithdraw        decimal.Decimal `json:"token_a_withdraw" gorm:"type:decimal(36,18);default:0"`    // token A 取出量
	TokenBWithdraw        decimal.Decimal `json:"token_b_withdraw" gorm:"type:decimal(36,18);default:0"`    // token B 取出量
	Tvl                   decimal.Decimal `json:"tvl" gorm:"type:decimal(36,18);default:0"`                 // 总锁仓量
	Date                  *time.Time      `json:"date" gorm:"not null;type:timestamp(6)"`                   // 统计日期
	DateType              string          `json:"date_type" gorm:"not null;type:varchar(16)"`               // 统计日期
}
