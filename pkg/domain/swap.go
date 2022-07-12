package domain

import (
	"time"

	"git.cplus.link/go/akit/util/decimal"
)

type SwapPairCount struct {
	ID                int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt         *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt         *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	TokenAVolume      decimal.Decimal `json:"token_a_volume" gorm:"type:decimal(36,18);"`
	TokenBVolume      decimal.Decimal `json:"token_b_volume" gorm:"type:decimal(36,18);"`
	TokenABalance     decimal.Decimal `json:"token_a_balance" gorm:"type:decimal(36,18);"`
	TokenBBalance     decimal.Decimal `json:"token_b_balance" gorm:"type:decimal(36,18);"`
	TokenAPoolAddress string          `json:"-" gorm:"type:varchar(64);  index"`
	TokenBPoolAddress string          `json:"-" gorm:"type:varchar(64);  index"`
	TokenSwapAddress  string          `json:"-" gorm:"type:varchar(64);  index"`
	LastTransaction   string          `json:"-" gorm:"type:varchar(1024);"`
	Signature         string          `json:"-" gorm:"type:varchar(1024);"`
	PairName          string          `json:"-" gorm:"type:varchar(64);"`
	TokenASymbol      string          `json:"-" gorm:"type:varchar(32);"`
	TokenBSymbol      string          `json:"-" gorm:"type:varchar(32);"`
	TokenADecimal     int             `json:"-" gorm:"type:int2;"`
	TokenBDecimal     int             `json:"-" gorm:"type:int2;"`
	TxNum             uint64          `json:"tx_num" gorm:""`
}

func (*SwapPairCount) TableName() string {
	return "swap_pairs_counts"
}

// UserCountKLine 用户统计表
type UserCountKLine struct {
	ID                            int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt                     *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt                     *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	LastSwapTransactionID         int64           `json:"last_swap_transaction_id" gorm:"not null;default:0;index"`                                                    // 最后同步的transaction id
	UserAddress                   string          `json:"user_address" gorm:"not null;type:varchar(64);  uniqueIndex:user_swap_tvl_count_day_swap_address_unique_key"` // 用户 address
	Date                          *time.Time      `json:"date" gorm:"not null;type:timestamp(6);uniqueIndex:user_swap_tvl_count_day_swap_address_unique_key"`          // 统计日期
	DateType                      DateType        `json:"date_type" gorm:"not null;type:varchar(64);uniqueIndex:user_swap_tvl_count_day_swap_address_unique_key"`      // 时间类型（day,wek,mon） // TODO 新加字段
	SwapAddress                   string          `json:"swap_address" gorm:"not null;type:varchar(64);  uniqueIndex:user_swap_tvl_count_day_swap_address_unique_key"` // swap地址
	TokenASymbol                  string          `json:"token_a_symbol" gorm:"not null;type:varchar(64);  index"`                                                     // swap token a symbol // TODO 新加字段
	TokenBSymbol                  string          `json:"token_b_symbol" gorm:"not null;type:varchar(64);  index"`                                                     // swap token b symbol // TODO 新加字段
	TokenAAddress                 string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"`                                                    // swap token a 地址
	TokenBAddress                 string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"`                                                    // swap token b 地址
	UserTokenAVolume              decimal.Decimal `json:"user_token_a_volume" gorm:"type:decimal(36,18);default:0"`                                                    // swap token a 总交易额
	UserTokenBVolume              decimal.Decimal `json:"user_token_b_volume" gorm:"type:decimal(36,18);default:0"`                                                    // swap token b 总交易额
	TokenAQuoteVolume             decimal.Decimal `json:"token_a_quote_volume" gorm:"type:decimal(36,18);default:0"`                                                   // swap token a 获得量
	TokenBQuoteVolume             decimal.Decimal `json:"token_b_quote_volume" gorm:"type:decimal(36,18);default:0"`                                                   // swap token b 获得量
	UserTokenABalance             decimal.Decimal `json:"user_token_a_balance" gorm:"type:decimal(36,18);default:0"`                                                   // swap token a 余额
	UserTokenBBalance             decimal.Decimal `json:"user_token_b_balance" gorm:"type:decimal(36,18);default:0"`                                                   // swap token b 余额
	TxNum                         int64           `json:"tx_num"`                                                                                                      // 交易笔数
	TokenAWithdrawLiquidityVolume decimal.Decimal `json:"token_a_withdraw_liquidity_volume"  gorm:"type:decimal(36,18);default:0"`                                     // 移出流动性数量 // TODO 新加字段
	TokenADepositLiquidityVolume  decimal.Decimal `json:"token_a_deposit_liquidity_volume"  gorm:"type:decimal(36,18);default:0"`                                      // 添加流动性数量 // TODO 新加字段
	TokenBWithdrawLiquidityVolume decimal.Decimal `json:"token_b_withdraw_liquidity_volume"  gorm:"type:decimal(36,18);default:0"`                                     // 移出流动性数量 // TODO 新加字段
	TokenBDepositLiquidityVolume  decimal.Decimal `json:"token_b_deposit_liquidity_volume"  gorm:"type:decimal(36,18);default:0"`                                      // 添加流动性数量 // TODO 新加字段
	TokenAClaimVolume             decimal.Decimal `json:"token_a_claim_volume"  gorm:"type:decimal(36,18);default:0"`                                                  // Claim数量 // TODO 新加字段
	TokenBClaimVolume             decimal.Decimal `json:"token_b_claim_volume"  gorm:"type:decimal(36,18);default:0"`                                                  // Claim数量 // TODO 新加字段
}

