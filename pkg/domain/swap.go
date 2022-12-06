package domain

import (
	"time"

	"git.cplus.link/go/akit/util/decimal"
)

type SwapPairCountSharding struct {
	ID                int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt         *time.Time      `json:"-" gorm:"not null;index"`
	UpdatedAt         *time.Time      `json:"-" gorm:"not null;index"`
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
	TokenADecimal     int             `json:"-" gorm:"type:integer;"`
	TokenBDecimal     int             `json:"-" gorm:"type:integer;"`
	TxNum             uint64          `json:"tx_num" gorm:""`
}

func (*SwapPairCountSharding) TableName() string {
	return "swap_pairs_count_sharding"
}

// UserCountKLine 用户统计表
type UserCountKLine struct {
	ID                            int64           `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt                     *time.Time      `json:"-" gorm:"not null;index"`
	UpdatedAt                     *time.Time      `json:"-" gorm:"not null;index"`
	LastSwapTransactionID         int64           `json:"last_swap_transaction_id" gorm:"not null;default:0;index:idx_user_count_k_lines_last_swap_transaction_id_date_type,priority:3"`                                                          // 最后同步的transaction id
	UserAddress                   string          `json:"user_address" gorm:"not null;type:varchar(64);  uniqueIndex:user_swap_tvl_count_day_swap_address_unique_key"`                                                                            // 用户 address
	Date                          *time.Time      `json:"date" gorm:"not null;uniqueIndex:user_swap_tvl_count_day_swap_address_unique_key"`                                                                                                       // 统计日期
	DateType                      DateType        `json:"date_type" gorm:"not null;type:varchar(64);uniqueIndex:user_swap_tvl_count_day_swap_address_unique_key;index:idx_user_count_k_lines_last_swap_transaction_id_date_type,priority:1"`      // 时间类型（day,wek,mon）
	SwapAddress                   string          `json:"swap_address" gorm:"not null;type:varchar(64);  uniqueIndex:user_swap_tvl_count_day_swap_address_unique_key;index:idx_user_count_k_lines_last_swap_transaction_id_date_type,priority:2"` // swap地址
	TokenASymbol                  string          `json:"token_a_symbol" gorm:"not null;type:varchar(64);  index"`                                                                                                                                // swap token a symbol
	TokenBSymbol                  string          `json:"token_b_symbol" gorm:"not null;type:varchar(64);  index"`                                                                                                                                // swap token b symbol
	TokenAAddress                 string          `json:"token_a_address" gorm:"not null;type:varchar(64);  index"`                                                                                                                               // swap token a 地址
	TokenBAddress                 string          `json:"token_b_address" gorm:"not null;type:varchar(64);  index"`                                                                                                                               // swap token b 地址
	UserTokenAVolume              decimal.Decimal `json:"user_token_a_volume" gorm:"type:decimal(36,18);default:0"`                                                                                                                               // swap token a 总交易额
	UserTokenBVolume              decimal.Decimal `json:"user_token_b_volume" gorm:"type:decimal(36,18);default:0"`                                                                                                                               // swap token b 总交易额
	TokenAQuoteVolume             decimal.Decimal `json:"token_a_quote_volume" gorm:"type:decimal(36,18);default:0"`                                                                                                                              // swap token a 获得量
	TokenBQuoteVolume             decimal.Decimal `json:"token_b_quote_volume" gorm:"type:decimal(36,18);default:0"`                                                                                                                              // swap token b 获得量
	TxNum                         int64           `json:"tx_num"`                                                                                                                                                                                 // 交易笔数
	TokenAWithdrawLiquidityVolume decimal.Decimal `json:"token_a_withdraw_liquidity_volume"  gorm:"type:decimal(36,18);default:0"`                                                                                                                // 移出流动性数量
	TokenADepositLiquidityVolume  decimal.Decimal `json:"token_a_deposit_liquidity_volume"  gorm:"type:decimal(36,18);default:0"`                                                                                                                 // 添加流动性数量
	TokenBWithdrawLiquidityVolume decimal.Decimal `json:"token_b_withdraw_liquidity_volume"  gorm:"type:decimal(36,18);default:0"`                                                                                                                // 移出流动性数量
	TokenBDepositLiquidityVolume  decimal.Decimal `json:"token_b_deposit_liquidity_volume"  gorm:"type:decimal(36,18);default:0"`                                                                                                                 // 添加流动性数量
	TokenAClaimVolume             decimal.Decimal `json:"token_a_claim_volume"  gorm:"type:decimal(36,18);default:0"`                                                                                                                             // Claim数量
	TokenBClaimVolume             decimal.Decimal `json:"token_b_claim_volume"  gorm:"type:decimal(36,18);default:0"`                                                                                                                             // Claim数量
}

type SwapUserCount struct {
	ID          int64      `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt   *time.Time `json:"-" gorm:"not null;index"`
	UpdatedAt   *time.Time `json:"-" gorm:"not null;index"`
	SwapAddress string     `json:"swap_address" gorm:"not null;type:text;  uniqueIndex"` // swap地址 ，因为每个swap_address 同步进度不一致，通过swap地址来管理进度，如果是v1,那么SyncUtilID表示的是v1版本的进度
	SyncUtilID  int64      `json:"sync_util_id"`                                         // 解析数据位置
}