type SwapPairPriceKLine struct {
	ID          int64           `json:"-" gorm:"primaryKey;auto_increment;Index:SwapPairPriceKLine_ID_swap_address_date_date_type_index"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt   *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt   *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	SwapAddress string          `json:"swap_address" gorm:"not null;type:varchar(64);uniqueIndex:swap_pair_price_k_line_swap_address_date_date_type_unique_key;Index:SwapPairPriceKLine_ID_swap_address_date_date_type_index"` // swap地址
	Open        decimal.Decimal `json:"open" gorm:"type:decimal(36,18);default:0"`                                                                                                                                             // 统计时间段第一个值
	High        decimal.Decimal `json:"high" gorm:"type:decimal(36,18);default:0"`                                                                                                                                             // 最大值
	Low         decimal.Decimal `json:"low"  gorm:"type:decimal(36,18);default:0"`                                                                                                                                             // 最小值
	Settle      decimal.Decimal `json:"settle" gorm:"type:decimal(36,18);default:0"`                                                                                                                                           // 结束值
	Avg         decimal.Decimal `json:"avg" gorm:"type:decimal(36,18);default:0"`                                                                                                                                              // 平均值
	Num         int64           `json:"num"`                                                                                                                                                                                   // 获取次数
	Date        *time.Time      `json:"date" gorm:"not null;type:timestamp(6);uniqueIndex:swap_pair_price_k_line_swap_address_date_date_type_unique_key; index;Index:SwapPairPriceKLine_ID_swap_address_date_date_type_index"` // 统计日期
	DateType    DateType        `json:"date_type" gorm:"not null;type:varchar(64);uniqueIndex:swap_pair_price_k_line_swap_address_date_date_type_unique_key;Index:SwapPairPriceKLine_ID_swap_address_date_date_type_index"`    // 时间类型（min,quarter,hour,day,wek,mon）
}

type SwapTokenPriceKLine struct {
	ID        int64           `json:"-" gorm:"primaryKey;auto_increment; Index: SwapTokenPriceKLine_symbol_date_date_type_id_index"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	UpdatedAt *time.Time      `json:"-" gorm:"not null;type:timestamp(6);index"`
	Symbol    string          `json:"symbol" gorm:"not null;type:varchar(64);uniqueIndex:idx_swap_token_price_k_line_symbol_date_date_type_unique_key"`    // symbol
	Open      decimal.Decimal `json:"open" gorm:"type:decimal(36,18);default:0"`                                                                           // 统计时间段第一个值
	High      decimal.Decimal `json:"high" gorm:"type:decimal(36,18);default:0"`                                                                           // 最大值
	Low       decimal.Decimal `json:"low"  gorm:"type:decimal(36,18);default:0"`                                                                           // 最小值
	Settle    decimal.Decimal `json:"settle" gorm:"type:decimal(36,18);default:0"`                                                                         // 结束值
	Avg       decimal.Decimal `json:"avg" gorm:"type:decimal(36,18);default:0"`                                                                            // 平均值
	Num       int64           `json:"num"`                                                                                                                 // 获取次数
	DateType  DateType        `json:"date_type" gorm:"not null;type:varchar(64);uniqueIndex:idx_swap_token_price_k_line_symbol_date_date_type_unique_key"` // 时间类型（min,quarter,hour,day,wek,mon）
	Date      *time.Time      `json:"date" gorm:"not null;type:timestamp(6);uniqueIndex:idx_swap_token_price_k_line_symbol_date_date_type_unique_key"`     // 统计日期
}