type TransActionUserCount struct {
	ID          int64      `json:"-" gorm:"primaryKey;auto_increment"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt   *time.Time `json:"-" gorm:"not null;index"`
	UpdatedAt   *time.Time `json:"-" gorm:"not null;index"`
	UserAddress string     `json:"user_address" gorm:"not null;type:text;uniqueIndex"` // 用户 address
}

func (*TransActionUserCount) TableName() string {
	return "transaction_user_counts"
}

type SwapPairPriceKLine struct {
	ID          int64           `json:"-" gorm:"primaryKey;auto_increment;Index:SwapPairPriceKLine_ID_swap_address_date_date_type_index"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt   *time.Time      `json:"-" gorm:"not null;index"`
	UpdatedAt   *time.Time      `json:"-" gorm:"not null;index"`
	SwapAddress string          `json:"swap_address" gorm:"not null;type:varchar(64);uniqueIndex:swap_pair_price_k_line_swap_address_date_date_type_unique_key;Index:SwapPairPriceKLine_ID_swap_address_date_date_type_index"` // swap地址
	Open        decimal.Decimal `json:"open" gorm:"type:decimal(36,18);default:0"`                                                                                                                                             // 统计时间段第一个值
	High        decimal.Decimal `json:"high" gorm:"type:decimal(36,18);default:0"`                                                                                                                                             // 最大值
	Low         decimal.Decimal `json:"low"  gorm:"type:decimal(36,18);default:0"`                                                                                                                                             // 最小值
	Settle      decimal.Decimal `json:"settle" gorm:"type:decimal(36,18);default:0"`                                                                                                                                           // 结束值
	Avg         decimal.Decimal `json:"avg" gorm:"type:decimal(36,18);default:0"`                                                                                                                                              // 平均值
	Num         int64           `json:"num"`                                                                                                                                                                                   // 获取次数
	DateType    DateType        `json:"date_type" gorm:"not null;type:varchar(64);uniqueIndex:swap_pair_price_k_line_swap_address_date_date_type_unique_key;Index:SwapPairPriceKLine_ID_swap_address_date_date_type_index"`    // 时间类型（min,quarter,hour,day,wek,mon）
	Date        *time.Time      `json:"date" gorm:"not null;uniqueIndex:swap_pair_price_k_line_swap_address_date_date_type_unique_key; index;Index:SwapPairPriceKLine_ID_swap_address_date_date_type_index"`                   // 统计日期
}

func (*SwapPairPriceKLine) TableName() string {
	return "swap_pair_price_k_lines"
}

type SwapTokenPriceKLine struct {
	ID        int64           `json:"-" gorm:"primaryKey;auto_increment; Index:SwapTokenPriceKLine_symbol_date_date_type_id_index"` // 自增主键，自增主键不能有任何业务含义。
	CreatedAt *time.Time      `json:"-" gorm:"not null;index"`
	UpdatedAt *time.Time      `json:"-" gorm:"not null;index"`
	Symbol    string          `json:"symbol" gorm:"not null;type:varchar(64);uniqueIndex:idx_swap_token_price_k_line_symbol_date_date_type_unique_key; Index:SwapTokenPriceKLine_symbol_date_date_type_id_index"`    // symbol
	Open      decimal.Decimal `json:"open" gorm:"type:decimal(36,18);default:0"`                                                                                                                                     // 统计时间段第一个值
	High      decimal.Decimal `json:"high" gorm:"type:decimal(36,18);default:0"`                                                                                                                                     // 最大值
	Low       decimal.Decimal `json:"low"  gorm:"type:decimal(36,18);default:0"`                                                                                                                                     // 最小值
	Settle    decimal.Decimal `json:"settle" gorm:"type:decimal(36,18);default:0"`                                                                                                                                   // 结束值
	Avg       decimal.Decimal `json:"avg" gorm:"type:decimal(36,18);default:0"`                                                                                                                                      // 平均值
	Num       int64           `json:"num"`                                                                                                                                                                           // 获取次数
	DateType  DateType        `json:"date_type" gorm:"not null;type:varchar(64);uniqueIndex:idx_swap_token_price_k_line_symbol_date_date_type_unique_key; Index:SwapTokenPriceKLine_symbol_date_date_type_id_index"` // 时间类型（min,quarter,hour,day,wek,mon）
	Date      *time.Time      `json:"date" gorm:"not null;uniqueIndex:idx_swap_token_price_k_line_symbol_date_date_type_unique_key; Index:SwapTokenPriceKLine_symbol_date_date_type_id_index"`                       // 统计日期
}

func (*SwapTokenPriceKLine) TableName() string {
	return "swap_token_price_k_lines"
}
